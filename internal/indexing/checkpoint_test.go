package indexing

import (
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
)

func TestResolveStartLedger(t *testing.T) {
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

	t.Run("no checkpoint, no configured start: uses oldest ledger", func(t *testing.T) {
		start, err := ResolveStartLedger(ctx, conn, "resolve-test-empty", nil, 100, 200)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if start != 100 {
			t.Fatalf("expected 100, got %d", start)
		}
	})

	t.Run("no checkpoint, configured start within range: uses configured start", func(t *testing.T) {
		configured := uint32(150)
		start, err := ResolveStartLedger(ctx, conn, "resolve-test-configured", &configured, 100, 200)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if start != 150 {
			t.Fatalf("expected 150, got %d", start)
		}
	})

	t.Run("configured start below oldest retained: clamped to oldest", func(t *testing.T) {
		configured := uint32(10)
		start, err := ResolveStartLedger(ctx, conn, "resolve-test-clamped-low", &configured, 100, 200)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if start != 100 {
			t.Fatalf("expected clamp to oldest ledger 100, got %d", start)
		}
	})

	t.Run("existing checkpoint takes precedence over configured start", func(t *testing.T) {
		id := "resolve-test-checkpoint-precedence"
		if err := nexusdb.UpsertCheckpoint(ctx, conn, id, 175, nil); err != nil {
			t.Fatalf("seed checkpoint: %v", err)
		}
		configured := uint32(105)
		start, err := ResolveStartLedger(ctx, conn, id, &configured, 100, 200)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if start != 176 {
			t.Fatalf("expected checkpoint+1=176 to take precedence over configured start, got %d", start)
		}
	})

	t.Run("checkpoint beyond latest ledger is clamped", func(t *testing.T) {
		id := "resolve-test-checkpoint-overshoot"
		if err := nexusdb.UpsertCheckpoint(ctx, conn, id, 199, nil); err != nil {
			t.Fatalf("seed checkpoint: %v", err)
		}
		start, err := ResolveStartLedger(ctx, conn, id, nil, 100, 200)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if start != 200 {
			t.Fatalf("expected clamp to latest ledger 200, got %d", start)
		}
	})
}
