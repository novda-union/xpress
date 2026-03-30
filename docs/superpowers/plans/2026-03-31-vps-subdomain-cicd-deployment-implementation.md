# VPS Subdomain CI/CD Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the temporary local-HTTPS workflow with a VPS-hosted deployment on `customer.novdaunion.uz`, `admin.novdaunion.uz`, and `srvr.novdaunion.uz`, enforced by zero-error CI gates and automatic deploys on every push to `main`.

**Architecture:** Retire the laptop-only HTTPS proxy path and move to a production-style split-host setup. Keep Docker Compose for app services on the VPS, use host-level Nginx for TLS and hostname routing, point both frontends to `srvr.novdaunion.uz`, and gate deploys behind `make quality`, `make test`, Docker Compose validation, and Nginx validation in GitHub Actions.

**Tech Stack:** Docker Compose, Ubuntu Nginx, Let's Encrypt, GitHub Actions, Go, React/Vite, Nuxt 3, PostgreSQL

---

## File Structure

### Runtime And Config

- Modify: `docker-compose.yml`
  - remove the temporary local HTTPS proxy service and shift to VPS-oriented service definitions
- Modify: `Makefile`
  - remove local-HTTPS automation targets and add deployment-safe validation targets
- Modify: `server/internal/config/config.go`
  - replace the temporary local hostname fallback with production-oriented environment-driven values
- Modify: `server/cmd/server/main.go`
  - restore explicit CORS rules for the two frontend production origins
- Modify: `web/src/lib/api.ts`
  - make the customer app explicitly consume `https://srvr.novdaunion.uz` via env-driven configuration
- Modify: `web/vite.config.ts`
  - remove temporary local same-origin proxy assumptions that only served the local HTTPS workaround
- Modify: `admin/nuxt.config.ts`
  - switch admin API base config to env-driven production-friendly values

### Infrastructure

- Create: `infra/nginx/customer.novdaunion.uz.conf`
  - host-level Nginx config for the customer frontend
- Create: `infra/nginx/admin.novdaunion.uz.conf`
  - host-level Nginx config for the admin frontend
- Create: `infra/nginx/srvr.novdaunion.uz.conf`
  - host-level Nginx config for the backend API and WebSocket origin
- Create: `infra/deploy/vps.env.example`
  - example runtime environment file for the VPS
- Create: `infra/deploy/deploy.sh`
  - VPS-side deployment entrypoint for pulls, rebuilds, migrations, and restarts

### CI/CD

- Create: `.github/workflows/ci-cd.yml`
  - quality gates plus deploy-on-main automation

### Cleanup Of Temporary Local HTTPS Workflow

- Delete: `infra/local-https/nginx.conf`
- Delete: `infra/local-https/openssl.cnf`
- Delete: `scripts/local_https/ensure_certs.sh`
- Delete: `scripts/local_https/print_lan_ip.sh`
- Modify: `.gitignore`
  - remove temporary local HTTPS runtime ignore entries if no longer needed

### Documentation

- Modify: `README.md`
  - remove temporary local HTTPS instructions and replace with VPS deploy workflow
- Modify: `AGENTS.md`
  - replace temporary local HTTPS runtime notes with VPS deployment/runtime notes
- Delete: `docs/superpowers/specs/2026-03-31-local-https-telegram-miniapp-design.md`
- Delete: `docs/superpowers/plans/2026-03-31-local-https-telegram-miniapp-implementation.md`
- Modify: `docs/superpowers/specs/2026-03-31-vps-subdomain-cicd-deployment-design.md`
  - update if implementation details sharpen during rollout

---

### Task 1: Remove The Temporary Local HTTPS Workflow

**Files:**
- Delete: `infra/local-https/nginx.conf`
- Delete: `infra/local-https/openssl.cnf`
- Delete: `scripts/local_https/ensure_certs.sh`
- Delete: `scripts/local_https/print_lan_ip.sh`
- Modify: `docker-compose.yml`
- Modify: `Makefile`
- Modify: `.gitignore`

- [ ] **Step 1: Remove the local HTTPS proxy service from Docker Compose**

