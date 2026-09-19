# Testing Strategy

## Principle: prefer the real thing over a mock

Every integration test in this repository runs against a **real PostgreSQL** database and, where relevant, **real
Stellar Testnet RPC** — never a mocked database driver or a stubbed RPC client. This caught two genuinely
production-breaking bugs during development that a mock would have hidden:

- Soroban's `#[contractevent]` macro always encodes event data as a `Map`, even for a single field — an assumption
  that a 1-field event uses a bare value instead would have silently mis-decoded real events. Verified against the
  actual compiled ABI, not just documentation.
- The original indexer wrote `order_events` rows before their corresponding `stellar_events` row existed, which
  would violate `order_events.event_id`'s foreign key in production. A mocked `*sql.DB` that doesn't enforce
  constraints would have let this pass.

## Go tests

Integration tests skip themselves (`t.Skip(...)`) when `DATABASE_URL` is unset, so `go test ./...` still runs
cleanly with no database available, but every meaningful assertion lives in the skippable integration tests, not in
mock-backed unit tests. To run them:

```bash
docker run -d --name nexus-pg-test -p 127.0.0.1:<port>:5432 \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=nexus postgres:18.6
DATABASE_URL="postgres://postgres:postgres@127.0.0.1:<port>/nexus?sslmode=disable" go test ./... -count=1
```

Tests that exercise the indexer or the reconciliation worker also make real calls to
`https://soroban-testnet.stellar.org` (e.g. `TestIntegrationPollerRealRPC`, `TestIntegrationReconcileRealRPC`) —
they use syntactically valid but certainly-nonexistent Testnet addresses/hashes (via `strkey.Encode` /
`crypto/rand`) so they don't depend on any specific real on-chain state existing, while still proving the real RPC
round trip works.

Each Go package's tests cover:

- `internal/db` — migration idempotency and the Postgres advisory lock serializing concurrent `Migrate()` calls.
- `internal/httpapi` — every list/get endpoint against seeded fixture rows: pagination correctness (every fixture
  row seen exactly once), filters, and the stable 404/400 error shapes.
- `internal/indexing` — event decoding against real compiled event ABI shapes, and the full
  decode → raw-insert → project pipeline against a real database, including the terminal-status guard that stops a
  late `Settled`/`Cancelled`/`Expired` event from moving an order backward.
- `internal/reconcile` — the pending-hash discovery query and a real `getTransaction` round trip.

## TypeScript checks

`packages/sdk`, `packages/ui`, and `packages/contracts/*` are type-checked (`tsc`) as part of `pnpm build`; `apps/web`
adds `next build`'s own type check and static generation for every route. There is currently no dedicated Vitest/Jest
suite for the SDK's pure logic (validation, cursor decoding, error mapping) — `tsc` plus the Go-side integration
tests covering the same data shapes are the current coverage. Adding SDK-level unit tests for the write-helper
validation (`buildCreateOrder` et al.) and the wallet transaction lifecycle state machine
(`packages/sdk/src/transactions.ts`) is tracked as follow-up work, not yet done.

## What is not tested against a live wallet

The Sandbox's actual signing flow (`apps/web/src/hooks/use-transaction-flow.ts`) has not been exercised end-to-end
with a real browser wallet extension in this repository's test suite — doing so would require a funded Testnet
account and a real wallet, which is out of scope for an automated CI run. It has been verified by direct code
review against the SDK's own `executeTransaction` implementation and the generated contract client's real method
signatures, and the "Order contract not yet deployed" path has been smoke-tested (see the commit history for the
Sandbox pages).

## `make check`

`make check` runs `lint`, `test`, and `build` for every workspace (`pnpm -r`) and every Go service, and is exactly
what should pass before any commit lands — see `docs/local-development.md` for the full command set.
