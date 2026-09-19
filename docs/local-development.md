# Local Development

## Prerequisites

- Node.js `24.21.0`, pnpm `12.4.2`
- Go `1.27.1` (or any Go toolchain — `go.work`'s `toolchain go1.27.1` directive downloads `1.27.1` automatically
  via `GOTOOLCHAIN=auto`)
- Docker (for local PostgreSQL)
- The `stellar` CLI, only if you need to regenerate contract bindings (see `docs/contract-integration.md`)

## Setup

```bash
cp .env.example .env
pnpm check-env          # validates .env against the full required set (scripts/check-env.ts)
make install
make db-up               # starts Postgres via docker-compose, bound to 127.0.0.1 only
make db-migrate           # applies db/migrations/*.sql
```

Then, in separate shells:

```bash
make web-dev      # Next.js dev server
make api-dev       # Go API (reads DATABASE_URL, NEXT_PUBLIC_NETWORK*, API_BIND_ADDR)
make indexer-dev   # Go indexer — no-ops until a Registry or Order contract ID is configured
make worker-dev    # Go reconciliation worker
```

The indexer and worker both need `NEXT_PUBLIC_REGISTRY_CONTRACT_ID`/`NEXT_PUBLIC_ORDER_CONTRACT_ID` to actually do
anything; until a contract is deployed, the indexer logs that it has nothing to index and exits cleanly rather than
polling an unconfigured contract.

## Environment variables

See `.env.example` for the full set. A few notes beyond what the variable names already say:

- `NEXT_PUBLIC_*` variables are the only ones the browser bundle ever sees — the Next.js app never reads a
  server-only variable (`DATABASE_URL`, `API_KEY_PEPPER`) directly; see `apps/web/src/lib/env.ts`.
- `NEXT_PUBLIC_REGISTRY_CONTRACT_ID`/`NEXT_PUBLIC_ORDER_CONTRACT_ID` stay `<SET_AFTER_DEPLOYMENT>` (treated as
  unset) until a real contract is deployed — every part of the stack (indexer, worker, SDK, Sandbox) handles that
  state honestly rather than assuming a contract exists.
- `WORKER_RECONCILE_INTERVAL` is a plain integer number of seconds (not a Go duration string), matching
  `services/worker/cmd/worker/main.go`'s `envDuration` helper.
- `API_KEY_PEPPER` and `INDEXER_BIND_ADDR` are reserved for future work (API key authentication and an indexer
  health endpoint, respectively) — neither is read by any service yet. `pnpm check-env` still requires them so the
  full intended configuration surface is validated up front.

## Database

`make db-up` starts a single Postgres container on `127.0.0.1:5432` (see `docker-compose.yml` — loopback-only, not
exposed on all interfaces). `make db-migrate` runs `go run ./services/api/cmd/api -migrate`, which applies every
migration in `db/migrations/` inside a Postgres advisory lock (`internal/db/migrations.go`) so concurrent
`-migrate` invocations can't race.

To run a service's integration tests, point `DATABASE_URL` at a real Postgres (a throwaway Docker container is
fine) and run `go test ./...` — see `docs/testing.md`.

## Common tasks

| Command | What it does |
| --- | --- |
| `make check` | lint + test + build, exactly what CI runs |
| `make clean` | removes build artifacts (`.next`, `dist`, Go binaries, `.tsbuildinfo`) |
| `pnpm bindings` | regenerates contract bindings from a local `nexus-contract` checkout |
| `pnpm verify-contracts` | verifies the checked-in bindings match `packages/contracts/provenance.json` |

## Troubleshooting

- **"NEXT_PUBLIC_NETWORK must be..."** from a Go service: the Go services validate the same `NEXT_PUBLIC_NETWORK*`
  variables as the frontend (`internal/network`), so they must be set even though the Go services aren't browser
  code — this keeps every part of the stack pinned to one real network's config.
- **Tailwind classes from `packages/ui` not applying**: confirm `apps/web/src/app/globals.css` still has the
  `@source` directive pointing at the sibling workspace package — Tailwind v4 doesn't auto-detect it otherwise.
