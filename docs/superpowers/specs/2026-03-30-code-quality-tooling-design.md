# Xpressgo Code Quality Tooling Design

## Overview

This spec defines a repo-wide quality-check workflow for the Xpressgo monorepo covering:

1. **Server** (`server`) Go formatting, static analysis, and tests
2. **Mini App** (`web`) linting, typechecking, and production build validation
3. **Admin Panel** (`admin`) linting, typechecking, and production build validation

The goal is to give the team one reliable way to detect bad code quality, lint errors, type issues, and integration breakage before shipping changes, while also providing explicit autofix commands for safe formatting and lint fixes.

This work should improve local developer workflow first and be easy to reuse in CI later.

---

## Goals

- Add **read-only quality checks** that fail on real issues
- Add **explicit fix commands** for safe automatic cleanup
- Keep commands available both:
  - at the app level (`server`, `web`, `admin`)
  - at the repo root (`make quality`, etc.)
- Use stack-appropriate tooling rather than one generic wrapper
- Catch more than style issues:
  - formatting drift
  - lint errors
  - static analysis findings
  - type errors
  - build breakage

---

## Non-Goals

- Introducing a full pre-commit or hook system in this pass
- Enforcing maximum strictness on day one
- Adding snapshot/UI/browser test frameworks
- Reformatting the entire repo with a new formatter standard
- Replacing existing build scripts or local dev commands

---

## Design Principles

### 1. Checks must be explicit

`quality` commands are read-only and should never mutate files.

`fix` commands are separate and intentionally named so developers know when a command may edit files.

### 2. Quality includes integration signal

Builds and typechecks stay inside quality commands because many real regressions are not plain lint failures.

### 3. Keep tooling idiomatic per project

- Go code should use Go-native tooling
- React/Vite code should use ESLint + TypeScript build checks
- Nuxt/Vue code should use ESLint + `nuxi typecheck` + build

### 4. Start practical, then tighten

The initial lint configuration should be useful, not punitive. The first pass should surface real defects and obvious quality issues without creating unmanageable noise.

---

## Command Structure

### Per-Project Commands

Each project should expose its own quality scripts.

#### Server

Defined via root orchestration and helper scripts because `server` is a Go module rather than an npm package.

Commands:

- `fmt` — run `gofmt -w` on Go files
- `fmt-check` — run `gofmt -l` and fail on unformatted files
- `vet` — run `go vet ./...`
- `test` — run `go test ./...`
- `lint` — run `golangci-lint run`
- `quality` — run `fmt-check`, `vet`, `lint`, `test`

#### Web

Defined in [web/package.json](/home/dasturchioka/work/projects/xpressgo/web/package.json).

Commands:

- `lint` — ESLint read-only
- `lint:fix` — ESLint autofix
- `typecheck` — `tsc -b`
- `build` — existing Vite production build
- `quality` — `lint`, `typecheck`, `build`
- `quality:fix` — `lint:fix` followed by `quality`

#### Admin

Defined in [admin/package.json](/home/dasturchioka/work/projects/xpressgo/admin/package.json).

Commands:

- `lint` — ESLint read-only for Vue/Nuxt TypeScript files
- `lint:fix` — ESLint autofix
- `typecheck` — `nuxi typecheck`
- `build` — existing Nuxt production build
- `quality` — `lint`, `typecheck`, `build`
- `quality:fix` — `lint:fix` followed by `quality`

---

## Root-Level Orchestration

The root [Makefile](/home/dasturchioka/work/projects/xpressgo/Makefile) should become the main entry point for repo-wide checks.

### New Targets

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

### Target Behavior

#### `make quality`

Runs all project quality checks in sequence:

1. server quality
2. web quality
3. admin quality

Fails on the first error.

#### `make quality-fix`

Runs only safe automatic fixes, then revalidates:

1. server format fix
2. web lint fix
3. admin lint fix
4. full `make quality`

#### Focused targets

- `make quality-server` only checks Go code
- `make quality-web` only checks the mini app
- `make quality-admin` only checks the admin panel

This allows faster local iteration and easier CI decomposition later.

---

## Scripts Layout

Use a small helper directory for command composition where shell logic would otherwise make the `Makefile` noisy.

### New Directory

- `scripts/quality/`

### Planned Scripts

- `scripts/quality/server-quality.sh`
- `scripts/quality/server-fix.sh`
- optional:
  - `scripts/quality/web-quality.sh`
  - `scripts/quality/admin-quality.sh`

These scripts should:

- print clear step labels
- stop on first failure
- keep behavior deterministic
- avoid hidden mutation in read-only paths

