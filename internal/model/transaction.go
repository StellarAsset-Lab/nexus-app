package model

import "time"

// Transaction is the indexed/application observation of a real Stellar
// transaction. Ledger and SourceContract are nil until the transaction is
// actually confirmed on-chain — never inferred.
type Transaction struct {
	TransactionHash string    `json:"transactionHash"`
	Ledger          *int64    `json:"ledger,omitempty"`
	Status          string    `json:"status"`
	FirstObservedAt time.Time `json:"firstObservedAt"`
	LastObservedAt  time.Time `json:"lastObservedAt"`
	SourceContract  *string   `json:"sourceContract,omitempty"`
	OrderID         *int64    `json:"orderId,omitempty"`
}

// Checkpoint is the durable indexer resume point.
type Checkpoint struct {
	ID         string    `json:"id"`
	LastLedger int64     `json:"lastLedger"`
	Cursor     *string   `json:"cursor,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
