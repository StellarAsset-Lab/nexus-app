package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

// dbtx is satisfied by both *sql.DB and *sql.Tx, so callers can run a query
// either standalone or as part of a larger transaction (e.g. the indexer
// advancing a checkpoint atomically with the event batch that produced it).
type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// GetCheckpoint returns the durable indexer checkpoint for id, or nil if
// none has been recorded yet (not an error — this is the expected state
// before the indexer's first successful batch).
func GetCheckpoint(ctx context.Context, q dbtx, id string) (*model.Checkpoint, error) {
	row := q.QueryRowContext(ctx,
		`SELECT id, last_ledger, cursor, updated_at FROM indexer_checkpoints WHERE id = $1`, id)

	var checkpoint model.Checkpoint
	if err := row.Scan(&checkpoint.ID, &checkpoint.LastLedger, &checkpoint.Cursor, &checkpoint.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan checkpoint: %w", err)
	}
	return &checkpoint, nil
}

// UpsertCheckpoint records the indexer's resume point. Call this only after
// the corresponding event batch has been durably persisted — never before,
// per spec §15/§16 ("advance checkpoint only after successful persistence").
func UpsertCheckpoint(ctx context.Context, q dbtx, id string, lastLedger int64, cursor *string) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO indexer_checkpoints (id, last_ledger, cursor, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (id) DO UPDATE SET
			last_ledger = EXCLUDED.last_ledger,
			cursor = EXCLUDED.cursor,
			updated_at = now()`,
		id, lastLedger, cursor,
	)
	if err != nil {
		return fmt.Errorf("upsert checkpoint: %w", err)
	}
	return nil
}
