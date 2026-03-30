# Code Quality Tooling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add repo-wide quality commands with explicit check vs fix behavior for `server`, `web`, and `admin`, covering formatting, linting, static analysis, typechecking, tests, and build validation.

**Architecture:** Keep quality checks tool-native per project, then unify them with root `Makefile` targets and small helper scripts for the Go module. The root commands remain the stable entrypoint, while project-level scripts stay directly usable in local development and CI.

**Tech Stack:** Go toolchain (`gofmt`, `go vet`, `go test`), `golangci-lint`, ESLint, TypeScript, Vite, Nuxt `nuxi typecheck`, GNU Make, POSIX shell.

---

## File Structure Map

### Root

- Modify: `Makefile`
- Create: `scripts/quality/server-quality.sh`
- Create: `scripts/quality/server-fix.sh`

### Server

- Create: `server/.golangci.yml`

### Web

- Modify: `web/package.json`

### Admin

- Modify: `admin/package.json`
- Create: `admin/eslint.config.mjs`

## Task 1: Add Server Lint Configuration

**Files:**
- Create: `server/.golangci.yml`

- [ ] **Step 1: Create a practical `golangci-lint` config**

Use a focused ruleset with these linters enabled:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - gofmt
    - goimports
    - govet
    - ineffassign
    - staticcheck
    - unused
```

- [ ] **Step 2: Keep the initial policy practical**

Add exclusions only if needed for generated or vendor paths. Do not enable an overly broad lint set in the first pass.

- [ ] **Step 3: Verify the config parses**

Run: `golangci-lint run ./...`

Expected: either lint findings or success, but not config parse errors

- [ ] **Step 4: Commit**

```bash
git add server/.golangci.yml
git commit -m "chore: add server lint configuration"
```

## Task 2: Add Server Quality Helper Scripts

**Files:**
- Create: `scripts/quality/server-quality.sh`
- Create: `scripts/quality/server-fix.sh`

- [ ] **Step 1: Add the read-only server quality script**

Create `scripts/quality/server-quality.sh` with:

```sh
#!/usr/bin/env sh
set -eu

echo "server: fmt-check"
unformatted="$(gofmt -l $(find server -type f -name '*.go'))"
if [ -n "${unformatted}" ]; then
  printf '%s\n' "$unformatted"
  echo "server: gofmt check failed"
  exit 1
fi

echo "server: vet"
(cd server && go vet ./...)

echo "server: lint"
(cd server && golangci-lint run ./...)

echo "server: test"
(cd server && go test ./...)
```

- [ ] **Step 2: Add the explicit server fix script**

Create `scripts/quality/server-fix.sh` with:

```sh
#!/usr/bin/env sh
set -eu

echo "server: fmt"
find server -type f -name '*.go' -exec gofmt -w {} +

