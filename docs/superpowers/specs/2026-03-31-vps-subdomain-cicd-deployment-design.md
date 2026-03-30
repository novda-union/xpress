# Xpressgo VPS Subdomain CI/CD Deployment Design

## Purpose

Define the permanent deployment direction for Xpressgo on a Contabo Ubuntu VPS with strict subdomain separation, automated CI/CD, and explicit retirement of the temporary local-HTTPS Telegram workflow.

This design replaces the temporary laptop-hosted HTTPS workaround with a VPS-hosted deployment that uses real public HTTPS endpoints and push-to-main deployment automation.

## Goals

- deploy the monorepo to a public Ubuntu VPS
- use real HTTPS subdomains for the customer app, admin app, and backend
- keep the backend as the single source of truth for both frontends
- run deployment automatically on every push to `main`
- block deployment unless all quality and validation checks pass with zero errors
- retire the temporary local-HTTPS workflow and all related code and docs

## Public Domain Model

Use three public hostnames:

- `customer.novdaunion.uz`
- `admin.novdaunion.uz`
- `srvr.novdaunion.uz`

Responsibilities:

- `customer.novdaunion.uz`
  - serves only the customer-facing mini app frontend
  - is intended to be opened from Telegram Mini App flows
  - does not serve the admin panel
- `admin.novdaunion.uz`
  - serves only the admin frontend
  - does not serve the customer mini app
- `srvr.novdaunion.uz`
  - serves the Go backend API and WebSocket endpoints
  - is the single backend origin for both frontends

## Application Architecture

The VPS runs the Xpressgo apps with Docker Compose.

Containers:

- `postgres`
- `server`
- `web`
- `admin`

Host-level Nginx runs on Ubuntu outside Docker.

Nginx responsibilities:

- terminate HTTPS for all three public hostnames
- route `customer.novdaunion.uz` to the web container
- route `admin.novdaunion.uz` to the admin container
- route `srvr.novdaunion.uz` to the server container

The Go server remains the only backend service. Both frontends call it directly through:

- `https://srvr.novdaunion.uz`

This means:

- the web app uses `srvr.novdaunion.uz` as its API and WebSocket origin
- the admin app uses `srvr.novdaunion.uz` as its API and WebSocket origin
- the backend must explicitly allow both frontend origins

## Backend And Frontend Runtime Rules

### Server

The server must:

- treat `srvr.novdaunion.uz` as the canonical backend origin
- expose REST and WebSocket endpoints there
- explicitly allow requests from:
  - `https://customer.novdaunion.uz`
  - `https://admin.novdaunion.uz`
- use the customer app public URL for Telegram Mini App launch links

Telegram bot app URL:

- `https://customer.novdaunion.uz`

### Web

The web app must:

- load from `https://customer.novdaunion.uz`
- call the backend at `https://srvr.novdaunion.uz`
- use `wss://srvr.novdaunion.uz/ws` for order updates

### Admin

The admin app must:

- load from `https://admin.novdaunion.uz`
- call the backend at `https://srvr.novdaunion.uz`
- use `wss://srvr.novdaunion.uz/admin/ws` or the existing admin WebSocket route on the backend origin

## VPS Infrastructure Model

The Contabo Ubuntu VPS is the single runtime host for phase 1 deployment.

Recommended setup:

- Ubuntu host
- Docker Engine
- Docker Compose plugin
- Nginx installed on the host
- Let’s Encrypt certificates managed on the host

Recommended file ownership model:

- application code lives in a dedicated deploy directory on the VPS
- runtime secrets live in VPS env files outside Git
- Nginx site configs live in standard host Nginx paths

## CI/CD Model

Deployments are triggered on every push to `main`.

The deploy pipeline has two stages:

### 1. Quality Gate

This stage must pass completely before deployment can begin.

Required commands:

- `make quality`
- `make test`

Additional validation:

- Docker Compose config validation
- Nginx config validation for the production Nginx files

If any check fails:

- deployment does not run

### 2. Deploy

Runs only after the quality gate succeeds.

Recommended deploy flow:

1. GitHub Actions connects to the VPS over SSH
2. update repo state on the VPS to the pushed `main`
3. refresh runtime env files if needed
4. rebuild and restart app containers with Docker Compose
5. run database migrations
6. verify public health checks

Required post-deploy verification:

- `https://srvr.novdaunion.uz/health`
- customer frontend reachable
- admin frontend reachable

If deploy verification fails:

- the deploy job fails

## Secrets And Configuration

### GitHub Actions Secrets

Store deploy credentials in GitHub Actions secrets:

- deploy host
- deploy user
- SSH private key
- optional SSH port

### VPS Runtime Secrets

Keep runtime secrets on the VPS, not in Git:

- `TELEGRAM_BOT_TOKEN`
- `TELEGRAM_GATEWAY_TOKEN`
- `JWT_SECRET`
- database credentials if changed from default

### Public Runtime Values

Expected production-style values:

- customer app URL: `https://customer.novdaunion.uz`
- admin app URL: `https://admin.novdaunion.uz`
- backend API URL: `https://srvr.novdaunion.uz`
- Telegram Mini App URL: `https://customer.novdaunion.uz`

## Nginx Requirements

Host-level Nginx must define separate server blocks for:

- `customer.novdaunion.uz`
- `admin.novdaunion.uz`
- `srvr.novdaunion.uz`

Requirements:

- TLS termination for each hostname
- proxy pass to the correct container upstream
- WebSocket upgrade headers for backend WebSocket routes
- forwarding headers preserved correctly
- optional redirect from HTTP to HTTPS

## Cleanup Of Temporary Local HTTPS Workflow

The previous temporary local-HTTPS solution must be removed as part of this infra migration.

Remove:

- local `xpressgo.home.arpa` runtime defaults
- local cert bootstrap scripts if no longer needed
- local-only HTTPS proxy configuration used for the temporary laptop flow
- local temporary workflow sections in docs
- the local-HTTPS spec and implementation plan

Replace with:

- VPS runtime docs
- deployment docs
- CI/CD docs
- subdomain and Nginx docs

This migration should leave one clear deployment story, not two competing workflows.

## Verification Targets

After implementation, verify:

- Telegram bot opens the customer mini app at `https://customer.novdaunion.uz`
- customer app loads successfully from Telegram
- admin app loads successfully in the browser at `https://admin.novdaunion.uz`
- web app API calls succeed against `https://srvr.novdaunion.uz`
- admin API calls succeed against `https://srvr.novdaunion.uz`
- customer and admin WebSocket flows work against `srvr.novdaunion.uz`
- CORS only allows the intended public frontend origins
- CI blocks deploys on any failed quality or config checks
- push to `main` deploys successfully when all checks pass

## Out Of Scope

- blue/green deployment
- Kubernetes
- multi-server clustering
- CDN-specific frontend optimization beyond standard Cloudflare proxying
- automatic rollback orchestration beyond deploy failure reporting
