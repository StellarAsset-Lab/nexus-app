package indexing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

// PollerConfig configures one indexer polling run. RegistryContractID and
// OrderContractID are both optional — a deployment may only have one
// contract configured yet, but at least one must be set, per spec §16.2
// ("do not query every contract on the network").
type PollerConfig struct {
	NetworkName        string
	RegistryContractID string
	OrderContractID    string
	BatchLimit         uint
	// ConfirmationLookback re-checks this many trailing ledgers on every
	// poll, in case RPC briefly reordered or backfilled recent events.
	ConfirmationLookback  uint32
	ConfiguredStartLedger *uint32
	PollInterval          time.Duration
}

// Poller runs the bounded, checkpointed event-polling loop described in spec
// §16.1/§16.2. This commit persists raw event rows only — contract-specific
// decoding and projection land in the Registry/Order event indexing commits;
// undecoded rows are not the same as decode failures, so decoded_payload and
// decode_error both stay NULL here rather than being marked as failed.
type Poller struct {
	rpc    *rpcclient.Client
	db     *sql.DB
	logger *slog.Logger
	cfg    PollerConfig
}

func NewPoller(rpc *rpcclient.Client, db *sql.DB, logger *slog.Logger, cfg PollerConfig) (*Poller, error) {
	if cfg.RegistryContractID == "" && cfg.OrderContractID == "" {
		return nil, errors.New("at least one of RegistryContractID or OrderContractID must be configured")
	}
	if cfg.BatchLimit == 0 {
		cfg.BatchLimit = 100
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 5 * time.Second
	}
	return &Poller{rpc: rpc, db: db, logger: logger, cfg: cfg}, nil
}

func (p *Poller) contractIDs() []string {
	ids := make([]string, 0, 2)
	if p.cfg.RegistryContractID != "" {
		ids = append(ids, p.cfg.RegistryContractID)
	}
	if p.cfg.OrderContractID != "" {
		ids = append(ids, p.cfg.OrderContractID)
	}
	return ids
}

// Run polls for events until ctx is cancelled, retrying transient RPC
// failures with exponential backoff and falling back to periodic polling
// once it has caught up to the chain head.
func (p *Poller) Run(ctx context.Context) error {
	backoff := newBackoff()

	for {
		if err := ctx.Err(); err != nil {
			return nil // graceful shutdown, not a failure
		}

		caughtUp, err := p.pollOnce(ctx)
		if err != nil {
			delay := backoff.next()
			p.logger.Warn("poll iteration failed, retrying with backoff", "error", err, "delay", delay)
			if !sleepOrDone(ctx, delay) {
				return nil
			}
			continue
		}
		backoff.reset()

		if caughtUp {
			if !sleepOrDone(ctx, p.cfg.PollInterval) {
				return nil
			}
		}
	}
}

// pollOnce fetches and persists one bounded window of events. It reports
// whether the indexer is now caught up to the chain head (i.e. the caller
// should wait before polling again rather than immediately looping).
func (p *Poller) pollOnce(ctx context.Context) (caughtUp bool, err error) {
	health, err := p.rpc.GetHealth(ctx)
	if err != nil {
		return false, fmt.Errorf("get rpc health: %w", err)
	}

	start, err := ResolveStartLedger(ctx, p.db, p.cfg.NetworkName, p.cfg.ConfiguredStartLedger, health.OldestLedger, health.LatestLedger)
	if err != nil {
		return false, err
	}

	if start >= p.cfg.ConfirmationLookback {
		start -= p.cfg.ConfirmationLookback
	} else {
		start = health.OldestLedger
	}
	if start < health.OldestLedger {
		start = health.OldestLedger
	}

	if start > health.LatestLedger {
		return true, nil
	}

	filter := protocol.EventFilter{
		EventType:   protocol.EventTypeSet{protocol.EventTypeContract: struct{}{}},
		ContractIDs: p.contractIDs(),
	}

	resp, err := p.rpc.GetEvents(ctx, protocol.GetEventsRequest{
		StartLedger: start,
		Filters:     []protocol.EventFilter{filter},
		Pagination:  &protocol.PaginationOptions{Limit: p.cfg.BatchLimit},
		Format:      protocol.FormatBase64,
	})
	if err != nil {
		return false, fmt.Errorf("get events: %w", err)
	}

	if err := persistEventBatch(ctx, p.db, p.cfg.NetworkName, resp, p.cfg.BatchLimit); err != nil {
		return false, fmt.Errorf("persist event batch: %w", err)
	}

	p.logger.Info("indexer poll",
		"network", p.cfg.NetworkName,
		"start_ledger", start,
		"latest_ledger", resp.LatestLedger,
		"events", len(resp.Events),
	)

	return uint(len(resp.Events)) < p.cfg.BatchLimit, nil
}

// persistEventBatch writes the raw event rows and advances the checkpoint
// in a single transaction — the checkpoint only moves forward once the
// batch it describes is durably persisted, per spec §15/§16.1. Duplicate
// events (event_id already seen) are silently no-ops, making re-processing
// of an overlapping confirmation-lookback window idempotent.
func persistEventBatch(ctx context.Context, conn *sql.DB, networkName string, resp protocol.GetEventsResponse, batchLimit uint) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, event := range resp.Events {
		var valueXDR string
		if event.ValueXDR != "" {
			valueXDR = event.ValueXDR
		}

		var topicsXDR string
		if len(event.TopicXDR) > 0 {
			// Topics are stored as their base64 XDR values joined; the
			// decoder (added alongside Registry/Order event indexing)
			// re-splits and decodes each one.
			for i, t := range event.TopicXDR {
				if i > 0 {
					topicsXDR += ","
				}
				topicsXDR += t
			}
		}

		ledgerClosedAt, err := time.Parse(time.RFC3339, event.LedgerClosedAt)
		if err != nil {
			ledgerClosedAt = time.Unix(0, 0).UTC()
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO stellar_events (
				event_id, ledger, ledger_closed_at, transaction_hash, transaction_index,
				operation_index, event_type, contract_id, topics_xdr, value_xdr
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (event_id) DO NOTHING`,
			event.ID, event.Ledger, ledgerClosedAt, event.TransactionHash, event.TxIndex,
			event.OpIndex, event.EventType, event.ContractID, topicsXDR, valueXDR,
		)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", event.ID, err)
		}
	}

	// Advance the checkpoint even when no matching events were found in this
	// window: a batch smaller than the requested limit means RPC has no
	// more matching events up to its reported latest ledger, so it is safe
	// to advance all the way there. A full batch means more events may
	// exist at or after the last event's ledger, so advance only through
	// it — never skipping a ledger that hasn't actually been scanned.
	var newCheckpointLedger int64
	if uint(len(resp.Events)) < batchLimit {
		newCheckpointLedger = int64(resp.LatestLedger)
	} else if len(resp.Events) > 0 {
		newCheckpointLedger = int64(resp.Events[len(resp.Events)-1].Ledger)
	}

	if newCheckpointLedger > 0 {
		if err := nexusdb.UpsertCheckpoint(ctx, tx, CheckpointID(networkName), newCheckpointLedger, nil); err != nil {
			return err
		}
	}

	return tx.Commit()
}
