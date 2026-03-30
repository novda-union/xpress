# AGENTS Guide

This file is the project-specific operating guide for coding agents working in the Xpressgo repository.

## Purpose

Use this document to stay aligned with the actual architecture, development flow, and repo rules before making changes.

This repository is a multi-app system with:

- `server` - Go backend API, auth, order logic, branch logic, WebSocket hub
- `web` - React mini app for discovery, branch menu browsing, cart, and orders
- `admin` - Nuxt admin panel for branch operations, menu management, staff, and orders
- `postgres` - PostgreSQL via Docker

## Core Rules

### 1. Keep the branch-aware model intact

The system is no longer store-only in its operational flow.

- `store_id` is still the main tenant boundary
- `branch_id` is now required for branch operations
- discovery, menus, staff scope, and orders are branch-aware

Do not add new operational features that ignore `branch_id` unless the behavior is explicitly store-wide.

### 2. Preserve compatibility routes and flows unless intentionally removing them

The web app still contains a compatibility route:

- `/:slug` redirects into branch discovery or branch detail flow

Do not remove compatibility behavior casually. If compatibility is changed, verify the deep-link and redirect behavior intentionally.

### 3. Respect role and permission boundaries

Admin behavior depends on role and branch scope.

Current roles:

- `director`
- `manager`
- `barista`

Important behavior:

- directors are store-wide
- managers and baristas are branch-scoped
- admin UI and server routes should continue to honor these boundaries

Relevant backend logic lives in:

- [server/internal/middleware/auth.go](/home/dasturchioka/work/projects/xpressgo/server/internal/middleware/auth.go)
- [server/internal/service/permission_service.go](/home/dasturchioka/work/projects/xpressgo/server/internal/service/permission_service.go)

### 4. Prefer extending existing structure over inventing new parallel patterns

Follow the current layering:

- handlers orchestrate HTTP concerns
- repositories handle persistence
- services hold business rules
- middleware extracts auth and scope

On the frontend:

- pages define route-level screens
- components are grouped by domain
- hooks own reusable app behavior
- `lib` contains integration and helper utilities

### 5. Do not track local agent/tooling files

These are intentionally ignored and should stay untracked:

- `.agents/`
- `.claude/`

Do not reintroduce local skill/tool directories into Git.

## Project Map

## Server

Main entrypoints:

- `server/cmd/server`
- `server/cmd/migrate`
- `server/cmd/seed`

Important folders:

- `server/internal/handler`
- `server/internal/repository`
- `server/internal/service`
- `server/internal/middleware`
- `server/internal/ws`
- `server/internal/telegram`
- `server/migrations`

Current routing is defined in:

- [server/internal/handler/router.go](/home/dasturchioka/work/projects/xpressgo/server/internal/handler/router.go)

Key public routes:

- `/discover`
- `/branches/:id`
- `/branches/:id/menu`
- `/stores/:slug`
- `/stores/:slug/menu`
- `/auth/telegram`
- `/auth/dev`

Key authenticated user routes:

- `POST /orders`
- `GET /orders`
- `GET /orders/:id`
- `PUT /orders/:id/cancel`
- `GET /ws`

Key admin routes:

- `/admin/auth`
- `/admin/branches`
- `/admin/staff`
- `/admin/store`
- `/admin/menu`
- `/admin/orders`
- `/admin/ws`

## Web

The web app uses React, Vite, React Router, and Zustand.

Entry route map:

- `/`
- `/branch/:id`
- `/item/:id`
- `/cart`
- `/order/:id`
- `/orders`
- `/:slug` compatibility redirect route

Route setup is in:

- [web/src/App.tsx](/home/dasturchioka/work/projects/xpressgo/web/src/App.tsx)

Important directories:

- `web/src/pages`
- `web/src/components/auth`
- `web/src/components/discovery`
- `web/src/components/menu`
- `web/src/components/cart`
- `web/src/hooks`
- `web/src/store`
- `web/src/lib`

Important state/detail:

- cart state is branch-scoped
- discovery is location-based and category-filtered
- map UI is lazy loaded
- routes are lazy loaded for smaller initial bundles

## Admin

The admin app uses Nuxt 3 and is organized around a shell layout plus composables.

Important directories:

- `admin/layouts`
- `admin/pages`
- `admin/components/layout`
- `admin/components/branches`
- `admin/components/staff`
- `admin/components/ui`
- `admin/composables`
- `admin/types`

Important state/detail:

- auth state is stored in composables
- branch selection is managed in `useBranchContext`
- directors can switch branch context
- branch-scoped roles should not get store-wide controls

Relevant files:

- [admin/layouts/default.vue](/home/dasturchioka/work/projects/xpressgo/admin/layouts/default.vue)
- [admin/composables/useBranchContext.ts](/home/dasturchioka/work/projects/xpressgo/admin/composables/useBranchContext.ts)
- [admin/composables/usePermissions.ts](/home/dasturchioka/work/projects/xpressgo/admin/composables/usePermissions.ts)

## Runtime Workflows

Use the root `Makefile` as the default operator interface.

### Normal start with existing data

```bash
make up
```

This preserves:

- PostgreSQL data
- Docker volumes
- existing local state inside the running stack

### Fully fresh destructive reset

```bash
make fresh
```

This is intentionally destructive. It:

1. stops containers
2. removes containers
3. removes Docker volumes including PostgreSQL data
4. clears repo-local generated runtime artifacts
5. rebuilds and starts the stack
6. runs migrations
7. runs seed data

### Stop only

```bash
make down
```

### Restart without wiping DB

```bash
make restart
```

## Quality Workflow

