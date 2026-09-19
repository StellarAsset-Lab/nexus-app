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
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
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

// AssetFilter narrows ListAssets. A nil Active means "no filter" (both
// active and inactive assets), never defaulted to a specific value.
type AssetFilter struct {
	Active *bool
	// Cursor is the last asset address seen on the previous page (keyset
	// pagination — deterministic ordering by primary key, no OFFSET drift
	// under concurrent writes, per spec §30).
	Cursor *string
	Limit  int
}

// ListAssets returns up to filter.Limit+1 assets ordered by asset address
// ascending; the caller uses the extra row (if present) to determine
// whether there is a next page, then drops it before returning to callers.
func ListAssets(ctx context.Context, q dbtx, filter AssetFilter) ([]model.Asset, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT asset, issuer, active, first_seen_ledger, last_updated_ledger
		FROM assets
		WHERE ($1::boolean IS NULL OR active = $1)
		  AND ($2::text IS NULL OR asset > $2)
		ORDER BY asset ASC
		LIMIT $3`,
		filter.Active, filter.Cursor, filter.Limit+1,
	)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var assets []model.Asset
	for rows.Next() {
		var a model.Asset
		if err := rows.Scan(&a.Asset, &a.Issuer, &a.Active, &a.FirstSeenLedger, &a.LastUpdatedLedger); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

// GetAsset returns the asset, or nil if it has never been indexed.
func GetAsset(ctx context.Context, q dbtx, asset string) (*model.Asset, error) {
	row := q.QueryRowContext(ctx,
		`SELECT asset, issuer, active, first_seen_ledger, last_updated_ledger FROM assets WHERE asset = $1`, asset)

	var a model.Asset
	if err := row.Scan(&a.Asset, &a.Issuer, &a.Active, &a.FirstSeenLedger, &a.LastUpdatedLedger); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan asset: %w", err)
	}
	return &a, nil
}

// ListOrdersByAsset returns the most recent orders for asset, newest first,
// bounded by limit.
func ListOrdersByAsset(ctx context.Context, q dbtx, asset string, limit int) ([]model.Order, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT order_id, buyer, distributor, asset, payment_asset, asset_amount, payment_amount,
		       created_at_ledger, expires_at_ledger, status, payment_funded, asset_funded,
		       created_transaction_hash, last_transaction_hash, last_updated_ledger
		FROM orders
		WHERE asset = $1
		ORDER BY created_at_ledger DESC, order_id DESC
		LIMIT $2`,
		asset, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list orders by asset: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows)
}

// scanOrders scans every remaining row of rows into model.Order values,
// matching the exact column order used by ListOrdersByAsset and the order
// list/detail queries added alongside the order API.
func scanOrders(rows *sql.Rows) ([]model.Order, error) {
	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.OrderID, &o.Buyer, &o.Distributor, &o.Asset, &o.PaymentAsset,
			&o.AssetAmount, &o.PaymentAmount, &o.CreatedAtLedger, &o.ExpiresAtLedger,
			&o.Status, &o.PaymentFunded, &o.AssetFunded,
			&o.CreatedTransactionHash, &o.LastTransactionHash, &o.LastUpdatedLedger,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}
