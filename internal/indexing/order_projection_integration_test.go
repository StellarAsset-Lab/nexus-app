package indexing

import (
	"database/sql"
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

func TestIntegrationProjectOrderEvent(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database integration test")
	}

	ctx := t.Context()
	conn, err := nexusdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := nexusdb.Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	buyer := "COREVPROJECTIONTESTBUYER0000000000000000000000000001"
	distributor := "COREVPROJECTIONTESTDIST00000000000000000000000000001"
	asset := "COREVPROJECTIONTESTASSET0000000000000000000000000001"
	paymentAsset := "COREVPROJECTIONTESTPAY000000000000000000000000000001"

	orderID1 := int64(910001)
	orderID2 := int64(910002)

	// Clean slate for this test's fixture order IDs.
	for _, id := range []int64{orderID1, orderID2} {
		if _, err := conn.ExecContext(ctx, `DELETE FROM order_events WHERE order_id = $1`, id); err != nil {
			t.Fatalf("cleanup order_events: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM orders WHERE order_id = $1`, id); err != nil {
			t.Fatalf("cleanup orders: %v", err)
		}
	}
	if _, err := conn.ExecContext(ctx, `DELETE FROM stellar_events WHERE event_id LIKE 'evt-%'`); err != nil {
		t.Fatalf("cleanup stellar_events: %v", err)
	}

	// order_events has a foreign key on stellar_events(event_id) — in the
	// real poller, the raw event row is always inserted before projection
	// (see poller.go). Mirror that here for each event_id this test uses.
	seedStellarEvent := func(tx *sql.Tx, eventID string, ledger int64, contractID string) {
		t.Helper()
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO stellar_events (
				event_id, ledger, ledger_closed_at, transaction_hash, transaction_index,
				operation_index, event_type, contract_id, topics_xdr, value_xdr
			) VALUES ($1, $2, now(), 'stub', 0, 0, 'contract', $3, '', '')
			ON CONFLICT (event_id) DO NOTHING`, eventID, ledger, contractID); err != nil {
			t.Fatalf("seed stellar_events for %s: %v", eventID, err)
		}
	}

	t.Run("full lifecycle: created, funded both sides, settled", func(t *testing.T) {
		assetAmount, _ := model.ParseAmount("170141183460469231731687303715884105727") // i128 max
		paymentAmount, _ := model.ParseAmount("5000000000000")

		withTx(t, conn, func(tx *sql.Tx) {
			seedStellarEvent(tx, "evt-created-1", 100, "CORDER")
			err := ProjectOrderEvent(ctx, tx, EventOrderCreated, map[string]any{
				"order_id": uint64(orderID1), "buyer": buyer, "distributor": distributor,
				"asset": asset, "payment_asset": paymentAsset,
				"asset_amount": assetAmount, "payment_amount": paymentAmount,
				"expires_at_ledger": uint32(5000),
			}, 100, "txhash-created", "evt-created-1")
			if err != nil {
				t.Fatalf("project OrderCreated: %v", err)
			}
		})

		var status string
		var paymentFunded, assetFunded bool
		var gotAssetAmount model.Amount
		row := conn.QueryRowContext(ctx, `SELECT status, payment_funded, asset_funded, asset_amount FROM orders WHERE order_id = $1`, orderID1)
		if err := row.Scan(&status, &paymentFunded, &assetFunded, &gotAssetAmount); err != nil {
			t.Fatalf("query order: %v", err)
		}
		if status != "Created" || paymentFunded || assetFunded {
			t.Fatalf("expected Created/unfunded, got status=%s paymentFunded=%v assetFunded=%v", status, paymentFunded, assetFunded)
		}
		if gotAssetAmount.String() != assetAmount.String() {
			t.Fatalf("expected exact asset_amount=%s, got %s", assetAmount.String(), gotAssetAmount.String())
		}

		withTx(t, conn, func(tx *sql.Tx) {
			seedStellarEvent(tx, "evt-payment-1", 110, "CORDER")
			if err := ProjectOrderEvent(ctx, tx, EventPaymentFunded, map[string]any{
				"order_id": uint64(orderID1), "buyer": buyer, "payment_amount": paymentAmount,
			}, 110, "txhash-payment", "evt-payment-1"); err != nil {
				t.Fatalf("project PaymentFunded: %v", err)
			}
			seedStellarEvent(tx, "evt-asset-1", 120, "CORDER")
			if err := ProjectOrderEvent(ctx, tx, EventAssetFunded, map[string]any{
				"order_id": uint64(orderID1), "distributor": distributor, "asset_amount": assetAmount,
			}, 120, "txhash-asset", "evt-asset-1"); err != nil {
				t.Fatalf("project AssetFunded: %v", err)
			}
		})

		row = conn.QueryRowContext(ctx, `SELECT status, payment_funded, asset_funded FROM orders WHERE order_id = $1`, orderID1)
		if err := row.Scan(&status, &paymentFunded, &assetFunded); err != nil {
			t.Fatalf("query order: %v", err)
		}
		if status != "Created" || !paymentFunded || !assetFunded {
			t.Fatalf("expected Created/both funded, got status=%s paymentFunded=%v assetFunded=%v", status, paymentFunded, assetFunded)
		}

		withTx(t, conn, func(tx *sql.Tx) {
			seedStellarEvent(tx, "evt-settled-1", 130, "CORDER")
			if err := ProjectOrderEvent(ctx, tx, EventOrderSettled, map[string]any{
				"order_id": uint64(orderID1), "buyer": buyer, "distributor": distributor,
				"asset_amount": assetAmount, "payment_amount": paymentAmount,
			}, 130, "txhash-settled", "evt-settled-1"); err != nil {
				t.Fatalf("project OrderSettled: %v", err)
			}
		})

		row = conn.QueryRowContext(ctx, `SELECT status FROM orders WHERE order_id = $1`, orderID1)
		if err := row.Scan(&status); err != nil {
			t.Fatalf("query order: %v", err)
		}
		if status != "Settled" {
			t.Fatalf("expected Settled, got %s", status)
		}

		t.Run("terminal state cannot be moved by a later duplicate/malformed event", func(t *testing.T) {
			// A stray/duplicate OrderCancelled arriving after settlement must
			// never flip a Settled order back to Cancelled.
			withTx(t, conn, func(tx *sql.Tx) {
				seedStellarEvent(tx, "evt-bogus-cancel-1", 140, "CORDER")
				if err := ProjectOrderEvent(ctx, tx, EventOrderCancelled, map[string]any{
					"order_id": uint64(orderID1), "cancelled_by": buyer,
				}, 140, "txhash-bogus-cancel", "evt-bogus-cancel-1"); err != nil {
					t.Fatalf("project stray OrderCancelled: %v", err)
				}
			})

			row := conn.QueryRowContext(ctx, `SELECT status FROM orders WHERE order_id = $1`, orderID1)
			if err := row.Scan(&status); err != nil {
				t.Fatalf("query order: %v", err)
			}
			if status != "Settled" {
				t.Fatalf("terminal status must not move backward; expected Settled, got %s", status)
			}
		})

		// order_events records every observed lifecycle event, including the
		// stray cancel above whose *status* update was correctly rejected —
		// the event was still genuinely observed and must not be hidden
		// from the audit trail (spec: never discard/hide raw observations).
		var eventCount int
		row = conn.QueryRowContext(ctx, `SELECT count(*) FROM order_events WHERE order_id = $1`, orderID1)
		if err := row.Scan(&eventCount); err != nil {
			t.Fatalf("count order_events: %v", err)
		}
		if eventCount != 5 {
			t.Fatalf("expected 5 order_events (created, payment, asset, settled, rejected-cancel), got %d", eventCount)
		}
	})

	t.Run("cancelled order", func(t *testing.T) {
		amount, _ := model.ParseAmount("1000")
		withTx(t, conn, func(tx *sql.Tx) {
			seedStellarEvent(tx, "evt-created-2", 10, "CORDER")
			if err := ProjectOrderEvent(ctx, tx, EventOrderCreated, map[string]any{
				"order_id": uint64(orderID2), "buyer": buyer, "distributor": distributor,
				"asset": asset, "payment_asset": paymentAsset,
				"asset_amount": amount, "payment_amount": amount, "expires_at_ledger": uint32(200),
			}, 10, "txhash-created2", "evt-created-2"); err != nil {
				t.Fatalf("project OrderCreated: %v", err)
			}
			seedStellarEvent(tx, "evt-cancelled-2", 20, "CORDER")
			if err := ProjectOrderEvent(ctx, tx, EventOrderCancelled, map[string]any{
				"order_id": uint64(orderID2), "cancelled_by": buyer,
			}, 20, "txhash-cancelled2", "evt-cancelled-2"); err != nil {
				t.Fatalf("project OrderCancelled: %v", err)
			}
		})

		var status string
		row := conn.QueryRowContext(ctx, `SELECT status FROM orders WHERE order_id = $1`, orderID2)
		if err := row.Scan(&status); err != nil {
			t.Fatalf("query order: %v", err)
		}
		if status != "Cancelled" {
			t.Fatalf("expected Cancelled, got %s", status)
		}
	})
}
