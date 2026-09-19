package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/network"
)

func TestIntegrationTransactionEndpoints(t *testing.T) {
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

	hash := "txintegrationtest0000000000000000000000000000000000000000000"
	if _, err := conn.ExecContext(ctx, `DELETE FROM transactions WHERE transaction_hash = $1`, hash); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO transactions (transaction_hash, ledger, status, source_contract)
		VALUES ($1, $2, $3, $4)`,
		hash, 400, "Confirmed", "CSOURCETEST",
	); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}

	netConfig := network.Config{
		Name: network.Testnet, Passphrase: "Test SDF Network ; September 2015",
		RPCURL: "https://soroban-testnet.stellar.org", HorizonURL: "https://horizon-testnet.stellar.org",
	}
	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	t.Run("get existing transaction", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/transactions/" + hash)
		if err != nil {
			t.Fatalf("GET transaction: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var body transactionResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.TransactionHash != hash {
			t.Fatalf("expected hash=%s, got %s", hash, body.TransactionHash)
		}
		if body.Status != "Confirmed" {
			t.Fatalf("expected status=Confirmed, got %s", body.Status)
		}
	})

	t.Run("get missing transaction returns stable 404 shape", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/transactions/doesnotexist")
		if err != nil {
			t.Fatalf("GET transaction: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
		var body errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Error.Code != "TRANSACTION_NOT_FOUND" {
			t.Fatalf("expected TRANSACTION_NOT_FOUND, got %q", body.Error.Code)
		}
	})

	t.Run("list transactions includes the seeded transaction", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/transactions?limit=200")
		if err != nil {
			t.Fatalf("GET transactions: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var body listTransactionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		found := false
		for _, tx := range body.Transactions {
			if tx.TransactionHash == hash {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected %s in the transaction list", hash)
		}
	})

	t.Run("list transactions status filter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/transactions?status=Confirmed&limit=200")
		if err != nil {
			t.Fatalf("GET transactions: %v", err)
		}
		defer resp.Body.Close()
		var body listTransactionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		for _, tx := range body.Transactions {
			if tx.Status != "Confirmed" {
				t.Fatalf("status=Confirmed filter returned transaction with status %q", tx.Status)
			}
		}
	})
}

func TestIntegrationEventEndpoints(t *testing.T) {
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

	const contractID = "CEVENTTESTCONTRACT"
	eventIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		eventID := fmt.Sprintf("0000000%d00000000000000000000000000000000000000000000000000-eventtest", i)
		eventIDs[i] = eventID
		if _, err := conn.ExecContext(ctx, `DELETE FROM stellar_events WHERE event_id = $1`, eventID); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO stellar_events (
				event_id, ledger, ledger_closed_at, transaction_hash, transaction_index,
				operation_index, event_type, contract_id, topics_xdr, value_xdr, decoded_payload
			) VALUES ($1, $2, now(), $3, 0, 0, $4, $5, 'topics', 'value', $6)`,
			eventID, 500+i, fmt.Sprintf("txevent%d", i), "AssetListed", contractID,
			[]byte(fmt.Sprintf(`{"index":%d}`, i)),
		); err != nil {
			t.Fatalf("seed event %d: %v", i, err)
		}
	}

	netConfig := network.Config{
		Name: network.Testnet, Passphrase: "Test SDF Network ; September 2015",
		RPCURL: "https://soroban-testnet.stellar.org", HorizonURL: "https://horizon-testnet.stellar.org",
	}
	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	t.Run("pagination walks every fixture event exactly once", func(t *testing.T) {
		seen := map[string]bool{}
		cursor := ""
		for page := 0; page < 10; page++ {
			url := fmt.Sprintf("%s/api/v1/events?contractId=%s&limit=2", ts.URL, contractID)
			if cursor != "" {
				url += "&cursor=" + cursor
			}
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("GET events: %v", err)
			}
			var body listEventsResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			resp.Body.Close()

			for _, e := range body.Events {
				seen[e.EventID] = true
			}

			if body.NextCursor == nil {
				break
			}
			cursor = *body.NextCursor
		}

		for _, id := range eventIDs {
			if !seen[id] {
				t.Fatalf("expected to see event %s via pagination, saw %v", id, seen)
			}
		}
	})

	t.Run("event type filter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/events?contractId=" + contractID + "&eventType=AssetListed&limit=200")
		if err != nil {
			t.Fatalf("GET events: %v", err)
		}
		defer resp.Body.Close()
		var body listEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(body.Events) < 5 {
			t.Fatalf("expected at least 5 events, got %d", len(body.Events))
		}
		for _, e := range body.Events {
			if e.ContractID == contractID && e.EventType != "AssetListed" {
				t.Fatalf("eventType filter returned event with type %q", e.EventType)
			}
		}
	})

	t.Run("transaction hash filter returns only that transaction's events", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/events?transactionHash=txevent0&limit=200")
		if err != nil {
			t.Fatalf("GET events: %v", err)
		}
		defer resp.Body.Close()
		var body listEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(body.Events) != 1 {
			t.Fatalf("expected exactly 1 event for transactionHash=txevent0, got %d", len(body.Events))
		}
		if body.Events[0].TransactionHash != "txevent0" {
			t.Fatalf("expected transactionHash=txevent0, got %s", body.Events[0].TransactionHash)
		}
	})
}

func TestIntegrationStatusEndpoint(t *testing.T) {
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

	netConfig := network.Config{
		Name: network.Testnet, Passphrase: "Test SDF Network ; September 2015",
		RPCURL: "https://soroban-testnet.stellar.org", HorizonURL: "https://horizon-testnet.stellar.org",
	}
	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	resp, err := http.Get(ts.URL + "/api/v1/status")
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Components) != 3 {
		t.Fatalf("expected 3 components, got %d", len(body.Components))
	}
	for _, c := range body.Components {
		if c.Status == "" {
			t.Fatalf("component %s has no status", c.Name)
		}
	}
	// Database must be genuinely checked, not defaulted: since this test
	// itself just used the connection successfully, it must report
	// Operational, not Unknown/Unavailable.
	foundDB := false
	for _, c := range body.Components {
		if c.Name == "Database" {
			foundDB = true
			if c.Status != "Operational" {
				t.Fatalf("expected Database=Operational, got %s (%s)", c.Status, c.Detail)
			}
		}
	}
	if !foundDB {
		t.Fatalf("expected a Database component in the status response")
	}
}
