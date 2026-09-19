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

// OrderFilter narrows ListOrders. Every field is optional (nil/empty means
// "no filter") — callers combine only the filters they need.
type OrderFilter struct {
	Status        *string
	Asset         *string
	Distributor   *string
	Buyer         *string
	CreatedAfter  *int64 // inclusive, in created_at_ledger terms
	CreatedBefore *int64 // inclusive, in created_at_ledger terms
	// Cursor is the (created_at_ledger, order_id) pair of the last row seen
	// on the previous page — matches the composite
	// (created_at_ledger DESC, order_id DESC) ordering used throughout, per
	// spec §30 deterministic keyset pagination.
	CursorLedger  *int64
	CursorOrderID *int64
	Limit         int
}

// ListOrders returns up to filter.Limit+1 orders ordered by
// (created_at_ledger DESC, order_id DESC); the caller uses the extra row (if
// present) to determine whether there is a next page, then drops it.
func ListOrders(ctx context.Context, q dbtx, filter OrderFilter) ([]model.Order, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT order_id, buyer, distributor, asset, payment_asset, asset_amount, payment_amount,
		       created_at_ledger, expires_at_ledger, status, payment_funded, asset_funded,
		       created_transaction_hash, last_transaction_hash, last_updated_ledger
		FROM orders
		WHERE ($1::text IS NULL OR status = $1)
		  AND ($2::text IS NULL OR asset = $2)
		  AND ($3::text IS NULL OR distributor = $3)
		  AND ($4::text IS NULL OR buyer = $4)
		  AND ($5::bigint IS NULL OR created_at_ledger >= $5)
		  AND ($6::bigint IS NULL OR created_at_ledger <= $6)
		  AND ($7::bigint IS NULL OR $8::bigint IS NULL
		       OR (created_at_ledger, order_id) < ($7, $8))
		ORDER BY created_at_ledger DESC, order_id DESC
		LIMIT $9`,
		filter.Status, filter.Asset, filter.Distributor, filter.Buyer,
		filter.CreatedAfter, filter.CreatedBefore,
		filter.CursorLedger, filter.CursorOrderID,
		filter.Limit+1,
	)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows)
}

// GetOrder returns the order, or nil if it has never been indexed.
func GetOrder(ctx context.Context, q dbtx, orderID int64) (*model.Order, error) {
	row := q.QueryRowContext(ctx, `
		SELECT order_id, buyer, distributor, asset, payment_asset, asset_amount, payment_amount,
		       created_at_ledger, expires_at_ledger, status, payment_funded, asset_funded,
		       created_transaction_hash, last_transaction_hash, last_updated_ledger
		FROM orders WHERE order_id = $1`, orderID)

	var o model.Order
	if err := row.Scan(
		&o.OrderID, &o.Buyer, &o.Distributor, &o.Asset, &o.PaymentAsset,
		&o.AssetAmount, &o.PaymentAmount, &o.CreatedAtLedger, &o.ExpiresAtLedger,
		&o.Status, &o.PaymentFunded, &o.AssetFunded,
		&o.CreatedTransactionHash, &o.LastTransactionHash, &o.LastUpdatedLedger,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan order: %w", err)
	}
	return &o, nil
}

// PendingReconciliationHashes returns transaction hashes seen in
// stellar_events that either have no row in transactions yet, or whose
// transactions row is still non-terminal ("Pending") — the set the worker
// needs to (re-)check against RPC. Bounded by limit so one reconciliation
// pass never tries to process an unbounded backlog at once.
func PendingReconciliationHashes(ctx context.Context, q dbtx, limit int) ([]string, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT DISTINCT se.transaction_hash
		FROM stellar_events se
		LEFT JOIN transactions t ON t.transaction_hash = se.transaction_hash
		WHERE t.transaction_hash IS NULL OR t.status = 'Pending'
		ORDER BY se.transaction_hash
		LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("pending reconciliation hashes: %w", err)
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, fmt.Errorf("scan transaction hash: %w", err)
		}
		hashes = append(hashes, h)
	}
	return hashes, rows.Err()
}

// TransactionContext returns the contract id of the (deterministically)
// first event observed for hash, and the order id it relates to, if any.
// Both may be zero-valued when hash has no matching stellar_events row.
func TransactionContext(ctx context.Context, q dbtx, hash string) (contractID *string, orderID *int64, err error) {
	row := q.QueryRowContext(ctx, `
		SELECT se.contract_id, oe.order_id
		FROM stellar_events se
		LEFT JOIN order_events oe ON oe.transaction_hash = se.transaction_hash
		WHERE se.transaction_hash = $1
		ORDER BY se.ledger ASC, se.event_id ASC
		LIMIT 1`,
		hash,
	)
	var contract sql.NullString
	var order sql.NullInt64
	if scanErr := row.Scan(&contract, &order); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("scan transaction context: %w", scanErr)
	}
	if contract.Valid {
		contractID = &contract.String
	}
	if order.Valid {
		orderID = &order.Int64
	}
	return contractID, orderID, nil
}

// UpsertTransaction records the worker's reconciled observation of a
// transaction's status. FirstObservedAt is only set on first insert;
// LastObservedAt always advances to now().
func UpsertTransaction(ctx context.Context, q dbtx, t model.Transaction) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO transactions (
			transaction_hash, ledger, status, first_observed_at, last_observed_at,
			source_contract, order_id
		) VALUES ($1, $2, $3, now(), now(), $4, $5)
		ON CONFLICT (transaction_hash) DO UPDATE SET
			ledger = EXCLUDED.ledger,
			status = EXCLUDED.status,
			last_observed_at = now(),
			source_contract = COALESCE(EXCLUDED.source_contract, transactions.source_contract),
			order_id = COALESCE(EXCLUDED.order_id, transactions.order_id)`,
		t.TransactionHash, t.Ledger, t.Status, t.SourceContract, t.OrderID,
	)
	if err != nil {
		return fmt.Errorf("upsert transaction: %w", err)
	}
	return nil
}

