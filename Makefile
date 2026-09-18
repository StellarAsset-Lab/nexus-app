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
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go build ./...) || exit 1; \
	done

test:
	pnpm test
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go test ./...) || exit 1; \
	done

lint:
	pnpm lint
	for service in services/api services/indexer services/worker; do \
		(cd $$service && go vet ./...) || exit 1; \
	done

check: lint test build

clean:
	rm -rf apps/web/.next packages/*/dist services/*/bin
	find . -name '*.tsbuildinfo' -delete

bindings:
	pnpm bindings

verify-contracts:
	pnpm verify-contracts