Update `docker-compose.yml` to remove the `proxy` service and stop hardcoding `xpressgo.home.arpa`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: xpressgo
      POSTGRES_USER: xpressgo
      POSTGRES_PASSWORD: xpressgo
    ports:
      - "5433:5432"
    volumes:
      - postgres-data:/var/lib/postgresql/data

  server:
    build: ./server
    depends_on:
      - postgres
    environment:
      DATABASE_URL: ${DATABASE_URL:-postgres://xpressgo:xpressgo@postgres:5432/xpressgo?sslmode=disable}
      TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN:-}
      TELEGRAM_GATEWAY_TOKEN: ${TELEGRAM_GATEWAY_TOKEN:-}
      JWT_SECRET: ${JWT_SECRET:-dev-secret-change-me}
      APP_URL: ${APP_URL:-https://customer.novdaunion.uz}
    ports:
      - "8080:8080"

  web:
    build: ./web
    ports:
      - "5173:5173"
    volumes:
      - ./web/src:/app/src

  admin:
    build: ./admin
    ports:
      - "3000:3000"
    volumes:
      - ./admin/pages:/app/pages
      - ./admin/components:/app/components
      - ./admin/composables:/app/composables
      - ./admin/layouts:/app/layouts

volumes:
  postgres-data:
```

- [ ] **Step 2: Remove local HTTPS make targets and restore clean local runtime targets**

Update `Makefile` to remove `ensure-local-https` and `print-lan-ip` from the workflow:

```make
.PHONY: up down restart fresh logs wait-db migrate seed server web admin docs-check docs-refresh quality quality-fix quality-server quality-web quality-admin fmt fmt-check lint typecheck test validate-compose validate-nginx

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose down && docker compose up -d

fresh:
	docker compose down -v --remove-orphans
	rm -rf web/dist admin/.nuxt admin/.output server/tmp server/bin
	docker compose up -d --build
	$(MAKE) migrate
	$(MAKE) seed

validate-compose:
	docker compose config >/dev/null

validate-nginx:
	nginx -t -c $(PWD)/infra/nginx/nginx.test.conf -p $(PWD)
```

- [ ] **Step 3: Remove local HTTPS runtime ignore**

Update `.gitignore` by removing:

```gitignore
# Local HTTPS runtime
.local-certs/
```

- [ ] **Step 4: Delete temporary local HTTPS assets**

Delete these files:

```text
infra/local-https/nginx.conf
infra/local-https/openssl.cnf
scripts/local_https/ensure_certs.sh
scripts/local_https/print_lan_ip.sh
```

- [ ] **Step 5: Validate the cleanup diff**

Run: `git diff --check`

Expected: no output

- [ ] **Step 6: Commit**

```bash
git add docker-compose.yml Makefile .gitignore
git add -u infra/local-https scripts/local_https
git commit -m "refactor(runtime): remove temporary local https workflow"
```

### Task 2: Switch Runtime Configuration To VPS Public Domains

**Files:**
- Modify: `server/internal/config/config.go`
- Modify: `server/cmd/server/main.go`
- Modify: `web/src/lib/api.ts`
- Modify: `web/vite.config.ts`
- Modify: `admin/nuxt.config.ts`

- [ ] **Step 1: Change server config defaults to public deployment values**

Update `server/internal/config/config.go` so the fallback app URL reflects the customer domain:

```go
AppURL: getEnv("APP_URL", "https://customer.novdaunion.uz"),
```

- [ ] **Step 2: Restore explicit CORS for the two frontend origins**

Update `server/cmd/server/main.go` to use explicit frontend origins instead of the prototype-wide allow-all:

```go
e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
	AllowOrigins: []string{
		"https://customer.novdaunion.uz",
		"https://admin.novdaunion.uz",
	},
	AllowMethods: []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
	},
	AllowHeaders: []string{
		echo.HeaderOrigin,
		echo.HeaderContentType,
		echo.HeaderAccept,
		echo.HeaderAuthorization,
	},
}))
```

- [ ] **Step 3: Make the web app consume an explicit backend origin**

Update `web/src/lib/api.ts` so the default API base is the server hostname instead of same-origin fallback:

```ts
const API_BASE = import.meta.env.VITE_API_BASE_URL?.trim() ?? 'https://srvr.novdaunion.uz'
```

Keep `resolveApiUrl()` and `getWsUrl()` aligned to that base so WebSocket traffic becomes:

```ts
const url = new URL('/ws', API_BASE)
url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
```

- [ ] **Step 4: Remove temporary dev-proxy assumptions from the web Vite config**

Update `web/vite.config.ts` so local dev can still use an override, but production-oriented defaults are explicit:

```ts
const env = loadEnv(mode, process.cwd(), 'VITE_')
const proxyTarget = env.VITE_PROXY_TARGET || 'http://localhost:8080'
```

Keep the local proxy only for host-machine dev convenience, not as part of the production deployment story.

- [ ] **Step 5: Make the admin app API base environment-driven**

Update `admin/nuxt.config.ts`:

```ts
runtimeConfig: {
  public: {
    apiBase: process.env.NUXT_PUBLIC_API_BASE || 'https://srvr.novdaunion.uz',
  },
},
```

- [ ] **Step 6: Run focused checks**

Run:

```bash
cd server && go test ./...
npm --prefix web run lint
npm --prefix web run typecheck
npm --prefix admin run lint
npm --prefix admin run typecheck
```

Expected: all pass, or only pre-existing known warnings remain in admin lint

- [ ] **Step 7: Commit**

```bash
git add server/internal/config/config.go server/cmd/server/main.go web/src/lib/api.ts web/vite.config.ts admin/nuxt.config.ts
git commit -m "feat(runtime): target public vps subdomains"
```

### Task 3: Add VPS Nginx Configuration

**Files:**
- Create: `infra/nginx/customer.novdaunion.uz.conf`
- Create: `infra/nginx/admin.novdaunion.uz.conf`
- Create: `infra/nginx/srvr.novdaunion.uz.conf`
- Create: `infra/nginx/nginx.test.conf`

- [ ] **Step 1: Add customer hostname config**

Create `infra/nginx/customer.novdaunion.uz.conf`:

```nginx
server {
    listen 80;
    server_name customer.novdaunion.uz;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name customer.novdaunion.uz;

    ssl_certificate /etc/letsencrypt/live/customer.novdaunion.uz/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/customer.novdaunion.uz/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:5173;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```

- [ ] **Step 2: Add admin hostname config**

Create `infra/nginx/admin.novdaunion.uz.conf`:

```nginx
server {
    listen 80;
    server_name admin.novdaunion.uz;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name admin.novdaunion.uz;

    ssl_certificate /etc/letsencrypt/live/admin.novdaunion.uz/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.novdaunion.uz/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
    }
}
```

- [ ] **Step 3: Add backend hostname config**

Create `infra/nginx/srvr.novdaunion.uz.conf`:

```nginx
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}

