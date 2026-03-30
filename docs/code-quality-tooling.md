# Code Quality Tooling

This repository now has a single quality workflow for the Go server, the React mini app, and the Nuxt admin panel.

The quality checks are split into two categories:

- `check` commands: read-only validation, no file changes
- `fix` commands: safe automatic cleanup, then revalidation

## Tooling by App

### Server

The server quality flow runs:

1. `gofmt -l` via `fmt-check`
2. `go vet ./...`
3. `golangci-lint run ./...`
4. `go test ./...`

Useful commands:

- `make quality-server`
- `make fmt`
- `make fmt-check`

### Web

The mini app quality flow runs:

1. `eslint .`
2. `tsc -b`
3. `vite build`

Useful commands:

- `cd web && npm run quality`
- `cd web && npm run lint`
- `cd web && npm run lint:fix`
- `cd web && npm run typecheck`

### Admin

The admin quality flow runs:

1. `eslint .`
2. `nuxi typecheck`
3. `nuxt build`

Useful commands:

- `cd admin && npm run quality`
- `cd admin && npm run lint`
- `cd admin && npm run lint:fix`
- `cd admin && npm run typecheck`

## Root Commands

Use the root `Makefile` when you want one entry point.

- `make quality`
- `make quality-fix`
- `make quality-server`
- `make quality-web`
- `make quality-admin`
- `make lint`
- `make typecheck`
- `make test`
- `make fmt`
- `make fmt-check`

## Documentation Follow-Up

After a substantial architecture, workflow, or quality-tooling change:

1. run the relevant quality commands
2. run the advisory documentation workflow
3. review which docs likely need updates before committing

Advisory tooling:

- `docs-sync` skill
- `make docs-check`
- `make docs-refresh`

Likely docs to review:

- `README.md`
- `AGENTS.md`
- `docs/superpowers/specs/`
- `docs/superpowers/plans/`

## Suggested Working Order

This is the recommended local order when you are making changes.

### If you changed only one app

1. Run that app's focused quality command.
2. If the failure is formatting or autofixable lint, run the app fix command.
3. Rerun the same focused quality command until it is clean.
4. Before commit, run `make quality`.

Examples:

- server-only change:
  1. `make quality-server`
  2. if formatting failed, run `make fmt`
  3. rerun `make quality-server`
  4. finish with `make quality`

- web-only change:
  1. `cd web && npm run quality`
  2. if lint can autofix, run `cd web && npm run lint:fix`
  3. rerun `cd web && npm run quality`
  4. finish with `make quality`

- admin-only change:
  1. `cd admin && npm run quality`
  2. if lint can autofix, run `cd admin && npm run lint:fix`
  3. rerun `cd admin && npm run quality`
  4. finish with `make quality`

### If you changed multiple apps

Use this order:

1. `make quality-server`
2. `make quality-web`
3. `make quality-admin`
4. `make quality`

This order is intentional:

- server first, because backend errors usually block both apps
- web second, because it is faster to iterate than the Nuxt admin build
- admin last, because its full typecheck and build are heavier
- full `make quality` at the end, because the final gate should match the repo-wide workflow

## Fast Fix Flow

If you want the repository to apply safe fixes first, use:

1. `make quality-fix`
2. review the changed files
3. run `make quality` again if you made additional manual edits

`make quality-fix` will:

1. format Go files in `server`
2. run ESLint autofix in `web`
3. run ESLint autofix in `admin`
4. rerun the full repo quality flow

## Pre-Commit Checklist

Before committing:

1. run `make quality`
2. confirm there are no unexpected generated or local-only files in `git status`
3. commit only after the quality gate is green

## Notes

- The server scripts set local cache directories for Go and `golangci-lint` so the checks work in restricted environments too.
- The web build currently emits a large chunk warning. That is a performance follow-up, not a failing quality check.