The `web` and `admin` quality commands can remain package-script driven unless command composition becomes messy.

---

## Tooling Choices

### Server Tooling

#### `gofmt`

Use `gofmt` as the canonical formatter.

- Check mode: `gofmt -l`
- Fix mode: `gofmt -w`

#### `go vet`

Use `go vet ./...` for baseline Go correctness checks.

#### `golangci-lint`

Add `golangci-lint` with a curated config in:

- [server/.golangci.yml](/home/dasturchioka/work/projects/xpressgo/server/.golangci.yml)

Initial linters:

- `govet`
- `errcheck`
- `staticcheck`
- `ineffassign`
- `unused`
- `gofmt`
- `goimports`

This is intentionally practical rather than exhaustive.

### Web Tooling

The web app already has ESLint installed and a `lint` script.

Keep the current ESLint setup as the base signal and pair it with:

- `tsc -b`
- `vite build`

Type-aware linting is explicitly deferred unless the current codebase stabilizes enough to justify stricter rules.

### Admin Tooling

The admin app currently lacks lint coverage.

Add ESLint support for:

- `.vue`
- `.ts`
- Nuxt config and composables

Then combine it with:

- `nuxi typecheck`
- `nuxt build`

This gives good framework-aware signal without overengineering the stack.

---

## Fix Commands

Only safe, predictable fixes should run automatically.

### Allowed Autofixes

- Go formatting via `gofmt -w`
- ESLint autofix for `web`
- ESLint autofix for `admin`

### Not Included in Autofix

- dependency upgrades
- codemods
- broad refactors
- manual static analysis fixes
- test rewrites

If `quality:fix` still fails, developers must handle the remaining issues manually.

---

## Output and UX

Quality commands should be easy to scan.

### Output Rules

- Print which project is being checked
- Print which tool is currently running
- Fail with non-zero exit codes
- Do not swallow tool output

Example sequence:

1. `server: fmt-check`
2. `server: vet`
3. `server: lint`
4. `server: test`
5. `web: lint`
6. `web: typecheck`
7. `web: build`
8. `admin: lint`
9. `admin: typecheck`
10. `admin: build`

---

## Documentation

Add a short developer-facing guide that explains:

- which commands are read-only
- which commands mutate files
- when to use per-project vs root commands
- which tools must be installed locally

Candidate locations:

- root `README.md` if added later
- `docs/`
- comments in `Makefile`

At minimum, the command surface should be discoverable from the `Makefile`.

---

## Error Handling

### Missing tools

If `golangci-lint` is missing locally, the command should fail clearly and tell the user what is missing.

### Partial success

If one project fails, the repo-wide quality command should stop and return failure.

This keeps the output focused and avoids hiding the first blocking issue in noise.

---

## Testing Strategy

Verification for this tooling change should include:

1. Run server quality commands successfully on the repo
2. Run web quality commands successfully on the repo
3. Run admin quality commands successfully on the repo
4. Confirm `fix` commands mutate files only when intended
5. Confirm root `make quality` and `make quality-fix` behave as documented

This work is successful when a developer can run one command at the repo root and reliably learn whether the codebase currently has:

- formatting problems
- lint failures
- static analysis findings
- type issues
- build breakage

---

## File Changes Planned

### Root

- Modify: [Makefile](/home/dasturchioka/work/projects/xpressgo/Makefile)
- Create: `scripts/quality/server-quality.sh`
- Create: `scripts/quality/server-fix.sh`

### Server

- Create: [server/.golangci.yml](/home/dasturchioka/work/projects/xpressgo/server/.golangci.yml)

### Web

- Modify: [web/package.json](/home/dasturchioka/work/projects/xpressgo/web/package.json)

### Admin

- Modify: [admin/package.json](/home/dasturchioka/work/projects/xpressgo/admin/package.json)
- Create: `admin/eslint.config.mjs` or `admin/eslint.config.js`

---

## Open Decisions Resolved

- **Fix commands included?** Yes
- **Read-only quality path?** Yes
- **Repo root orchestration?** Yes
- **Go linting via `golangci-lint`?** Yes
- **Vue/Nuxt linting added to admin?** Yes
- **Heavy hook/CI framework in this pass?** No

---

## Final Recommendation

Implement a practical repo-wide quality system with:

- root `make` targets
- Go formatting + vet + `golangci-lint` + tests
- React ESLint + TypeScript + build
- Nuxt/Vue ESLint + typecheck + build
- explicit `fix` commands for safe autofixes only

This gives the team immediate value, improves code health, and creates a clean foundation for future CI enforcement.
