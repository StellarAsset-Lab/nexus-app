package httpapi

import (
	"net/http"
	"strconv"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

type listOrdersResponse struct {
	Orders     []orderSummary `json:"orders"`
	NextCursor *string        `json:"nextCursor,omitempty"`
}

// encodeOrderCursor/decodeOrderCursor represent the composite
// (created_at_ledger, order_id) keyset cursor as "ledger:orderId" — the same
// pair the list query orders and filters by, per spec §30.
func encodeOrderCursor(ledger, orderID int64) string {
	return strconv.FormatInt(ledger, 10) + ":" + strconv.FormatInt(orderID, 10)
}

func decodeOrderCursor(raw string) (ledger int64, orderID int64, ok bool) {
	for i := 0; i < len(raw); i++ {
		if raw[i] != ':' {
			continue
		}
		l, err1 := strconv.ParseInt(raw[:i], 10, 64)
		o, err2 := strconv.ParseInt(raw[i+1:], 10, 64)
		if err1 != nil || err2 != nil {
			return 0, 0, false
		}
		return l, o, true
	}
	return 0, 0, false
}

func parseStringFilter(r *http.Request, key string) *string {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	return &raw
}

func parseInt64Filter(r *http.Request, key string) *int64 {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &n
}

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)

	filter := nexusdb.OrderFilter{
		Status:        parseStringFilter(r, "status"),
		Asset:         parseStringFilter(r, "asset"),
		Distributor:   parseStringFilter(r, "distributor"),
		Buyer:         parseStringFilter(r, "buyer"),
		CreatedAfter:  parseInt64Filter(r, "createdAfterLedger"),
		CreatedBefore: parseInt64Filter(r, "createdBeforeLedger"),
		Limit:         limit,
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		ledger, orderID, ok := decodeOrderCursor(raw)
		if !ok {
			writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "The cursor parameter is malformed.")
			return
		}
		filter.CursorLedger = &ledger
		filter.CursorOrderID = &orderID
	}

	orders, err := nexusdb.ListOrders(r.Context(), s.db, filter)
	if err != nil {
		s.logger.Error("list orders failed", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list orders.")
		return
	}

	hasMore := len(orders) > limit
	if hasMore {
		orders = orders[:limit]
	}
	resp := listOrdersResponse{Orders: toOrderResponses(orders)}
	if hasMore && len(orders) > 0 {
		last := orders[len(orders)-1]
		cursor := encodeOrderCursor(last.CreatedAtLedger, last.OrderID)
		resp.NextCursor = &cursor
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	raw := r.PathValue("id")
	orderID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ORDER_ID", "The order id must be an integer.")
		return
	}

	order, err := nexusdb.GetOrder(r.Context(), s.db, orderID)
	if err != nil {
		s.logger.Error("get order failed", "error", err, "orderId", orderID)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load order.")
		return
	}
	if order == nil {
		writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found.")
		return
	}

	writeJSON(w, http.StatusOK, toOrderResponses([]model.Order{*order})[0])
}
