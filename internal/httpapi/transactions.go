package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	nexusdb "github.com/StellarAsset-Lab/nexus-app/internal/db"
	"github.com/StellarAsset-Lab/nexus-app/internal/model"
)

type transactionResponse struct {
	TransactionHash string    `json:"transactionHash"`
	Ledger          *int64    `json:"ledger,omitempty"`
	Status          string    `json:"status"`
	FirstObservedAt time.Time `json:"firstObservedAt"`
	LastObservedAt  time.Time `json:"lastObservedAt"`
	SourceContract  *string   `json:"sourceContract,omitempty"`
	OrderID         *int64    `json:"orderId,omitempty"`
}

func toTransactionResponse(t model.Transaction) transactionResponse {
	return transactionResponse{
		TransactionHash: t.TransactionHash, Ledger: t.Ledger, Status: t.Status,
		FirstObservedAt: t.FirstObservedAt, LastObservedAt: t.LastObservedAt,
		SourceContract: t.SourceContract, OrderID: t.OrderID,
	}
}

func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue("hash")

	tx, err := nexusdb.GetTransaction(r.Context(), s.db, hash)
	if err != nil {
		s.logger.Error("get transaction failed", "error", err, "transactionHash", hash)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not load transaction.")
		return
	}
	if tx == nil {
		writeError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaction not found.")
		return
	}

	writeJSON(w, http.StatusOK, toTransactionResponse(*tx))
}

// eventResponse never omits DecodedPayload as null vs. absent distinctly
// from DecodeError — both are surfaced exactly as stored, since a nil
// DecodedPayload alongside a non-nil DecodeError is itself meaningful
// (decode failure), not an error in this API layer.
type eventResponse struct {
	EventID          string          `json:"eventId"`
	Ledger           int64           `json:"ledger"`
	LedgerClosedAt   time.Time       `json:"ledgerClosedAt"`
	TransactionHash  string          `json:"transactionHash"`
	TransactionIndex int             `json:"transactionIndex"`
	OperationIndex   int             `json:"operationIndex"`
	EventType        string          `json:"eventType"`
	ContractID       string          `json:"contractId"`
	DecodedPayload   json.RawMessage `json:"decodedPayload,omitempty"`
	DecodeError      *string         `json:"decodeError,omitempty"`
	FirstObservedAt  time.Time       `json:"firstObservedAt"`
}

func toEventResponse(e model.StellarEvent) eventResponse {
	return eventResponse{
		EventID: e.EventID, Ledger: e.Ledger, LedgerClosedAt: e.LedgerClosedAt,
		TransactionHash: e.TransactionHash, TransactionIndex: e.TransactionIndex,
		OperationIndex: e.OperationIndex, EventType: e.EventType, ContractID: e.ContractID,
		DecodedPayload: json.RawMessage(e.DecodedPayload), DecodeError: e.DecodeError,
		FirstObservedAt: e.FirstObservedAt,
	}
}

type listEventsResponse struct {
	Events     []eventResponse `json:"events"`
	NextCursor *string         `json:"nextCursor,omitempty"`
}

func (s *Server) handleListEvents(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)

	filter := nexusdb.EventFilter{
		EventType:  parseStringFilter(r, "eventType"),
		ContractID: parseStringFilter(r, "contractId"),
		Limit:      limit,
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		ledger, eventID, ok := decodeEventCursor(raw)
		if !ok {
			writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "The cursor parameter is malformed.")
			return
		}
		filter.CursorLedger = &ledger
		filter.CursorEventID = &eventID
	}

	events, err := nexusdb.ListEvents(r.Context(), s.db, filter)
	if err != nil {
		s.logger.Error("list events failed", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list events.")
		return
	}

	hasMore := len(events) > limit
	if hasMore {
		events = events[:limit]
	}
	resp := listEventsResponse{Events: make([]eventResponse, 0, len(events))}
	for _, e := range events {
		resp.Events = append(resp.Events, toEventResponse(e))
	}
	if hasMore && len(events) > 0 {
		last := events[len(events)-1]
		cursor := encodeEventCursor(last.Ledger, last.EventID)
		resp.NextCursor = &cursor
	}

	writeJSON(w, http.StatusOK, resp)
}

// encodeEventCursor/decodeEventCursor represent the composite
// (ledger, event_id) keyset cursor. event_id itself never contains ":" (it
// is Stellar's own hex-and-hyphen event id format), so a single split is
// unambiguous.
func encodeEventCursor(ledger int64, eventID string) string {
	return strconv.FormatInt(ledger, 10) + ":" + eventID
}

func decodeEventCursor(raw string) (ledger int64, eventID string, ok bool) {
	for i := 0; i < len(raw); i++ {
		if raw[i] != ':' {
			continue
		}
		l, err := strconv.ParseInt(raw[:i], 10, 64)
		if err != nil || i+1 >= len(raw) {
			return 0, "", false
		}
		return l, raw[i+1:], true
	}
	return 0, "", false
}
