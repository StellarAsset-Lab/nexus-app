package model

import "time"

// StellarEvent is the raw, normalized projection of one Nexus contract
// event. The raw XDR is always retained, even when DecodedPayload is nil
// because decoding failed — see spec §16.3.
type StellarEvent struct {
	EventID          string    `json:"eventId"`
	Ledger           int64     `json:"ledger"`
	LedgerClosedAt   time.Time `json:"ledgerClosedAt"`
	TransactionHash  string    `json:"transactionHash"`
	TransactionIndex int       `json:"transactionIndex"`
	OperationIndex   int       `json:"operationIndex"`
	EventType        string    `json:"eventType"`
	ContractID       string    `json:"contractId"`
	TopicsXDR        string    `json:"topicsXdr"`
	ValueXDR         string    `json:"valueXdr"`
	DecodedPayload   []byte    `json:"decodedPayload,omitempty"`
	DecodeError      *string   `json:"decodeError,omitempty"`
	FirstObservedAt  time.Time `json:"firstObservedAt"`
}
