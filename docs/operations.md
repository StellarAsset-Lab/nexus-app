# Operations Runbook

## Services

| Service | Purpose | Failure mode if down |
| --- | --- | --- |
| `services/api` | Serves every `GET /api/v1/*` endpoint from indexed Postgres. | Frontend pages fail to load data; `/readyz` returns 503. |
| `services/indexer` | Polls Soroban RPC `getEvents`, persists and projects Registry/Order events. | New on-chain activity stops appearing; `indexerCheckpoint` in `/api/v1/network` stalls. |
| `services/worker` | Reconciles `transactions` against real RPC `getTransaction` calls. | Transaction status stays `Pending` indefinitely even after real on-chain confirmation; visible on the Operations Console's stale-transaction list. |

None of the three holds a wallet key or can sign a transaction — a compromise of any one of them cannot move funds.

## The Operations Console (`/operations`)

Shows, live:

- **Indexer checkpoint** and **latest ledger** (`GET /api/v1/network`) — their difference is **ingest lag**.
  Elevated lag (checkpoint falling behind latest ledger) means the indexer is polling too slowly or has stalled;
  check its logs and confirm it hasn't crash-looped.
- **Stale pending transactions** — transactions still `Pending` more than 5 minutes after first observation
  (`GET /api/v1/transactions?status=Pending`). This means the worker has not yet seen a terminal RPC status for
  them; a large or growing count suggests the worker is down or RPC itself is degraded (check `/api/v1/status`'s
  "Soroban RPC" component first).

## The System Status page (`/status`)

Reports three components, each from a real check made on that request — never a cached or assumed value:

- **Database** — `PingContext` against the configured `DATABASE_URL`.
- **Soroban RPC** — a live `getHealth` call against the configured RPC endpoint.
- **Indexer** — `Degraded` if the checkpoint hasn't advanced in over 5 minutes (`internal/httpapi/status.go`'s
  `indexerLagThreshold`), `Unknown` if no checkpoint has ever been recorded.

## Incident checklist

1. **API returning 500s**: check `/readyz` first — if it's 503, the database is unreachable; check Postgres itself
   before the API's own logs.
2. **Indexer checkpoint not advancing**: check the indexer's logs for repeated `poll iteration failed, retrying with
   backoff` — this means `getEvents` calls are failing (RPC outage, or a misconfigured contract ID). The indexer
   backs off exponentially and retries indefinitely; it does not need to be restarted for a transient RPC outage.
3. **Transactions stuck `Pending`**: confirm the worker process is actually running (its reconciliation pass logs
   `reconciliation pass complete` on every pass that processed at least one hash). If it's running but transactions
   are still stale, check RPC health — the worker maps RPC's `NOT_FOUND` to `Pending` by design (the transaction
   hasn't been ingested by RPC yet), so a real RPC outage looks identical to "worker is stuck" from the outside;
   `/api/v1/status`'s RPC component disambiguates this.
4. **A migration hangs**: `internal/db/migrations.go` serializes concurrent `Migrate()` calls with a Postgres
   advisory lock (key `72_727_272`). If a migration appears stuck, check for another process holding that lock
   (`pg_locks` joined on `pg_advisory_locks`) before assuming the migration itself is broken.

## What this runbook does not cover

Contract-level incidents (a paused Order gateway, a compromised admin key) are Registry/Order contract operations,
not this application's — see `docs/contract-integration.md` for the contracts' own admin/pause methods. This
application's services have no privileged on-chain role and cannot pause, upgrade, or otherwise administer either
contract.