echo "server: re-run quality"
"$(dirname "$0")/server-quality.sh"
```

- [ ] **Step 3: Make scripts executable**

Run: `chmod +x scripts/quality/server-quality.sh scripts/quality/server-fix.sh`

Expected: scripts are executable

- [ ] **Step 4: Smoke-test scripts**

Run: `scripts/quality/server-quality.sh`

Expected: executes gofmt check, vet, lint, and tests in order

- [ ] **Step 5: Commit**

```bash
git add scripts/quality/server-quality.sh scripts/quality/server-fix.sh
git commit -m "chore: add server quality scripts"
```

## Task 3: Add Web Quality Scripts

**Files:**
- Modify: `web/package.json`

- [ ] **Step 1: Add explicit web quality scripts**

Update scripts to include:

```json
{
  "scripts": {
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "typecheck": "tsc -b",
    "quality": "npm run lint && npm run typecheck && npm run build",
    "quality:fix": "npm run lint:fix && npm run quality"
  }
}
```

- [ ] **Step 2: Preserve existing dev/build scripts**

Do not remove existing `dev`, `build`, or `preview` commands.

- [ ] **Step 3: Verify web quality**

Run: `npm --prefix web run quality`

Expected: lint, typecheck, and build all pass

- [ ] **Step 4: Commit**

```bash
git add web/package.json
git commit -m "chore: add web quality scripts"
```

## Task 4: Add Admin ESLint and Quality Scripts

**Files:**
- Create: `admin/eslint.config.mjs`
- Modify: `admin/package.json`

- [ ] **Step 1: Add ESLint config for Nuxt/Vue TypeScript files**

Create `admin/eslint.config.mjs` with a practical flat config for:

- `*.ts`
- `*.vue`
- Nuxt app files

Use Vue/Nuxt-compatible parsing and a minimal ruleset that catches real issues without blocking the repo on stylistic noise.

- [ ] **Step 2: Add admin lint and quality scripts**

Update `admin/package.json` scripts to include:

```json
{
  "scripts": {
    "lint": "eslint .",
    "lint:fix": "eslint . --fix",
    "typecheck": "nuxi typecheck",
    "quality": "npm run lint && npm run typecheck && npm run build",
    "quality:fix": "npm run lint:fix && npm run quality"
  }
}
```

- [ ] **Step 3: Add any missing ESLint dependencies**

Install the smallest set needed to lint Vue/Nuxt files correctly.

Run: `npm --prefix admin install -D ...`

Expected: lint can execute successfully

- [ ] **Step 4: Verify admin quality**

Run: `npm --prefix admin run quality`

Expected: lint, typecheck, and build all pass

- [ ] **Step 5: Commit**

```bash
git add admin/eslint.config.mjs admin/package.json admin/package-lock.json
git commit -m "chore: add admin quality scripts"
```

## Task 5: Add Root Makefile Orchestration

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Add focused quality targets**

Extend the `Makefile` with:

```make
.PHONY: quality quality-fix quality-server quality-web quality-admin lint typecheck test fmt fmt-check

quality-server:
	./scripts/quality/server-quality.sh

quality-web:
	npm --prefix web run quality

quality-admin:
	npm --prefix admin run quality
```

- [ ] **Step 2: Add repo-wide quality targets**

Add:

```make
quality: quality-server quality-web quality-admin

quality-fix:
	./scripts/quality/server-fix.sh
	npm --prefix web run quality:fix
	npm --prefix admin run quality:fix
```

- [ ] **Step 3: Add focused convenience targets**

Add:

```make
lint:
	cd server && golangci-lint run ./...
	npm --prefix web run lint
	npm --prefix admin run lint

typecheck:
	npm --prefix web run typecheck
	npm --prefix admin run typecheck

test:
	cd server && go test ./...

fmt:
	./scripts/quality/server-fix.sh

fmt-check:
	./scripts/quality/server-quality.sh
```

Then adjust `fmt` and `fmt-check` so they run only the formatting-related portions instead of full quality if that separation is needed for clarity.

- [ ] **Step 4: Verify root orchestration**

Run: `make quality`

Expected: server, web, and admin quality commands execute in sequence and fail on the first real issue

- [ ] **Step 5: Commit**

```bash
git add Makefile
git commit -m "chore: add root quality targets"
```

## Task 6: Final Validation and Docs Touch-Up

**Files:**
- Modify: `Makefile`
- Optionally modify: `web/README.md`

- [ ] **Step 1: Run the full read-only flow**

Run: `make quality`

Expected: all checks pass

- [ ] **Step 2: Run the explicit fix flow**

Run: `make quality-fix`

Expected: safe fixes apply, then the full validation reruns successfully

- [ ] **Step 3: Confirm command discoverability**

Make sure root commands are obvious from the `Makefile`, and add a short note in docs if discoverability still feels weak.

- [ ] **Step 4: Commit**

```bash
git add Makefile web/README.md
git commit -m "docs: document quality workflow"
```

## Self-Review

- Spec coverage:
  - server formatting, vet, tests, and `golangci-lint` are covered by Tasks 1 and 2
  - web quality scripts are covered by Task 3
  - admin ESLint and quality scripts are covered by Task 4
  - root-level orchestration is covered by Task 5
  - fix vs check separation and final validation are covered by Task 6
- Placeholder scan: no `TODO`, `TBD`, or vague “add appropriate checks” steps remain
- Consistency check:
  - `quality` stays read-only
  - `quality:fix` is explicitly mutating
  - root commands orchestrate project-native commands rather than duplicating logic
