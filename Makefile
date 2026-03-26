.PHONY: up down restart logs migrate seed server web admin

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
