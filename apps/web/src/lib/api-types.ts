// These types mirror the Go API's JSON response shapes exactly (field names
// and optionality), matching internal/httpapi/assets.go, orders.go,
// transactions.go, and status.go. Keep them in lockstep with those handlers
// — this file has no runtime behavior of its own.

export interface AssetResponse {
  readonly asset: string;
  readonly issuer: string;
  readonly active: boolean;
  readonly firstSeenLedger: number;
  readonly lastUpdatedLedger: number;
}

export interface ListAssetsResponse {
  readonly assets: readonly AssetResponse[];
  readonly nextCursor?: string;
}

export type OrderStatus = "Created" | "Settled" | "Cancelled" | "Expired";

export interface OrderSummary {
  readonly orderId: number;
  readonly buyer: string;
  readonly distributor: string;
  readonly asset: string;
  readonly paymentAsset: string;
  readonly assetAmount: string;
  readonly paymentAmount: string;
  readonly createdAtLedger: number;
  readonly expiresAtLedger: number;
  readonly status: OrderStatus;
  readonly paymentFunded: boolean;
  readonly assetFunded: boolean;
  readonly createdTransactionHash: string;
  readonly lastTransactionHash: string;
  readonly lastUpdatedLedger: number;
}

export interface AssetActivityResponse {
  readonly asset: string;
  readonly orders: readonly OrderSummary[];
}

export interface ListOrdersResponse {
  readonly orders: readonly OrderSummary[];
  readonly nextCursor?: string;
}

export interface ListTransactionsResponse {
  readonly transactions: readonly TransactionResponse[];
  readonly nextCursor?: string;
}

export interface TransactionResponse {
  readonly transactionHash: string;
  readonly ledger?: number;
  readonly status: string;
  readonly firstObservedAt: string;
  readonly lastObservedAt: string;
  readonly sourceContract?: string;
  readonly orderId?: number;
}

export interface EventResponse {
  readonly eventId: string;
  readonly ledger: number;
  readonly ledgerClosedAt: string;
  readonly transactionHash: string;
  readonly transactionIndex: number;
  readonly operationIndex: number;
  readonly eventType: string;
  readonly contractId: string;
  readonly decodedPayload?: unknown;
  readonly decodeError?: string;
  readonly firstObservedAt: string;
}

export interface ListEventsResponse {
  readonly events: readonly EventResponse[];
  readonly nextCursor?: string;
}

export type ComponentHealth = "Operational" | "Degraded" | "Unavailable" | "Unknown";

export interface ComponentStatus {
  readonly name: string;
  readonly status: ComponentHealth;
  readonly detail?: string;
}

export interface StatusResponse {
  readonly components: readonly ComponentStatus[];
}

export interface NetworkResponse {
  readonly network: string;
  readonly networkPassphrase: string;
  readonly rpcHealthy: boolean;
  readonly latestLedger?: number;
  readonly oldestLedger?: number;
  readonly indexerCheckpoint?: number;
}
