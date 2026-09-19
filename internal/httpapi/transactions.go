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

type listTransactionsResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	NextCursor   *string               `json:"nextCursor,omitempty"`
}

func (s *Server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r)

	filter := nexusdb.TransactionFilter{
		Status: parseStringFilter(r, "status"),
		Limit:  limit,
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		observedAt, hash, ok := decodeTransactionCursor(raw)
		if !ok {
			writeError(w, http.StatusBadRequest, "INVALID_CURSOR", "The cursor parameter is malformed.")
			return
		}
		filter.CursorObservedAt = &observedAt
		filter.CursorHash = &hash
	}

	txs, err := nexusdb.ListTransactions(r.Context(), s.db, filter)
	if err != nil {
		s.logger.Error("list transactions failed", "error", err)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not list transactions.")
		return
	}

	hasMore := len(txs) > limit
	if hasMore {
		txs = txs[:limit]
	}
	resp := listTransactionsResponse{Transactions: make([]transactionResponse, 0, len(txs))}
	for _, t := range txs {
		resp.Transactions = append(resp.Transactions, toTransactionResponse(t))
	}
	if hasMore && len(txs) > 0 {
		last := txs[len(txs)-1]
		cursor := encodeTransactionCursor(last.LastObservedAt, last.TransactionHash)
		resp.NextCursor = &cursor
	}

	writeJSON(w, http.StatusOK, resp)
}

// encodeTransactionCursor/decodeTransactionCursor represent the composite
// (last_observed_at, transaction_hash) keyset cursor as an RFC 3339
// timestamp, ":", then the hash — a transaction hash never contains ":".
func encodeTransactionCursor(observedAt time.Time, hash string) string {
	return observedAt.Format(time.RFC3339Nano) + ":" + hash
}

// The RFC3339Nano timestamp itself contains colons (HH:MM:SS), so the split
// point is the LAST ':' in the cursor — transaction hashes are hex-only and
// never contain one.
func decodeTransactionCursor(raw string) (observedAt time.Time, hash string, ok bool) {
	lastColon := -1
	for i := len(raw) - 1; i >= 0; i-- {
		if raw[i] == ':' {
			lastColon = i
			break
		}
	}
	if lastColon <= 0 || lastColon+1 >= len(raw) {
		return time.Time{}, "", false
	}
	t, err := time.Parse(time.RFC3339Nano, raw[:lastColon])
	if err != nil {
		return time.Time{}, "", false
	}
	return t, raw[lastColon+1:], true
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
		EventType:       parseStringFilter(r, "eventType"),
		ContractID:      parseStringFilter(r, "contractId"),
		TransactionHash: parseStringFilter(r, "transactionHash"),
		Limit:           limit,
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