server {
    listen 80;
    server_name srvr.novdaunion.uz;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name srvr.novdaunion.uz;

    ssl_certificate /etc/letsencrypt/live/srvr.novdaunion.uz/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/srvr.novdaunion.uz/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
    }
}
```

- [ ] **Step 4: Add a local test harness Nginx config for CI validation**

Create `infra/nginx/nginx.test.conf`:

```nginx
events {}

http {
    include /etc/nginx/mime.types;
    map $http_upgrade $connection_upgrade {
        default upgrade;
        '' close;
    }

    include infra/nginx/customer.novdaunion.uz.conf;
    include infra/nginx/admin.novdaunion.uz.conf;
    include infra/nginx/srvr.novdaunion.uz.conf;
}
```

- [ ] **Step 5: Validate the Nginx config**

Run: `nginx -t -c $(pwd)/infra/nginx/nginx.test.conf -p $(pwd)`

Expected: `syntax is ok` and `test is successful`

- [ ] **Step 6: Commit**

```bash
git add infra/nginx/customer.novdaunion.uz.conf infra/nginx/admin.novdaunion.uz.conf infra/nginx/srvr.novdaunion.uz.conf infra/nginx/nginx.test.conf
git commit -m "feat(infra): add vps nginx host configs"
```

### Task 4: Add VPS Deploy Scripts And Runtime Templates

**Files:**
- Create: `infra/deploy/vps.env.example`
- Create: `infra/deploy/deploy.sh`

- [ ] **Step 1: Add the VPS env template**

Create `infra/deploy/vps.env.example`:

```dotenv
DATABASE_URL=postgres://xpressgo:xpressgo@postgres:5432/xpressgo?sslmode=disable
TELEGRAM_BOT_TOKEN=
TELEGRAM_GATEWAY_TOKEN=
JWT_SECRET=
APP_URL=https://customer.novdaunion.uz
VITE_API_BASE_URL=https://srvr.novdaunion.uz
NUXT_PUBLIC_API_BASE=https://srvr.novdaunion.uz
```

- [ ] **Step 2: Add the VPS deploy script**

Create `infra/deploy/deploy.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

