package db_test

import (
	"context"
	"math/big"
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

// Real database integration test. Skips when DATABASE_URL is not set (the
// default for `go test ./...` without a running Postgres) rather than mocking
// the database — see spec's testing philosophy: integration tests should hit
// a real database. Run with, e.g.:
//
//	DATABASE_URL=postgres://postgres:postgres@localhost:5432/nexus?sslmode=disable go test ./internal/db/...
func TestIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database integration test")
	}

	ctx := context.Background()
	conn, err := nexusdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if err := nexusdb.Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Run("migrate is idempotent", func(t *testing.T) {
		if err := nexusdb.Migrate(ctx, conn); err != nil {
			t.Fatalf("re-migrate: %v", err)
		}
	})

	t.Run("checkpoint round trip", func(t *testing.T) {
		if err := nexusdb.UpsertCheckpoint(ctx, conn, "test-integration", 12345, nil); err != nil {
			t.Fatalf("upsert checkpoint: %v", err)
		}
		cp, err := nexusdb.GetCheckpoint(ctx, conn, "test-integration")
		if err != nil {
			t.Fatalf("get checkpoint: %v", err)
		}
		if cp == nil || cp.LastLedger != 12345 {
			t.Fatalf("checkpoint round-trip mismatch: %+v", cp)
		}
	})

	t.Run("missing checkpoint returns nil not error", func(t *testing.T) {
		missing, err := nexusdb.GetCheckpoint(ctx, conn, "does-not-exist-integration")
		if err != nil {
			t.Fatalf("get missing checkpoint: %v", err)
		}
		if missing != nil {
			t.Fatalf("expected nil for missing checkpoint, got %+v", missing)
		}
	})

	t.Run("i128-scale amount round-trips exactly", func(t *testing.T) {
		huge, ok := new(big.Int).SetString("170141183460469231731687303715884105727", 10) // i128 max
		if !ok {
			t.Fatal("bad test literal")
		}

		if _, err := conn.ExecContext(ctx, `
			INSERT INTO assets (asset, issuer, active, first_seen_ledger, last_updated_ledger)
			VALUES ('CINTEGRATIONTEST', 'GINTEGRATIONTEST', true, 1, 1)
			ON CONFLICT (asset) DO NOTHING
		`); err != nil {
			t.Fatalf("insert asset: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO orders (order_id, buyer, distributor, asset, payment_asset, asset_amount, payment_amount,
				created_at_ledger, expires_at_ledger, status, created_transaction_hash, last_transaction_hash, last_updated_ledger)
			VALUES (999999999, 'GBUYERTEST', 'GDISTTEST', 'CINTEGRATIONTEST', 'CPAYTEST', $1, $1, 1, 100, 'Created', 'deadbeef', 'deadbeef', 1)
			ON CONFLICT (order_id) DO UPDATE SET asset_amount = EXCLUDED.asset_amount
		`, huge.String()); err != nil {
			t.Fatalf("insert order with i128-max amount: %v", err)
		}

		var scanned model.Amount
		row := conn.QueryRowContext(ctx, `SELECT asset_amount FROM orders WHERE order_id = 999999999`)
		if err := row.Scan(&scanned); err != nil {
			t.Fatalf("scan amount: %v", err)
		}
		if scanned.String() != huge.String() {
			t.Fatalf("exact amount round-trip mismatch: got %s want %s", scanned.String(), huge.String())
		}

		jsonBytes, err := scanned.MarshalJSON()
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if string(jsonBytes) != `"`+huge.String()+`"` {
			t.Fatalf("expected JSON string representation, got %s", jsonBytes)
		}
	})
}
