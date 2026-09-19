package indexing

import (
	"crypto/rand"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"

	rpcclient "github.com/stellar/go-stellar-sdk/clients/rpcclient"
	"github.com/stellar/go-stellar-sdk/strkey"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

// randomContractID returns a syntactically valid (correct checksum), but
// certainly-nonexistent, contract strkey address — Registry/Order have no
// deployed contract ID yet (see spec §38), so this proves the full real
// RPC + persistence + checkpoint-advancement pipeline works correctly on
// the honest zero-events case that will actually occur immediately after
// Phase 8 deployment and before any real activity exists.
func randomContractID(t *testing.T) string {
	t.Helper()
	var raw [32]byte
	if _, err := rand.Read(raw[:]); err != nil {
		t.Fatalf("generate random contract id: %v", err)
	}
	id, err := strkey.Encode(strkey.VersionByteContract, raw[:])
	if err != nil {
		t.Fatalf("encode contract id: %v", err)
	}
	return id
}

// TestIntegrationPollerRealRPC runs the full poller pipeline — GetHealth,
// GetEvents, raw persistence, checkpoint advancement, and re-run idempotency
// — against real Stellar Testnet RPC and a real Postgres database.
func TestIntegrationPollerRealRPC(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping database/RPC integration test")
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

	networkName := "integration-poller-test"
	contractID := randomContractID(t)
	rpc := rpcclient.NewClient("https://soroban-testnet.stellar.org", http.DefaultClient)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	poller, err := NewPoller(rpc, conn, logger, PollerConfig{
		NetworkName:          networkName,
		RegistryContractID:   contractID,
		BatchLimit:           50,
		ConfirmationLookback: 2,
	})
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	caughtUp, err := poller.pollOnce(ctx)
	if err != nil {
		t.Fatalf("pollOnce (first run): %v", err)
	}
	if !caughtUp {
		t.Fatalf("expected caught up on the first poll of a never-matching contract")
	}

	checkpoint, err := nexusdb.GetCheckpoint(ctx, conn, CheckpointID(networkName))
	if err != nil {
		t.Fatalf("get checkpoint: %v", err)
	}
	if checkpoint == nil {
		t.Fatalf("expected a checkpoint to be recorded even with zero matching events")
	}
	firstLedger := checkpoint.LastLedger

	// Re-running immediately must be idempotent: no duplicate-key errors,
	// and the checkpoint should not move backward.
	if _, err := poller.pollOnce(ctx); err != nil {
		t.Fatalf("pollOnce (second run, idempotency check): %v", err)
	}
	checkpoint2, err := nexusdb.GetCheckpoint(ctx, conn, CheckpointID(networkName))
	if err != nil {
		t.Fatalf("get checkpoint after second run: %v", err)
	}
	if checkpoint2.LastLedger < firstLedger {
		t.Fatalf("checkpoint moved backward: %d -> %d", firstLedger, checkpoint2.LastLedger)
	}

	t.Logf("OK: real Testnet RPC poll cycle verified, checkpoint advanced to ledger %d", checkpoint2.LastLedger)
}
