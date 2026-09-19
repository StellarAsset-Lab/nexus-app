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

func TestIntegrationAssetEndpoints(t *testing.T) {
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

	// Deterministic, sortable fixture asset addresses so pagination order
	// is predictable regardless of any other test's leftover rows.
	prefix := "CASSETAPITEST"
	for i := 0; i < 5; i++ {
		asset := fmt.Sprintf("%s%02d0000000000000000000000000000000000000000", prefix, i)
		if _, err := conn.ExecContext(ctx, `DELETE FROM assets WHERE asset = $1`, asset); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
		active := i%2 == 0
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO assets (asset, issuer, active, first_seen_ledger, last_updated_ledger)
			VALUES ($1, $2, $3, $4, $4)`,
			asset, "GISSUERTEST", active, 100+i,
		); err != nil {
			t.Fatalf("seed asset %d: %v", i, err)
		}
	}

	netConfig := network.Config{
		Name: network.Testnet, Passphrase: "Test SDF Network ; September 2015",
		RPCURL: "https://soroban-testnet.stellar.org", HorizonURL: "https://horizon-testnet.stellar.org",
	}
	server := NewServer(testLogger(), conn, netConfig)
	ts := httptest.NewServer(server.Routes())
	t.Cleanup(ts.Close)

	t.Run("pagination walks every fixture asset exactly once", func(t *testing.T) {
		seen := map[string]bool{}
		cursor := ""
		for page := 0; page < 10; page++ {
			url := fmt.Sprintf("%s/api/v1/assets?limit=2", ts.URL)
			if cursor != "" {
				url += "&cursor=" + cursor
			}
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("GET assets: %v", err)
			}
			var body listAssetsResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			resp.Body.Close()

			for _, a := range body.Assets {
				if len(a.Asset) >= len(prefix) && a.Asset[:len(prefix)] == prefix {
					if seen[a.Asset] {
						t.Fatalf("asset %s returned twice across pages", a.Asset)
					}
					seen[a.Asset] = true
				}
			}

			if body.NextCursor == nil {
				break
			}
			cursor = *body.NextCursor
		}

		if len(seen) != 5 {
			t.Fatalf("expected to see all 5 fixture assets via pagination, saw %d: %v", len(seen), seen)
		}
	})

	t.Run("active filter", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/assets?limit=200&active=true")
		if err != nil {
			t.Fatalf("GET assets: %v", err)
		}
		defer resp.Body.Close()
		var body listAssetsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		for _, a := range body.Assets {
			if len(a.Asset) >= len(prefix) && a.Asset[:len(prefix)] == prefix && !a.Active {
				t.Fatalf("active=true filter returned an inactive fixture asset: %s", a.Asset)
			}
		}
	})

	t.Run("get single asset", func(t *testing.T) {
		asset := fmt.Sprintf("%s000000000000000000000000000000000000000000", prefix)
		resp, err := http.Get(ts.URL + "/api/v1/assets/" + asset)
		if err != nil {
			t.Fatalf("GET asset: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var body assetResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Asset != asset {
			t.Fatalf("expected asset=%s, got %s", asset, body.Asset)
		}
	})

	t.Run("get missing asset returns stable 404 shape", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/assets/CDOESNOTEXIST0000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("GET asset: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
		var body errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.Error.Code != "ASSET_NOT_FOUND" {
			t.Fatalf("expected ASSET_NOT_FOUND, got %q", body.Error.Code)
		}
	})

	t.Run("limit is bounded even when caller asks for more", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/v1/assets?limit=999999")
		if err != nil {
			t.Fatalf("GET assets: %v", err)
		}
		defer resp.Body.Close()
		var body listAssetsResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(body.Assets) > maxPageLimit {
			t.Fatalf("expected at most %d assets, got %d", maxPageLimit, len(body.Assets))
		}
	})
}
