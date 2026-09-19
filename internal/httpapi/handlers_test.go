package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHandleHealthz(t *testing.T) {
	server := &Server{logger: testLogger()}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	server.handleHealthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %+v", body)
	}
}

func TestWriteErrorShape(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}

	var body errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != "ORDER_NOT_FOUND" || body.Error.Message != "Order not found" {
		t.Fatalf("unexpected error body: %+v", body)
	}
}

// TestIntegrationReadyz exercises /healthz, /readyz, and /api/v1/network
// against a real database and, for the network check, real Stellar Testnet
// RPC — skipped unless DATABASE_URL is set (see internal/db's integration
// test for the same pattern).
func TestIntegrationReadyz(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping httpapi integration test")
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

	netConfig := network.Config{
		Name:       network.Testnet,
		Passphrase: "Test SDF Network ; September 2015",
		RPCURL:     "https://soroban-testnet.stellar.org",
		HorizonURL: "https://horizon-testnet.stellar.org",
	}

	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	t.Run("readyz is 200 with a live database", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/readyz")
		if err != nil {
			t.Fatalf("GET /readyz: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("readyz is 503 once the database is closed", func(t *testing.T) {
		_ = conn.Close()
		resp, err := http.Get(ts.URL + "/readyz")
		if err != nil {
			t.Fatalf("GET /readyz: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected 503 once the database is unreachable, got %d", resp.StatusCode)
		}
		var body errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Error.Code != "NOT_READY" {
			t.Fatalf("expected NOT_READY error code, got %q", body.Error.Code)
		}
	})
}

// TestIntegrationNetworkEndpoint exercises /api/v1/network against real
// Stellar Testnet RPC — proving RPC health is measured, not assumed.
func TestIntegrationNetworkEndpoint(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set; skipping httpapi integration test")
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

	netConfig := network.Config{
		Name:       network.Testnet,
		Passphrase: "Test SDF Network ; September 2015",
		RPCURL:     "https://soroban-testnet.stellar.org",
		HorizonURL: "https://horizon-testnet.stellar.org",
	}

	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/api/v1/network")
	if err != nil {
		t.Fatalf("GET /api/v1/network: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body networkResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Network != "testnet" {
		t.Fatalf("expected network=testnet, got %q", body.Network)
	}
	if !body.RPCHealthy {
		t.Fatalf("expected real Testnet RPC to report healthy")
	}
	if body.LatestLedger == nil || *body.LatestLedger == 0 {
		t.Fatalf("expected a real latest ledger from Testnet RPC, got %v", body.LatestLedger)
	}
}