// GetTransaction returns the transaction, or nil if it has never been
// observed.
func GetTransaction(ctx context.Context, q dbtx, hash string) (*model.Transaction, error) {
	row := q.QueryRowContext(ctx, `
		SELECT transaction_hash, ledger, status, first_observed_at, last_observed_at,
		       source_contract, order_id
		FROM transactions WHERE transaction_hash = $1`, hash)

	var t model.Transaction
	if err := row.Scan(
		&t.TransactionHash, &t.Ledger, &t.Status, &t.FirstObservedAt, &t.LastObservedAt,
		&t.SourceContract, &t.OrderID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan transaction: %w", err)
	}
	return &t, nil
}

// EventFilter narrows ListEvents. Every field is optional (nil/empty means
// "no filter").
type EventFilter struct {
	EventType  *string
	ContractID *string
	// Cursor is the (ledger, event_id) pair of the last row seen on the
	// previous page — matches the (ledger, event_id) ordering used
	// throughout for deterministic keyset pagination (spec §30).
	CursorLedger  *int64
	CursorEventID *string
	Limit         int
}

// ListEvents returns up to filter.Limit+1 events ordered by (ledger ASC,
// event_id ASC); the caller uses the extra row (if present) to determine
// whether there is a next page, then drops it.
func ListEvents(ctx context.Context, q dbtx, filter EventFilter) ([]model.StellarEvent, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT event_id, ledger, ledger_closed_at, transaction_hash, transaction_index,
		       operation_index, event_type, contract_id, topics_xdr, value_xdr,
		       decoded_payload, decode_error, first_observed_at
		FROM stellar_events
		WHERE ($1::text IS NULL OR event_type = $1)
		  AND ($2::text IS NULL OR contract_id = $2)
		  AND ($3::bigint IS NULL OR $4::text IS NULL
		       OR (ledger, event_id) > ($3, $4))
		ORDER BY ledger ASC, event_id ASC
		LIMIT $5`,
		filter.EventType, filter.ContractID,
		filter.CursorLedger, filter.CursorEventID,
		filter.Limit+1,
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []model.StellarEvent
	for rows.Next() {
		var e model.StellarEvent
		if err := rows.Scan(
			&e.EventID, &e.Ledger, &e.LedgerClosedAt, &e.TransactionHash, &e.TransactionIndex,
			&e.OperationIndex, &e.EventType, &e.ContractID, &e.TopicsXDR, &e.ValueXDR,
			&e.DecodedPayload, &e.DecodeError, &e.FirstObservedAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
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
