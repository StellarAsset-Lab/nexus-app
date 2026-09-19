package indexing

import (
	"database/sql"
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

func withTx(t *testing.T, conn *sql.DB, fn func(tx *sql.Tx)) {
	t.Helper()
	ctx := t.Context()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	fn(tx)
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}
}

func TestIntegrationProjectRegistryEvent(t *testing.T) {
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

	asset := "CPROJECTIONTESTASSET00000000000000000000000000000001"
	issuer := "CPROJECTIONTESTISSUER0000000000000000000000000000001"
	distributor := "CPROJECTIONTESTDIST000000000000000000000000000000001"
	authority := "CPROJECTIONTESTAUTH000000000000000000000000000000001"
	buyer := "CPROJECTIONTESTBUYER00000000000000000000000000000001"
	asset2 := "CPROJECTIONTESTASSET00000000000000000000000000000002"

	// This test's assertions depend on starting from a clean slate for its
	// fixture addresses; without this, a second run against the same
	// persistent database (e.g. re-running `go test` locally) would see
	// leftover rows from the previous run and the out-of-order guard would
	// correctly, but confusingly, reject re-registration at a lower ledger.
	for _, a := range []string{asset, asset2} {
		if _, err := conn.ExecContext(ctx, `DELETE FROM eligibilities WHERE asset = $1`, a); err != nil {
			t.Fatalf("cleanup eligibilities: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM distributions WHERE asset = $1`, a); err != nil {
			t.Fatalf("cleanup distributions: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM assets WHERE asset = $1`, a); err != nil {
			t.Fatalf("cleanup assets: %v", err)
		}
	}

	t.Run("asset lifecycle: registered then deactivated", func(t *testing.T) {
		withTx(t, conn, func(tx *sql.Tx) {
			if err := ProjectRegistryEvent(ctx, tx, EventAssetRegistered, map[string]any{"asset": asset, "issuer": issuer}, 100); err != nil {
				t.Fatalf("project AssetRegistered: %v", err)
			}
		})

		var active bool
		var lastLedger int64
		row := conn.QueryRowContext(ctx, `SELECT active, last_updated_ledger FROM assets WHERE asset = $1`, asset)
		if err := row.Scan(&active, &lastLedger); err != nil {
			t.Fatalf("query asset: %v", err)
		}
		if !active || lastLedger != 100 {
			t.Fatalf("expected active=true, ledger=100 after registration; got active=%v ledger=%d", active, lastLedger)
		}

		withTx(t, conn, func(tx *sql.Tx) {
			if err := ProjectRegistryEvent(ctx, tx, EventAssetDeactivated, map[string]any{"asset": asset}, 200); err != nil {
				t.Fatalf("project AssetDeactivated: %v", err)
			}
		})

		row = conn.QueryRowContext(ctx, `SELECT active, last_updated_ledger FROM assets WHERE asset = $1`, asset)
		if err := row.Scan(&active, &lastLedger); err != nil {
			t.Fatalf("query asset: %v", err)
		}
		if active || lastLedger != 200 {
			t.Fatalf("expected active=false, ledger=200 after deactivation; got active=%v ledger=%d", active, lastLedger)
		}

		t.Run("stale out-of-order event does not overwrite newer state", func(t *testing.T) {
			// A duplicate/reordered AssetRegistered at an OLDER ledger than
			// the current projection must not resurrect the asset — the
			// projection must never move backward in ledger time.
			withTx(t, conn, func(tx *sql.Tx) {
				if err := ProjectRegistryEvent(ctx, tx, EventAssetRegistered, map[string]any{"asset": asset, "issuer": issuer}, 150); err != nil {
					t.Fatalf("project stale AssetRegistered: %v", err)
				}
			})

			row := conn.QueryRowContext(ctx, `SELECT active, last_updated_ledger FROM assets WHERE asset = $1`, asset)
			if err := row.Scan(&active, &lastLedger); err != nil {
				t.Fatalf("query asset: %v", err)
			}
			if active || lastLedger != 200 {
				t.Fatalf("stale event must not overwrite newer state; expected active=false ledger=200, got active=%v ledger=%d", active, lastLedger)
			}
		})
	})

	t.Run("distribution and eligibility lifecycle", func(t *testing.T) {
		asset := asset2

		withTx(t, conn, func(tx *sql.Tx) {
			if err := ProjectRegistryEvent(ctx, tx, EventAssetRegistered, map[string]any{"asset": asset, "issuer": issuer}, 50); err != nil {
				t.Fatalf("project AssetRegistered: %v", err)
			}
			if err := ProjectRegistryEvent(ctx, tx, EventDistributionRegistered, map[string]any{
				"asset": asset, "distributor": distributor, "eligibility_authority": authority,
			}, 60); err != nil {
				t.Fatalf("project DistributionRegistered: %v", err)
			}
			if err := ProjectRegistryEvent(ctx, tx, EventEligibilitySet, map[string]any{
				"asset": asset, "distributor": distributor, "buyer": buyer, "valid_until_ledger": uint32(9999),
			}, 70); err != nil {
				t.Fatalf("project EligibilitySet: %v", err)
			}
		})

		var distActive, eligActive bool
		row := conn.QueryRowContext(ctx, `SELECT active FROM distributions WHERE asset = $1 AND distributor = $2`, asset, distributor)
		if err := row.Scan(&distActive); err != nil {
			t.Fatalf("query distribution: %v", err)
		}
		row = conn.QueryRowContext(ctx, `SELECT active FROM eligibilities WHERE asset = $1 AND distributor = $2 AND buyer = $3`, asset, distributor, buyer)
		if err := row.Scan(&eligActive); err != nil {
			t.Fatalf("query eligibility: %v", err)
		}
		if !distActive || !eligActive {
			t.Fatalf("expected both distribution and eligibility active, got distribution=%v eligibility=%v", distActive, eligActive)
		}

		withTx(t, conn, func(tx *sql.Tx) {
			if err := ProjectRegistryEvent(ctx, tx, EventEligibilityRevoked, map[string]any{
				"asset": asset, "distributor": distributor, "buyer": buyer,
			}, 80); err != nil {
				t.Fatalf("project EligibilityRevoked: %v", err)
			}
			if err := ProjectRegistryEvent(ctx, tx, EventDistributionRevoked, map[string]any{
				"asset": asset, "distributor": distributor,
			}, 90); err != nil {
				t.Fatalf("project DistributionRevoked: %v", err)
			}
		})

		row = conn.QueryRowContext(ctx, `SELECT active FROM distributions WHERE asset = $1 AND distributor = $2`, asset, distributor)
		if err := row.Scan(&distActive); err != nil {
			t.Fatalf("query distribution: %v", err)
		}
		row = conn.QueryRowContext(ctx, `SELECT active FROM eligibilities WHERE asset = $1 AND distributor = $2 AND buyer = $3`, asset, distributor, buyer)
		if err := row.Scan(&eligActive); err != nil {
			t.Fatalf("query eligibility: %v", err)
		}
		if distActive || eligActive {
			t.Fatalf("expected both revoked, got distribution=%v eligibility=%v", distActive, eligActive)
		}
	})
}
