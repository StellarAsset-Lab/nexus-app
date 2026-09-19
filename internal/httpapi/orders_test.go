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

func TestIntegrationOrderEndpoints(t *testing.T) {
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

	// Deterministic fixture orders: order IDs double as a stand-in for
	// insertion order so cursor-based pagination is predictable regardless
	// of any other test's leftover rows.
	const orderIDBase = 900_000_000
	statuses := []string{"Created", "Created", "Settled", "Cancelled", "Created"}
	for i := 0; i < 5; i++ {
		orderID := int64(orderIDBase + i)
		if _, err := conn.ExecContext(ctx, `DELETE FROM orders WHERE order_id = $1`, orderID); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO orders (
				order_id, buyer, distributor, asset, payment_asset,
				asset_amount, payment_amount, created_at_ledger, expires_at_ledger,
				status, payment_funded, asset_funded,
				created_transaction_hash, last_transaction_hash, last_updated_ledger
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
			orderID, "GBUYERTEST", "GDISTTEST", "CORDERTESTASSET", "CORDERTESTPAYMENT",
			"1000", "2000", 200+i, 300+i,
			statuses[i], true, false,
			fmt.Sprintf("txcreate%d", i), fmt.Sprintf("txlast%d", i), 200+i,
		); err != nil {
			t.Fatalf("seed order %d: %v", i, err)
		}
	}

	netConfig := network.Config{
		Name: network.Testnet, Passphrase: "Test SDF Network ; September 2015",
		RPCURL: "https://soroban-testnet.stellar.org", HorizonURL: "https://horizon-testnet.stellar.org",
	}
	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	t.Run("pagination walks every fixture order exactly once", func(t *testing.T) {
		seen := map[int64]bool{}
		cursor := ""
		for page := 0; page < 10; page++ {
			url := fmt.Sprintf("%s/api/v1/orders?asset=CORDERTESTASSET&limit=2", ts.URL)
			if cursor != "" {
				url += "&cursor=" + cursor
			}
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("GET orders: %v", err)
			}
			var body listOrdersResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			resp.Body.Close()

			for _, o := range body.Orders {
				if o.OrderID >= orderIDBase && o.OrderID < orderIDBase+5 {
					if seen[o.OrderID] {
						t.Fatalf("order %d returned twice across pages", o.OrderID)
					}
					seen[o.OrderID] = true
				}
			}

			if body.NextCursor == nil {
				break
			}
			cursor = *body.NextCursor
		}

		if len(seen) != 5 {
			t.Fatalf("expected to see all 5 fixture orders via pagination, saw %d: %v", len(seen), seen)
		}
	})

	t.Run("status filter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/orders?asset=CORDERTESTASSET&status=Settled&limit=200")
		if err != nil {
			t.Fatalf("GET orders: %v", err)
		}
		defer resp.Body.Close()
		var body listOrdersResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		for _, o := range body.Orders {
			if o.OrderID >= orderIDBase && o.OrderID < orderIDBase+5 && o.Status != "Settled" {
				t.Fatalf("status=Settled filter returned order with status %q", o.Status)
			}
		}
	})

	t.Run("get single order", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/orders/%d", ts.URL, int64(orderIDBase)))
		if err != nil {
			t.Fatalf("GET order: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var body orderSummary
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.OrderID != orderIDBase {
			t.Fatalf("expected orderId=%d, got %d", orderIDBase, body.OrderID)
		}
		if body.AssetAmount != "1000" {
			t.Fatalf("expected assetAmount=1000, got %s", body.AssetAmount)
		}
	})

	t.Run("get missing order returns stable 404 shape", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/orders/999999999999")
		if err != nil {
			t.Fatalf("GET order: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
		var body errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Error.Code != "ORDER_NOT_FOUND" {
			t.Fatalf("expected ORDER_NOT_FOUND, got %q", body.Error.Code)
		}
	})

	t.Run("get order with non-integer id returns 400", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/orders/not-an-id")
		if err != nil {
			t.Fatalf("GET order: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
		var body errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Error.Code != "INVALID_ORDER_ID" {
			t.Fatalf("expected INVALID_ORDER_ID, got %q", body.Error.Code)
		}
	})
}