Always prefer the repo-level quality flow before committing.

Primary commands:

- `make quality`
- `make quality-fix`
- `make quality-server`
- `make quality-web`
- `make quality-admin`

Supporting commands:

- `make fmt`
- `make fmt-check`
- `make lint`
- `make typecheck`
- `make test`

Default expectation before commit:

1. run focused checks while iterating
2. run `make quality`
3. review `git status`
4. commit only when the tree contains intended changes

Detailed quality usage is documented in:

- [docs/code-quality-tooling.md](/home/dasturchioka/work/projects/xpressgo/docs/code-quality-tooling.md)

## Documentation Maintenance

Use the advisory documentation workflow when a change affects architecture, workflow, shared patterns, or long-lived behavior.

Document ownership and routing live in:

- [docs/registry.yml](/home/dasturchioka/work/projects/xpressgo/docs/registry.yml)

Primary advisory tooling:

- `docs-sync` skill
- `make docs-check`
- `make docs-refresh`

Default workflow:

1. run `make quality`
2. run `make docs-check`
3. if the change is architecture-level, repo-wide, or a large workflow shift, run `make docs-refresh`
4. if the routing is still ambiguous, use the `docs-sync` skill for a deeper advisory pass
5. decide whether to update `README.md`, `AGENTS.md`, `docs/code-quality-tooling.md`, specs, or plans

What each tool is for:

- `make docs-check`
  - current diff review
  - best for normal post-change doc triage
- `make docs-refresh`
  - broader reflection pass over the tracked codebase
  - best after larger refactors, new subsystems, or significant architecture changes
- `docs-sync` skill
  - deeper advisory analysis
  - best when you want a tighter recommendation about which docs should change and why

Typical routing:

- runtime, Docker, or Makefile changes:
  - review `README.md`
  - usually review `AGENTS.md`
- auth, permissions, schema, routing, or scope changes:
  - review `AGENTS.md`
  - review relevant spec files
- quality tooling changes:
  - review `docs/code-quality-tooling.md`
  - review `README.md`
  - possibly review `AGENTS.md`
- major UI or shared pattern changes:
  - review relevant specs
  - review `AGENTS.md`

This workflow is advisory only. The user decides whether to apply the suggested documentation changes.

## Git Rules

- Do not commit ignored local tooling folders like `.agents/` and `.claude/`
- Do not commit generated runtime artifacts unless there is a deliberate reason
- Keep commits scoped to the change being made
- Check `git status` before and after edits
- Avoid reverting unrelated user changes

## Editing Guidance

### Backend

- preserve the handler -> service -> repository layering
- keep permission checks explicit in admin flows
- validate branch ownership when touching branch-scoped resources
- prefer server-side derivation of authoritative scope instead of trusting client payloads

### Web

- keep route-level code splitting in place
- avoid pulling heavy libraries into the initial bundle without reason
- preserve branch-scoped cart behavior
- prefer compatibility redirects over breaking legacy links abruptly

### Admin

- keep role-aware and branch-aware behavior consistent
- use existing composables for auth, API access, permissions, and branch context
- avoid introducing store-wide controls into branch-scoped screens

## Files to Read Before Large Changes

For backend work:

- [server/internal/handler/router.go](/home/dasturchioka/work/projects/xpressgo/server/internal/handler/router.go)
- [server/internal/middleware/auth.go](/home/dasturchioka/work/projects/xpressgo/server/internal/middleware/auth.go)
- [server/internal/service/permission_service.go](/home/dasturchioka/work/projects/xpressgo/server/internal/service/permission_service.go)

For web work:

- [web/src/App.tsx](/home/dasturchioka/work/projects/xpressgo/web/src/App.tsx)
- [web/src/store/cart.ts](/home/dasturchioka/work/projects/xpressgo/web/src/store/cart.ts)
- [web/src/types/index.ts](/home/dasturchioka/work/projects/xpressgo/web/src/types/index.ts)
- [web/vite.config.ts](/home/dasturchioka/work/projects/xpressgo/web/vite.config.ts)

For admin work:

- [admin/layouts/default.vue](/home/dasturchioka/work/projects/xpressgo/admin/layouts/default.vue)
- [admin/composables/useApi.ts](/home/dasturchioka/work/projects/xpressgo/admin/composables/useApi.ts)
- [admin/composables/useAuth.ts](/home/dasturchioka/work/projects/xpressgo/admin/composables/useAuth.ts)
- [admin/composables/useBranchContext.ts](/home/dasturchioka/work/projects/xpressgo/admin/composables/useBranchContext.ts)

For product intent and rollout context:

- [README.md](/home/dasturchioka/work/projects/xpressgo/README.md)
- [docs/superpowers/specs/2026-03-29-branches-discovery-design.md](/home/dasturchioka/work/projects/xpressgo/docs/superpowers/specs/2026-03-29-branches-discovery-design.md)
- [docs/superpowers/specs/2026-03-29-ui-ux-design-system.md](/home/dasturchioka/work/projects/xpressgo/docs/superpowers/specs/2026-03-29-ui-ux-design-system.md)
- [docs/code-quality-tooling.md](/home/dasturchioka/work/projects/xpressgo/docs/code-quality-tooling.md)

## Default Operating Checklist

When starting work:

1. read the relevant architecture files
2. inspect `git status`
3. decide whether the work is branch-scoped, store-scoped, or cross-app
4. make the smallest coherent change set that fits existing patterns
5. run focused verification
6. run `make quality` before finalizing

When finishing work:

1. verify behavior locally where practical
2. run the relevant quality commands
3. inspect the diff for accidental generated or local-only files
4. commit with a focused message
