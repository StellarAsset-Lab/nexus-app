# Nexus

Nexus is an infrastructure layer above Stellar that lets an existing financial application discover, qualify, order, settle, verify, and reconcile Stellar assets through a standardized integration surface.

This repository is the **application layer**: a web application, a typed TypeScript SDK, a Go read API, a Go event indexer, and a Go reconciliation worker, backed by PostgreSQL projections of on-chain state. It is not a wallet, a custody service, an exchange, or a token issuer. See `docs/architecture.md` for the full system boundary.

## Status

This repository is under active initial development. Contract IDs are not yet deployed; see `.env.example` for the required configuration.

## Stack

- Node.js `24.21.0`, pnpm `12.4.2`, Go `1.27.1`, PostgreSQL `18.6`
- Next.js `16.3.5`, React `19.2.8`, TypeScript `6.0.3`, Tailwind CSS `4.3.3`
- `@stellar/stellar-sdk` `17.1.0`, `@creit.tech/stellar-wallets-kit` `2.6.0`
- `github.com/stellar/go-stellar-sdk` `v0.7.2`

> **Go toolchain:** every `go.mod`/`go.work` in this repository declares `go 1.27.1` with an explicit `toolchain go1.27.1` directive. A locally installed Go command older than `1.27.1` will automatically download and use the `go1.27.1` toolchain per [Go's toolchain management](https://go.dev/doc/toolchain) (`GOTOOLCHAIN=auto`, the default) — no manual Go upgrade is required, and CI/Docker images should use `1.27.1` directly.

## Repository layout

```text
apps/web/              Next.js App Router frontend
packages/sdk/           Human-facing TypeScript integration package
packages/contracts/     Generated Soroban contract bindings (Registry, Order)
packages/ui/            Shared visual primitives
services/api/            Go read API over indexed PostgreSQL data
services/indexer/       Go Stellar event ingestion service
services/worker/        Go reconciliation and background processing
internal/               Shared Go packages (db, indexing, httpapi, model, network)
db/migrations/          PostgreSQL schema migrations
scripts/                Contract binding generation and verification scripts
```

## Local development

See `docs/local-development.md` for the full walkthrough. Quick start once dependencies land:

```bash
cp .env.example .env
make install
make db-up
make db-migrate
make web-dev      # in one shell
make api-dev      # in another
make indexer-dev  # in another
make worker-dev   # in another
```

## Documentation

- `docs/architecture.md` — system architecture and data flow
- `docs/api.md` — Go API reference
- `docs/contract-integration.md` — contract ABI and binding workflow
- `docs/local-development.md` — local setup
- `docs/operations.md` — operational runbook
- `docs/testing.md` — test strategy

## License

MIT. See `LICENSE`.