APP_DIR="${APP_DIR:-/opt/xpressgo}"

cd "$APP_DIR"

git fetch origin
git checkout main
git reset --hard origin/main

docker compose up -d --build
make migrate
docker compose ps
curl --fail --silent https://srvr.novdaunion.uz/health >/dev/null
```

- [ ] **Step 3: Make the deploy script executable**

Run: `chmod +x infra/deploy/deploy.sh`

Expected: no output

- [ ] **Step 4: Commit**

```bash
git add infra/deploy/vps.env.example infra/deploy/deploy.sh
git commit -m "feat(deploy): add vps deployment scripts"
```

### Task 5: Add GitHub Actions Quality And Deploy Pipeline

**Files:**
- Create: `.github/workflows/ci-cd.yml`

- [ ] **Step 1: Add the CI quality gate workflow**

Create `.github/workflows/ci-cd.yml`:

```yaml
name: CI/CD

on:
  push:
    branches:
      - main

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: server/go.mod

      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: npm
          cache-dependency-path: |
            web/package-lock.json
            admin/package-lock.json

      - name: Install nginx
        run: sudo apt-get update && sudo apt-get install -y nginx

      - name: Install web dependencies
        run: npm --prefix web ci

      - name: Install admin dependencies
        run: npm --prefix admin ci

      - name: Run quality
        run: make quality

      - name: Run server tests
        run: make test

      - name: Validate compose
        run: make validate-compose

      - name: Validate nginx
        run: make validate-nginx

  deploy:
    runs-on: ubuntu-latest
    needs: quality
    steps:
      - name: Configure SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.DEPLOY_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -p "${{ secrets.DEPLOY_PORT }}" "${{ secrets.DEPLOY_HOST }}" >> ~/.ssh/known_hosts

      - name: Deploy
        run: |
          ssh -p "${{ secrets.DEPLOY_PORT }}" "${{ secrets.DEPLOY_USER }}@${{ secrets.DEPLOY_HOST }}" \
            'APP_DIR=/opt/xpressgo bash /opt/xpressgo/infra/deploy/deploy.sh'
```

- [ ] **Step 2: Validate the workflow YAML**

Run: `python3 - <<'PY'\nimport yaml, pathlib\nprint(yaml.safe_load(pathlib.Path('.github/workflows/ci-cd.yml').read_text())['name'])\nPY`

Expected: `CI/CD`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci-cd.yml
git commit -m "feat(ci): add main-branch deploy pipeline"
```

### Task 6: Replace The Temporary Local HTTPS Documentation

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `docs/superpowers/specs/2026-03-31-vps-subdomain-cicd-deployment-design.md`
- Delete: `docs/superpowers/specs/2026-03-31-local-https-telegram-miniapp-design.md`
- Delete: `docs/superpowers/plans/2026-03-31-local-https-telegram-miniapp-implementation.md`

- [ ] **Step 1: Rewrite README runtime instructions for the VPS deployment model**

Replace temporary local HTTPS sections in `README.md` with:

```md
## Deployment Hosts

- `customer.novdaunion.uz` - customer mini app frontend
- `admin.novdaunion.uz` - admin panel frontend
- `srvr.novdaunion.uz` - backend API and WebSocket origin

## CI/CD

- every push to `main` runs the full quality gate
- deployment runs only if `make quality`, `make test`, Docker Compose validation, and Nginx validation all pass
```

