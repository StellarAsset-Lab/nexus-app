package reconcile

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

// randomTransactionHash returns a syntactically valid (64 hex chars) but
// certainly-nonexistent transaction hash, so the test can prove the full
// real RPC + persistence pipeline works on the honest "not yet ingested by
// RPC" case, without needing a real live Testnet transaction on hand.
func randomTransactionHash(t *testing.T) string {
	t.Helper()
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		t.Fatalf("generate random hash: %v", err)
	}
	return hex.EncodeToString(raw[:])
}

// TestIntegrationReconcileRealRPC runs the full reconciliation pipeline —
// pending-hash discovery, real getTransaction RPC calls, and upsert — against
// real Stellar Testnet RPC and a real Postgres database.
func TestIntegrationReconcileRealRPC(t *testing.T) {
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

	hash := randomTransactionHash(t)
	eventID := hash + "-0000000001"
	contractID := "CRECONCILETESTCONTRACT"

	if _, err := conn.ExecContext(ctx, `DELETE FROM transactions WHERE transaction_hash = $1`, hash); err != nil {
		t.Fatalf("cleanup transactions: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `DELETE FROM stellar_events WHERE event_id = $1`, eventID); err != nil {
		t.Fatalf("cleanup stellar_events: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO stellar_events (
			event_id, ledger, ledger_closed_at, transaction_hash, transaction_index,
			operation_index, event_type, contract_id, topics_xdr, value_xdr
		) VALUES ($1, 600, now(), $2, 0, 0, 'AssetListed', $3, 'topics', 'value')`,
		eventID, hash, contractID,
	); err != nil {
		t.Fatalf("seed stellar_events: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	rpc := rpcclient.NewClient("https://soroban-testnet.stellar.org", http.DefaultClient)
	reconciler := NewReconciler(rpc, conn, logger)

	t.Run("pending hash discovered before reconciliation", func(t *testing.T) {
		hashes, err := nexusdb.PendingReconciliationHashes(ctx, conn, 1000)
		if err != nil {
			t.Fatalf("pending hashes: %v", err)
		}
		found := false
		for _, h := range hashes {
			if h == hash {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected %s to be in the pending reconciliation set", hash)
		}
	})

	t.Run("run reconciles the hash against real RPC", func(t *testing.T) {
		n, err := reconciler.Run(ctx)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if n == 0 {
			t.Fatalf("expected at least one hash to be processed")
		}

		tx, err := nexusdb.GetTransaction(ctx, conn, hash)
		if err != nil {
			t.Fatalf("get transaction: %v", err)
		}
		if tx == nil {
			t.Fatalf("expected a transactions row for %s after reconciliation", hash)
		}
		// A random hash will never actually exist on Testnet, so RPC must
		// report NOT_FOUND, which this package maps to "Pending" — proving
		// the real RPC round trip happened rather than being faked.
		if tx.Status != "Pending" {
			t.Fatalf("expected status=Pending for a nonexistent hash, got %s", tx.Status)
		}
		if tx.SourceContract == nil || *tx.SourceContract != contractID {
			t.Fatalf("expected sourceContract=%s, got %v", contractID, tx.SourceContract)
		}
	})

	t.Run("hash no longer pending after reaching a terminal status", func(t *testing.T) {
		if _, err := conn.ExecContext(ctx, `UPDATE transactions SET status = 'Confirmed' WHERE transaction_hash = $1`, hash); err != nil {
			t.Fatalf("force confirmed: %v", err)
		}
		hashes, err := nexusdb.PendingReconciliationHashes(ctx, conn, 1000)
		if err != nil {
			t.Fatalf("pending hashes: %v", err)
		}
		for _, h := range hashes {
			if h == hash {
				t.Fatalf("expected %s to no longer be pending after reaching Confirmed", hash)
			}
		}
	})
}
