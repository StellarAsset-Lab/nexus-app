# API Reference

The Go API (`services/api`, handlers in `internal/httpapi`) is a **read-only** HTTP service over indexed
PostgreSQL projections of on-chain state. It never signs or submits a transaction, and holds no wallet key. The
same content is also rendered at `/developers/api` in the web app; this file is the source-controlled copy for
readers outside the running app.

All responses are JSON. Every error uses the stable shape:

```json
{ "error": { "code": "ASSET_NOT_FOUND", "message": "Asset not found." } }
```

## Health

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/healthz` | Liveness probe — always 200 if the process is running. |
| GET | `/readyz` | Readiness probe — checks the database connection. `503 NOT_READY` if unreachable. |
| GET | `/api/v1/network` | Current network config, live RPC health (`getHealth`), and the indexer's checkpoint. |
| GET | `/api/v1/status` | Directly-checked health of every dependency (database, RPC, indexer lag). |

## Assets

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/assets` | List assets, keyset-paginated by `asset` ascending. |
| GET | `/api/v1/assets/{asset}` | A single asset. `404 ASSET_NOT_FOUND` if never indexed. |
| GET | `/api/v1/assets/{asset}/activity` | Recent orders referencing this asset. |

`GET /api/v1/assets` query parameters:

- `active` — `true`/`false`. Omit for both active and inactive assets.
- `cursor` — the previous page's `nextCursor`.
- `limit` — page size, default 50, max 200.

## Orders

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/orders` | List orders, keyset-paginated by `(createdAtLedger, orderId)` descending. |
| GET | `/api/v1/orders/{id}` | A single order by its numeric ID. `400 INVALID_ORDER_ID` / `404 ORDER_NOT_FOUND`. |

`GET /api/v1/orders` query parameters: `status` (`Created`/`Settled`/`Cancelled`/`Expired`), `asset`, `distributor`,
`buyer`, `createdAfterLedger`, `createdBeforeLedger`, `cursor`, `limit`.

## Transactions and events

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/api/v1/transactions` | List transactions the worker has reconciled, newest first. |
| GET | `/api/v1/transactions/{hash}` | A single transaction's real, reconciled status. `404 TRANSACTION_NOT_FOUND`. |
| GET | `/api/v1/events` | List raw indexed contract events, keyset-paginated by `(ledger, eventId)` ascending. |

`GET /api/v1/transactions` query parameters: `status` (`Pending`/`Confirmed`/`Failed`), `cursor`, `limit`.

`GET /api/v1/events` query parameters: `eventType` (see `docs/contract-integration.md` for the full list),
`contractId`, `transactionHash`, `cursor`, `limit`.

## Pagination

Every list endpoint uses **keyset pagination**: `cursor` echoes an opaque value from the previous page's
`nextCursor`, never an offset. This is deterministic under concurrent writes — an `OFFSET`-based page can skip or
repeat rows when new rows are inserted between requests; a keyset cursor cannot (see `internal/httpapi/pagination.go`
and each list query in `internal/db/queries.go`).

## SDK error taxonomy

The TypeScript SDK (`@nexus/sdk`) throws typed errors distinct from the API's HTTP error codes above — see
`packages/sdk/src/errors.ts`: `ConfigurationError`, `ValidationError`, `RpcError`, `SimulationError`,
`WalletRejectedError`, `SubmissionError`, `ConfirmationTimeoutError`, `ContractError`, `IndexerLagError`. Each
extends `NexusError` with a stable `code` property.
