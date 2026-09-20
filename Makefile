.PHONY: install web-dev api-dev indexer-dev worker-dev db-up db-down db-migrate build test lint check clean bindings verify-contracts

install:
	pnpm install --frozen-lockfile

web-dev:
	pnpm --filter @nexus/web dev

api-dev:
	go run ./services/api/cmd/api

indexer-dev:
	go run ./services/indexer/cmd/indexer

worker-dev:
	go run ./services/worker/cmd/worker

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-migrate:
	go run ./services/api/cmd/api -migrate

build:
	pnpm build
	go build ./...
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go build ./...) || exit 1; \
	done

# `go test ./...` at the repo root covers internal/* (db, httpapi, indexing,
# reconcile) — the per-service loop below only covers each service's own
# cmd/ package, which has no tests of its own. Both are needed.
test:
	pnpm test
	go test ./...
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go test ./...) || exit 1; \
	done

lint:
	pnpm lint
	go vet ./...
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go vet ./...) || exit 1; \
	done

check: lint test build

clean:
	rm -rf apps/web/.next packages/*/dist services/*/bin
	find . -name node_modules -prune -o -name '*.tsbuildinfo' -print0 | xargs -0 rm -f

bindings:
	pnpm bindings

verify-contracts:
	pnpm verify-contracts
