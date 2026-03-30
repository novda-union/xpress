#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/xpressgo}"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.yml}"
ENV_FILE="${ENV_FILE:-infra/deploy/vps.env}"
HEALTH_URL="${HEALTH_URL:-https://srvr.novdaunion.uz/health}"

if [[ ! -d "$APP_DIR/.git" ]]; then
  echo "expected a git checkout at $APP_DIR" >&2
  exit 1
fi

if [[ ! -f "$APP_DIR/$ENV_FILE" ]]; then
  echo "missing runtime env file: $APP_DIR/$ENV_FILE" >&2
  echo "copy infra/deploy/vps.env.example to that path before deploying" >&2
  exit 1
fi

cd "$APP_DIR"

set -a
. "$ENV_FILE"
set +a

git fetch origin main
git reset --hard origin/main

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build
go run server/cmd/migrate/main.go

curl --fail --silent --show-error "$HEALTH_URL" >/dev/null
curl --fail --silent --show-error "https://customer.novdaunion.uz/" >/dev/null
curl --fail --silent --show-error "https://admin.novdaunion.uz/" >/dev/null

echo "deployment completed successfully"
