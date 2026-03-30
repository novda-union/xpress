---
name: xpressgo-architecture-review
description: Advisory-only architecture review for the Xpressgo repository. Use when reviewing diffs, planned changes, or current codebase structure for branch scope, auth and permissions, routing, cross-app impact, runtime workflow, shared patterns, or architectural drift across server, web, admin, and docs.
---

# Xpressgo Architecture Review

## Overview

Use this skill to review Xpressgo changes or architecture decisions without editing code. It is repo-specific and focuses on architectural fit, risk detection, cross-app impact, and drift from the project's current patterns.

This skill is advisory only. Do not implement fixes automatically unless the user explicitly asks for code changes after the review.

## Review Modes

Choose the lightest mode that matches the request.

### 1. Diff Review

Use when the user asks to review a change, patch, commit, or current worktree diff.

Focus on:

- branch and store scope correctness
- auth and permission boundaries
- API and route consistency
- cross-app impact on `server`, `web`, and `admin`
- documentation follow-up needs

### 2. Change Impact Review

Use when the user asks what a planned or completed change affects.

Focus on:

- which apps are impacted
- which layers are impacted
- whether docs, specs, or plans likely need updates
- whether there are hidden dependency edges

### 3. Drift Review

Use when the user asks whether the codebase is drifting from current architecture or patterns.

Focus on:

- layering violations
- duplicated patterns
- inconsistent branch-aware behavior
- runtime/workflow inconsistencies
- docs drifting from code reality

### 4. Pattern Fit Review

Use when the user proposes a new abstraction, structure, or implementation approach.

Focus on:

- whether it fits current repo patterns
- whether it introduces parallel structures
- whether the responsibility boundary is clear
- whether a smaller or more local change would be better

## Core Review Rules

### Branch-Aware Architecture

The current system is branch-aware operationally.

Review for:

- `store_id` remaining the tenant boundary
- `branch_id` being used for branch-scoped operations
- no accidental store-wide behavior where branch scope is required
- branch-aware menu, order, staff, and discovery behavior staying consistent

### Permissions and Roles

Roles are:

- `director`
- `manager`
- `barista`

Review for:

- directors remaining store-wide
- managers and baristas remaining branch-scoped
- admin server routes and admin UI behavior staying aligned

### Cross-App Impact

Server changes often have web/admin fallout.

Always check whether a change in one area implies updates in another:

- backend routes or payloads -> `web` and `admin`
- branch/permission logic -> backend, admin, and possibly docs
- runtime workflow -> `README.md`, `AGENTS.md`
- quality tooling -> `docs/code-quality-tooling.md`, `README.md`, `AGENTS.md`

### Docs and Reflection

If the reviewed change appears architectural or workflow-level, recommend the documentation workflow:

- `make docs-check`
- `make docs-refresh`
- `docs-sync`

Use `docs/registry.yml` as the routing source when deciding which docs likely need review.

## Files to Read First

Read only the files that match the review request.

For backend architecture reviews:

- `AGENTS.md`
- `server/internal/handler/router.go`
- `server/internal/middleware/auth.go`
- `server/internal/service/permission_service.go`

For web architecture reviews:

- `web/src/App.tsx`
- `web/src/store/cart.ts`
- `web/src/types/index.ts`
- `web/vite.config.ts`

For admin architecture reviews:

- `admin/layouts/default.vue`
- `admin/composables/useAuth.ts`
- `admin/composables/useBranchContext.ts`
- `admin/composables/usePermissions.ts`

For docs and workflow reviews:

- `README.md`
- `docs/code-quality-tooling.md`
- `docs/registry.yml`

## Output Format

Prefer clear, flat findings. If there are multiple findings, order them from highest risk to lowest risk.

Use this structure for each finding:

```text
Problem: <clear statement of the issue or risk>
Suggestion: <what should be reviewed, changed, or verified>
Evidence: <specific files, routes, flows, patterns, or diff details>
Reasoning: <why this matters architecturally>
```

If there are no material findings, say so explicitly and then list any residual risks or follow-up checks.

## Review Checklist

When reviewing, check these questions:

1. Does the change preserve branch-aware behavior correctly?
2. Does it respect current role and permission boundaries?
3. Does it fit the existing layering and folder structure?
4. Does it introduce a second parallel pattern where one already exists?
5. Does it imply changes in `web`, `admin`, or docs that were missed?
6. Does it change runtime or developer workflow?
7. Does it create documentation drift risk?

## Boundaries

Do not:

- auto-edit code
- auto-edit docs
- overstate certainty when evidence is weak
- demand changes for small local implementation details unless they create real architectural cost

Do:

- stay advisory
- cite concrete evidence
- call out cross-app or doc follow-up clearly
- distinguish between actual problems and watch items
