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

git fetch origin master
git reset --hard origin/master

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T server ./migrate

echo "waiting for server to be ready..."
for i in $(seq 1 30); do
  if curl --fail --silent --show-error "$HEALTH_URL" >/dev/null 2>&1; then
    echo "server is ready"
    break
  fi
  [ "$i" -eq 30 ] && { echo "server did not become ready in time"; exit 1; }
  sleep 2
done
echo "waiting for customer app to be ready..."
for i in $(seq 1 30); do
  if curl --fail --silent --show-error "https://customer.novdaunion.uz/" >/dev/null 2>&1; then
    echo "customer app is ready"
    break
  fi
  [ "$i" -eq 30 ] && { echo "customer app did not become ready in time"; exit 1; }
  sleep 2
done

echo "waiting for admin app to be ready..."
for i in $(seq 1 30); do
  if curl --fail --silent --show-error "https://admin.novdaunion.uz/" >/dev/null 2>&1; then
    echo "admin app is ready"
    break
  fi
  [ "$i" -eq 30 ] && { echo "admin app did not become ready in time"; exit 1; }
  sleep 2
done

echo "deployment completed successfully"
