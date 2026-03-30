# Xpressgo

Xpressgo is a multi-app project with:

- `server` - Go backend API and WebSocket server
- `web` - React mini app
- `admin` - Nuxt admin panel
- `postgres` - PostgreSQL database via Docker

The root `Makefile` is the main entry point for local development and quality checks.

## Run Flows

### Normal Start With Existing Data

Use this when you want to keep the current PostgreSQL data and existing Docker volume state.

1. Start the stack:
```bash
make up
```
2. View logs if needed:
```bash
make logs
```

This keeps:

- existing Postgres data
- existing Docker volume state
- previously seeded or manually created records

### Fully Fresh Start

Use this when you want a zero-state local environment with nothing preserved from previous runs.

```bash
make fresh
```

`make fresh` is destructive. It does all of the following:

1. stops all containers
2. removes containers
3. removes Docker volumes, including PostgreSQL data
4. removes repo-local generated runtime artifacts:
   - `web/dist`
   - `admin/.nuxt`
   - `admin/.output`
   - `server/tmp`
   - `server/bin`
5. rebuilds and starts the full Docker stack
6. runs database migrations
7. runs the seed command

After `make fresh`, the project is running again from a fully clean local state.

## Make Targets

### Runtime Targets

- `make up`
  - Start the Docker stack without deleting existing data.
- `make down`
  - Stop the Docker stack without deleting volumes.
- `make restart`
  - Restart containers while keeping existing data.
- `make fresh`
  - Fully destroy local runtime state and start again from zero.
- `make logs`
  - Follow Docker logs for the running services.

### Data Targets

- `make migrate`
  - Run database migrations from the `server` app.
- `make seed`
  - Seed demo data into the database.

### Local App Targets

These run apps directly from the host machine instead of Docker.

- `make server`
  - Run the Go server locally.
- `make web`
  - Run the Vite mini app locally.
- `make admin`
  - Run the Nuxt admin panel locally.

### Quality Targets

- `make quality`
  - Run repo-wide quality checks across server, web, and admin.
- `make quality-fix`
  - Apply safe autofixes, then rerun full quality checks.
- `make quality-server`
  - Run server-only quality checks.
- `make quality-web`
  - Run web-only quality checks.
- `make quality-admin`
  - Run admin-only quality checks.
- `make fmt`
  - Format Go files in `server`.
- `make fmt-check`
  - Fail if Go files are not properly formatted.
- `make lint`
  - Run lint checks across the repo.
- `make typecheck`
  - Run frontend type checks.
- `make test`
  - Run server tests.

### Documentation Advisory Targets

- `make docs-check`
  - Analyze the current diff and suggest which docs may need review.
- `make docs-refresh`
  - Run a broader reflection pass across the tracked codebase and suggest long-lived docs to review.

## Suggested Project Run Flow

### Most Common Daily Flow

Use this when you want to continue working with your existing local data:

1. `make up`
2. `make logs` if you need to inspect service output
3. make your changes
4. `make quality`

### Clean Rebuild Flow

Use this when you suspect stale local state or want to reset everything:

1. `make fresh`
2. confirm the apps are reachable:
   - web: `http://localhost:5173`
   - admin: `http://localhost:3000`
   - server: `http://localhost:8080`
3. continue development
4. `make quality`

### App-Only Local Flow

Use this when you want to run one or more apps outside Docker:

1. start dependencies you still need, usually Postgres:
```bash
make up
```
2. run the app you want locally:
```bash
make server
make web
make admin
```

## Documentation Maintenance

Use the documentation workflow when a change affects architecture, runtime commands, permissions, shared UI patterns, quality tooling, or any other long-lived repo behavior.

Primary tools:

- `make docs-check`
  - diff-based advisory review for the current working tree
- `make docs-refresh`
  - broader reflection review against the current tracked codebase
- `docs-sync` skill
  - deeper advisory analysis when the change is substantial and you want richer doc-routing suggestions

Recommended order after meaningful changes:

1. run `make quality`
2. run `make docs-check`
3. if the change is substantial, run `make docs-refresh`
4. if you need a more targeted advisory review, run the `docs-sync` skill
5. decide whether to update `README.md`, `AGENTS.md`, specs, plans, or workflow docs

Common document targets:

- `README.md`
  - runtime commands, operator flow, local development workflow
- `AGENTS.md`
  - architecture reflection, repo rules, agent guidance, shared implementation patterns
- `docs/code-quality-tooling.md`
  - lint, build, verification, and autofix workflow
- `docs/superpowers/specs/`
  - intended product or architecture behavior
- `docs/superpowers/plans/`
  - implementation sequencing, when the plan itself changes

This workflow is advisory only. It suggests likely stale docs; it does not edit them automatically.

## Notes

- `make up` is non-destructive.
- `make restart` is also non-destructive.
- `make fresh` is intentionally destructive and will wipe the local PostgreSQL data volume.
- If you need a true clean local environment, use `make fresh`, not `make restart`.
