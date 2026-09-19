package model

// Order is the projection of an Order contract OrderRecord plus indexing
// metadata (transaction hashes, last-updated ledger). A terminal status
// (Settled/Cancelled/Expired) must never be moved backward by a later
// duplicate or malformed event — enforced by the indexer's projection logic,
// not by this struct.
type Order struct {
	OrderID                int64       `json:"orderId"`
	Buyer                  string      `json:"buyer"`
	Distributor            string      `json:"distributor"`
	Asset                  string      `json:"asset"`
	PaymentAsset           string      `json:"paymentAsset"`
	AssetAmount            Amount      `json:"assetAmount"`
	PaymentAmount          Amount      `json:"paymentAmount"`
	CreatedAtLedger        int64       `json:"createdAtLedger"`
	ExpiresAtLedger        int64       `json:"expiresAtLedger"`
	Status                 OrderStatus `json:"status"`
	PaymentFunded          bool        `json:"paymentFunded"`
	AssetFunded            bool        `json:"assetFunded"`
	CreatedTransactionHash string      `json:"createdTransactionHash"`
	LastTransactionHash    string      `json:"lastTransactionHash"`
	LastUpdatedLedger      int64       `json:"lastUpdatedLedger"`
}

// OrderEvent links an ordered lifecycle event to its order.
type OrderEvent struct {
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	EventType       string `json:"eventType"`
	EventID         string `json:"eventId"`
	Ledger          int64  `json:"ledger"`
	TransactionHash string `json:"transactionHash"`
	Payload         []byte `json:"payload,omitempty"`
}
