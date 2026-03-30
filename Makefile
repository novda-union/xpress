.PHONY: up down restart fresh logs wait-db migrate seed server web admin docs-check docs-refresh quality quality-fix quality-server quality-web quality-admin fmt fmt-check lint typecheck test ensure-local-https print-lan-ip

up:
	$(MAKE) ensure-local-https
	docker compose up -d
	$(MAKE) print-lan-ip

down:
	docker compose down

restart:
	$(MAKE) ensure-local-https
	docker compose down && docker compose up -d
	$(MAKE) print-lan-ip

fresh:
	docker compose down -v --remove-orphans
	rm -rf web/dist admin/.nuxt admin/.output server/tmp server/bin
	$(MAKE) ensure-local-https
	docker compose up -d --build
	$(MAKE) migrate
	$(MAKE) seed
	$(MAKE) print-lan-ip

ensure-local-https:
	bash scripts/local_https/ensure_certs.sh

print-lan-ip:
	bash scripts/local_https/print_lan_ip.sh

wait-db:
	@printf '%s' 'Waiting for postgres'
	@until docker compose exec -T postgres pg_isready -U xpressgo -d xpressgo >/dev/null 2>&1; do \
		printf '%s' '.'; \
		sleep 1; \
	done; \
	printf '\n'

logs:
	docker compose logs -f

migrate:
	$(MAKE) wait-db
	cd server && go run cmd/migrate/main.go

seed:
	$(MAKE) wait-db
	cd server && go run cmd/seed/main.go

server:
	cd server && go run cmd/server/main.go

web:
	cd web && npm run dev

admin:
	cd admin && npm run dev

docs-check:
	python3 scripts/docs/docs_check.py

docs-refresh:
	python3 scripts/docs/docs_refresh.py

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
