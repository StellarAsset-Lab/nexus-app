# Architecture

Nexus is the **application layer** above two real Soroban contracts (Registry, Order). It is not a wallet, custody
service, exchange, or token issuer — it never holds funds or private keys.

## System boundary

```text
                         ┌─────────────────────┐
                         │   Stellar Testnet    │
                         │  Registry / Order     │
                         │      contracts        │
                         └─────────┬────────────┘
                                   │ Soroban RPC (getEvents, getTransaction)
                    ┌──────────────┼───────────────┐
                    │              │               │
              ┌─────▼─────┐  ┌────▼─────┐   ┌──────▼──────┐
              │  indexer   │  │  worker   │   │  web app     │
              │ (poll+     │  │ (reconcile│   │ (browser,    │
              │  project)  │  │  txns)    │   │  wallet)     │
              └─────┬─────┘  └────┬─────┘   └──────┬──────┘
                    │              │               │ signs & submits
                    │              │               │ transactions directly
              ┌─────▼──────────────▼─────┐         │ via the connected wallet
              │       PostgreSQL           │         │ (never through the API)
              │  (stellar_events, assets,  │         │
              │   orders, transactions)    │         │
              └─────────────┬─────────────┘         │
                             │ read-only              │
                       ┌─────▼─────┐                 │
                       │    api     │◄────────────────┘ reads only
                       └───────────┘
```

- **indexer** polls Soroban RPC's `getEvents` for the Registry and Order contracts, persists the raw event, then
  projects it into typed tables (`assets`, `distributions`, `eligibilities`, `orders`, `order_events`). See
  `internal/indexing`.
- **worker** reconciles the `transactions` table against real RPC `getTransaction` calls for every transaction hash
  the indexer has observed, since the indexer only processes emitted events, not submission-time transaction status.
  See `internal/reconcile`.
- **api** is a read-only Go HTTP service over the indexed PostgreSQL projections (`internal/httpapi`). It never signs
  or submits a transaction, and never holds a wallet key.
- **web app** (Next.js) reads indexed state through the API, and — for the Sandbox and any future write flow —
  builds and simulates transactions via the generated contract clients, then hands them to a wallet connected
  through Stellar Wallets Kit for signing. Signing and submission happen entirely client-side, in the user's own
  wallet; the API and backend services never see a private key.

## Why index at all?

Soroban RPC only retains recent ledger history and does not support the query patterns (pagination, filtering,
joins across an asset's orders) the frontend needs. The indexer/PostgreSQL layer is a durable, queryable projection
of on-chain state — never a source of truth that diverges from it. Every value in an indexed table is either a
direct on-chain observation (an event's decoded fields) or a directly-observed RPC status (the worker's
reconciliation) — nothing is estimated, defaulted, or fabricated.

## Data flow for one order

1. A distributor or buyer signs `create_order` in their wallet; the Sandbox (or any integrating app) submits it via
   the SDK's `executeTransaction`.
2. The Order contract emits `order_created`; the indexer's poller sees it via `getEvents`, inserts the raw
   `stellar_events` row, then projects it into `orders`/`order_events`.
3. The worker independently reconciles the transaction hash against `getTransaction`, recording its real terminal
   status (`Confirmed`/`Failed`) in `transactions` — this is how the Transaction Center and Operations Console know
   a transaction actually landed, since events alone don't carry submission-level success/failure detail.
4. Later lifecycle events (`payment_funded`, `asset_funded`, `order_settled`, `order_cancelled`, `order_expired`)
   follow the same event → raw row → projection path, with a terminal-status guard
   (`internal/indexing/order_events.go`) so a late or out-of-order event can never move a `Settled`/`Cancelled`/
   `Expired` order backward.
5. The API and every frontend page read only from these projections — `GET /api/v1/orders/{id}` never queries RPC
   directly.

## Exact amounts

Every token amount (`asset_amount`, `payment_amount`) is a Soroban `i128`. The Go side stores it as
`NUMERIC(39,0)` and wraps it in `model.Amount` (backed by `math/big.Int`), serialized as a JSON string — never a
JSON number, which would silently lose precision above 2^53. The TypeScript side represents the same value as a
native `bigint`, which the generated contract bindings already produce for `i128` fields. Neither layer ever
represents an amount as a floating-point number.

## Repository layout

See `README.md` for the top-level layout. Go services share code via `internal/` (a Go workspace with per-service
`go.mod` files using `replace` directives), not via a shared library published elsewhere — this keeps `internal/`
genuinely internal to this repository, as the Go compiler enforces.
