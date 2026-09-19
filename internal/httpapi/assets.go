package httpapi

import (
	"net/http"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

type assetResponse struct {
	Asset             string `json:"asset"`
	Issuer            string `json:"issuer"`
	Active            bool   `json:"active"`
	FirstSeenLedger   int64  `json:"firstSeenLedger"`
	LastUpdatedLedger int64  `json:"lastUpdatedLedger"`
}

func toAssetResponse(a model.Asset) assetResponse {
	return assetResponse{
		Asset: a.Asset, Issuer: a.Issuer, Active: a.Active,
		FirstSeenLedger: a.FirstSeenLedger, LastUpdatedLedger: a.LastUpdatedLedger,
	}
}

type listAssetsResponse struct {
	Assets     []assetResponse `json:"assets"`
	NextCursor *string         `json:"nextCursor,omitempty"`
}

func (s *Server) handleListAssets(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)

	assets, err := nexusdb.ListAssets(r.Context(), s.db, nexusdb.AssetFilter{
		Active: parseBoolFilter(r, "active"),
		Cursor: parseCursor(r),
		Limit:  limit,
	})
	if err != nil {
		s.logger.Error("list assets failed", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list assets.")
		return
	}

	resp := listAssetsResponse{Assets: make([]assetResponse, 0, len(assets))}
	hasMore := len(assets) > limit
	if hasMore {
		assets = assets[:limit]
	}
	for _, a := range assets {
		resp.Assets = append(resp.Assets, toAssetResponse(a))
	}
	if hasMore && len(assets) > 0 {
		cursor := assets[len(assets)-1].Asset
		resp.NextCursor = &cursor
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetAsset(w http.ResponseWriter, r *http.Request) {
	asset := r.PathValue("asset")

	found, err := nexusdb.GetAsset(r.Context(), s.db, asset)
	if err != nil {
		s.logger.Error("get asset failed", "error", err, "asset", asset)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load asset.")
		return
	}
	if found == nil {
		writeError(w, http.StatusNotFound, "ASSET_NOT_FOUND", "Asset not found.")
		return
	}

	writeJSON(w, http.StatusOK, toAssetResponse(*found))
}

// orderSummary is the minimal order shape needed for an asset's activity
// feed. The dedicated Order API endpoints (added alongside GET
// /api/v1/orders) may reuse or extend this.
type orderSummary struct {
	OrderID                int64  `json:"orderId"`
	Buyer                  string `json:"buyer"`
	Distributor            string `json:"distributor"`
	Asset                  string `json:"asset"`
	PaymentAsset           string `json:"paymentAsset"`
	AssetAmount            string `json:"assetAmount"`
	PaymentAmount          string `json:"paymentAmount"`
	CreatedAtLedger        int64  `json:"createdAtLedger"`
	ExpiresAtLedger        int64  `json:"expiresAtLedger"`
	Status                 string `json:"status"`
	PaymentFunded          bool   `json:"paymentFunded"`
	AssetFunded            bool   `json:"assetFunded"`
	CreatedTransactionHash string `json:"createdTransactionHash"`
	LastTransactionHash    string `json:"lastTransactionHash"`
	LastUpdatedLedger      int64  `json:"lastUpdatedLedger"`
}

func toOrderResponses(orders []model.Order) []orderSummary {
	out := make([]orderSummary, 0, len(orders))
	for _, o := range orders {
		out = append(out, orderSummary{
			OrderID: o.OrderID, Buyer: o.Buyer, Distributor: o.Distributor,
			Asset: o.Asset, PaymentAsset: o.PaymentAsset,
			AssetAmount: o.AssetAmount.String(), PaymentAmount: o.PaymentAmount.String(),
			CreatedAtLedger: o.CreatedAtLedger, ExpiresAtLedger: o.ExpiresAtLedger,
			Status: string(o.Status), PaymentFunded: o.PaymentFunded, AssetFunded: o.AssetFunded,
			CreatedTransactionHash: o.CreatedTransactionHash, LastTransactionHash: o.LastTransactionHash,
			LastUpdatedLedger: o.LastUpdatedLedger,
		})
	}
	return out
}

func (s *Server) handleAssetActivity(w http.ResponseWriter, r *http.Request) {
	asset := r.PathValue("asset")
	limit := parseLimit(r)

	found, err := nexusdb.GetAsset(r.Context(), s.db, asset)
	if err != nil {
		s.logger.Error("get asset failed", "error", err, "asset", asset)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load asset.")
		return
	}
	if found == nil {
		writeError(w, http.StatusNotFound, "ASSET_NOT_FOUND", "Asset not found.")
		return
	}

	orders, err := nexusdb.ListOrdersByAsset(r.Context(), s.db, asset, limit)
	if err != nil {
		s.logger.Error("list orders by asset failed", "error", err, "asset", asset)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load asset activity.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"asset":  asset,
		"orders": toOrderResponses(orders),
	})
}
