.PHONY: up down restart logs migrate seed server web admin quality quality-fix quality-server quality-web quality-admin fmt fmt-check lint typecheck test

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose down && docker compose up -d

logs:
	docker compose logs -f

migrate:
	cd server && go run cmd/migrate/main.go

seed:
	cd server && go run cmd/seed/main.go

server:
	cd server && go run cmd/server/main.go

web:
	cd web && npm run dev

admin:
	cd admin && npm run dev

quality: quality-server quality-web quality-admin

quality-fix:
	./scripts/quality/server-fix.sh
	cd web && npm run quality:fix
	cd admin && npm run quality:fix
	$(MAKE) quality

quality-server:
	./scripts/quality/server-quality.sh

quality-web:
	cd web && npm run quality

quality-admin:
	cd admin && npm run quality

fmt:
	./scripts/quality/server-fix.sh

fmt-check:
	@unformatted="$$(find server -type f -name '*.go' -print | xargs gofmt -l)"; \
	if [ -n "$$unformatted" ]; then \
		printf '%s\n' "$$unformatted"; \
		exit 1; \
	fi

lint:
	@GOLANGCI_LINT_BIN="$$(command -v golangci-lint 2>/dev/null || true)"; \
	if [ -z "$$GOLANGCI_LINT_BIN" ]; then \
		GOLANGCI_LINT_BIN="$$(go env GOPATH)/bin/golangci-lint"; \
	fi; \
	cd server && "$$GOLANGCI_LINT_BIN" run ./...
	cd web && npm run lint
	cd admin && npm run lint

typecheck:
	cd web && npm run typecheck
	cd admin && npm run typecheck

test:
	cd server && go test ./...