- [ ] **Step 2: Rewrite AGENTS runtime guidance**

Replace temporary local-HTTPS runtime references in `AGENTS.md` with:

```md
### VPS Runtime

- `customer.novdaunion.uz` serves the mini app
- `admin.novdaunion.uz` serves the admin panel
- `srvr.novdaunion.uz` serves the backend API and WebSocket origin
- the temporary local HTTPS workflow has been retired
```

- [ ] **Step 3: Remove the temporary local HTTPS spec and plan**

Delete:

```text
docs/superpowers/specs/2026-03-31-local-https-telegram-miniapp-design.md
docs/superpowers/plans/2026-03-31-local-https-telegram-miniapp-implementation.md
```

- [ ] **Step 4: Run documentation integrity checks**

Run:

```bash
git diff --check
rg -n "xpressgo.home.arpa|Temporary Local HTTPS|local-https-telegram-miniapp" README.md AGENTS.md docs docker-compose.yml Makefile server web infra scripts
```

Expected:

- `git diff --check` prints nothing
- ripgrep only finds local-HTTPS references in historical commits or nowhere in tracked files

- [ ] **Step 5: Commit**

```bash
git add README.md AGENTS.md docs/superpowers/specs/2026-03-31-vps-subdomain-cicd-deployment-design.md
git add -u docs/superpowers/specs docs/superpowers/plans
git commit -m "docs(runtime): replace local https docs with vps deployment flow"
```

### Task 7: End-To-End Verification And Deployment Readiness

**Files:**
- Verify only

- [ ] **Step 1: Run full local quality before deployment enablement**

Run:

```bash
make quality
make test
make validate-compose
make validate-nginx
```

Expected: all pass

- [ ] **Step 2: Verify frontend runtime config outputs**

Run:

```bash
rg -n "customer.novdaunion.uz|admin.novdaunion.uz|srvr.novdaunion.uz" server web admin docker-compose.yml infra
```

Expected: the public runtime values point to the intended three-host architecture

- [ ] **Step 3: Verify no temporary local HTTPS workflow remains**

Run:

```bash
rg -n "xpressgo.home.arpa|.local-certs|local HTTPS|Temporary Local HTTPS" . -g'!node_modules' -g'!.git'
```

Expected: no remaining tracked implementation references for the retired workflow

- [ ] **Step 4: Commit final verification checkpoint if implementation required minor fixes**

```bash
git status --short
```

Expected: clean working tree, or only intended untracked local files

- [ ] **Step 5: Prepare VPS operator checklist**

Operator checklist:

```text
1. Install Docker Engine, Docker Compose plugin, Nginx, and Certbot on the VPS.
2. Clone the repo to /opt/xpressgo.
3. Create the VPS env file from infra/deploy/vps.env.example.
4. Install the Nginx site configs for customer/admin/srvr hostnames.
5. Issue Let's Encrypt certificates for all three domains.
6. Confirm DNS points each subdomain to the VPS.
7. Add GitHub Actions secrets for SSH deployment.
8. Push to main and verify the CI/CD pipeline deploys successfully.
```

- [ ] **Step 6: Final commit if needed**

```bash
git add .
git commit -m "chore(infra): finalize vps deployment readiness"
```

---

## Self-Review

Spec coverage:

- public three-host model: covered by Tasks 2, 3, and 6
- CI/CD on push to `main`: covered by Task 5
- strict quality gates before deploy: covered by Tasks 5 and 7
- cleanup of temporary local HTTPS workflow: covered by Tasks 1 and 6
- VPS Nginx and runtime setup: covered by Tasks 3, 4, and 7

Placeholder scan:

- no `TODO`, `TBD`, or deferred implementation placeholders remain
- each task names exact files, commands, and commit checkpoints

Type consistency:

- server runtime hostname: `customer.novdaunion.uz`
- backend runtime hostname: `srvr.novdaunion.uz`
- admin runtime hostname: `admin.novdaunion.uz`
- cleanup target: `xpressgo.home.arpa` local-HTTPS workflow
