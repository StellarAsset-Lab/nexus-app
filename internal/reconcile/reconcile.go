// Package reconcile keeps the `transactions` observation table in sync with
// real on-chain outcomes. The indexer only projects Registry/Order contract
// events; it never writes to `transactions`. Any transaction hash that
// produced at least one event is a candidate the worker checks against RPC
// directly, so the Transaction Center can show a real, current status for
// every transaction it has ever seen — never a fabricated or defaulted one.
package reconcile

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

// BatchLimit bounds how many pending hashes a single Run processes, so one
// pass never blocks indefinitely behind an unbounded backlog.
const BatchLimit = 100

type Reconciler struct {
	rpc    *rpcclient.Client
	db     *sql.DB
	logger *slog.Logger
}

func NewReconciler(rpc *rpcclient.Client, db *sql.DB, logger *slog.Logger) *Reconciler {
	return &Reconciler{rpc: rpc, db: db, logger: logger}
}

// Run performs one reconciliation pass: for every transaction hash observed
// in stellar_events that has no terminal transactions row yet, it asks RPC
// for the transaction's real status and upserts the result. It returns the
// number of hashes it processed.
func (r *Reconciler) Run(ctx context.Context) (int, error) {
	hashes, err := nexusdb.PendingReconciliationHashes(ctx, r.db, BatchLimit)
	if err != nil {
		return 0, fmt.Errorf("load pending reconciliation hashes: %w", err)
	}

	for _, hash := range hashes {
		if err := r.reconcileOne(ctx, hash); err != nil {
			r.logger.Error("reconcile transaction failed", "error", err, "transactionHash", hash)
			continue
		}
	}
	return len(hashes), nil
}

// RunLoop calls Run repeatedly until ctx is cancelled, waiting interval
// between passes. A pass error is logged and retried on the next tick rather
// than aborting the worker — a single failed GetTransaction call (e.g. a
// transient RPC hiccup) must not take down reconciliation entirely.
func (r *Reconciler) RunLoop(ctx context.Context, interval time.Duration) {
	for {
		if ctx.Err() != nil {
			return
		}

		n, err := r.Run(ctx)
		if err != nil {
			r.logger.Error("reconciliation pass failed", "error", err)
		} else if n > 0 {
			r.logger.Info("reconciliation pass complete", "processed", n)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}

func (r *Reconciler) reconcileOne(ctx context.Context, hash string) error {
	resp, err := r.rpc.GetTransaction(ctx, protocol.GetTransactionRequest{Hash: hash})
	if err != nil {
		return fmt.Errorf("getTransaction: %w", err)
	}

	status, ledger := mapStatus(resp)

	contractID, orderID, err := nexusdb.TransactionContext(ctx, r.db, hash)
	if err != nil {
		return fmt.Errorf("load transaction context: %w", err)
	}

	if err := nexusdb.UpsertTransaction(ctx, r.db, model.Transaction{
		TransactionHash: hash,
		Ledger:          ledger,
		Status:          status,
		SourceContract:  contractID,
		OrderID:         orderID,
	}); err != nil {
		return fmt.Errorf("upsert transaction: %w", err)
	}
	return nil
}

// mapStatus translates RPC's getTransaction status into this application's
// own terminal-observation vocabulary. NOT_FOUND is reported as "Pending" —
// the transaction was submitted (it produced events we already indexed) but
// RPC has not yet ingested it into its transaction store; this is expected
// immediately after submission, not an error.
func mapStatus(resp protocol.GetTransactionResponse) (status string, ledger *int64) {
	switch resp.Status {
	case protocol.TransactionStatusSuccess:
		l := int64(resp.Ledger)
		return "Confirmed", &l
	case protocol.TransactionStatusFailed:
		l := int64(resp.Ledger)
		return "Failed", &l
	default:
		return "Pending", nil
	}
}
