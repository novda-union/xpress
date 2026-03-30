# Xpressgo Documentation Maintenance and Reflection Design

## Overview

This spec defines an advisory-only documentation maintenance system for the Xpressgo repository.

The problem is not the absence of documentation. The repository already has several important document classes:

- product and architecture specs in `docs/superpowers/specs/`
- implementation plans in `docs/superpowers/plans/`
- operator and workflow docs like `README.md` and `docs/code-quality-tooling.md`
- agent-facing repo reflection in `AGENTS.md`

The gap is that these documents can drift when the codebase changes. Architectural decisions, runtime workflow changes, permission changes, UI system changes, or shared pattern changes may land in code without anyone explicitly refreshing the corresponding docs.

This system is intended to reduce that drift by:

1. detecting meaningful code and workflow changes
2. classifying whether they affect architecture, operations, patterns, or implementation intent
3. mapping those changes to likely affected documents
4. producing advisory suggestions for document review and updates

The system must not auto-edit docs by default. The user remains the final decision-maker.

---

## Goals

- Add a low-risk advisory workflow that flags likely stale documents
- Make documentation maintenance repo-aware rather than generic
- Support both diff-based checks and periodic whole-codebase reflection
- Help keep `AGENTS.md`, `README.md`, workflow docs, and relevant specs aligned with reality
- Keep suggestions fast enough for regular use and detailed enough for larger changes

---

## Non-Goals

- Automatically editing documentation on every change
- Forcing documentation updates for every small fix
- Rewriting implementation plans after every code change
- Replacing normal code review or architecture judgment
- Creating a mandatory blocking hook in the first pass

---

## Design Principles

### 1. Advisory first

The system should suggest documentation work, not silently mutate files.

Default behavior:

- analyze
- classify
- recommend

Optional future behavior:

- generate suggested patches
- apply only on explicit user request

### 2. Significant changes only

The system should distinguish between local edits and lasting system-level changes.

It should prioritize:

- architecture changes
- workflow changes
- shared pattern changes
- API, permission, schema, and routing changes

It should avoid noisy suggestions for:

- isolated copy changes
- local styling-only fixes
- small bug fixes with no shared impact

### 3. Document classes have different responsibilities

Not all docs should be treated equally.

- specs capture intended system or product design
- plans capture implementation sequencing
- reflection docs describe current code reality and rules
- runbooks describe developer or operator workflows

The advisory logic must route changes accordingly.

### 4. Keep the routing logic explicit

The system should not hardcode all knowledge in one script. It should rely on a document registry or mapping file so the repo can evolve without rewriting the tool logic every time.

---

## Proposed System

The system consists of three layers.

## Layer 1: `make docs-check`

This is the lightweight diff-based advisory check.

Purpose:

- inspect current repo changes
- detect whether they are documentation-relevant
- print a recommendation report

Inputs:

- working tree diff
- optionally staged diff
- optionally a commit range later

Behavior:

1. collect changed files
2. classify the changes by area
3. determine severity of documentation relevance
4. map the change to candidate docs
5. print an advisory summary

Output example:

```text
Change classification: runtime workflow + architecture reflection
Severity: strongly recommend doc update
Recommended docs:
- README.md
- AGENTS.md
Why:
- Makefile run flow changed
- destructive fresh-start command added
Suggested next step:
- run docs-sync for deeper analysis
```

This command should be fast enough for frequent use.

---

## Layer 2: `docs-sync` Skill

This is the deeper repo-aware analysis layer.

Purpose:

- analyze diff plus the relevant docs
- provide higher-confidence documentation recommendations
- optionally draft suggested changes without applying them automatically

Responsibilities:

1. inspect diff and classify the change
2. consult the document registry
3. read the likely affected docs
4. determine whether each doc needs:
   - no action
   - review
   - update recommendation
5. produce:
   - a short diagnosis
   - candidate documents
   - reasons
   - optional suggested edits or patch plan

The skill should remain advisory by default.

Possible future mode:

- “generate suggested doc patch blocks”

But even in that mode, applying edits should still require explicit user approval.

---

## Layer 3: `make docs-refresh`

This is the periodic full reflection pass.

Purpose:

- re-explore the codebase at a snapshot in time
- refresh long-lived reflection docs that describe current architecture and repo conventions

This is not just diff-based.

Use cases:

- after major architecture changes
- after large feature rollouts
- after significant refactors
- when `AGENTS.md` or other reflection docs are suspected to be stale

Primary targets for this command:

- `AGENTS.md`
- architecture overview docs added in the future
- workflow reference docs that summarize current repo structure

It should produce an advisory report first, and optionally a suggested update plan.

---

## Change Severity Model

The advisory system should classify documentation relevance into three levels.

### Level 1: No doc action

Criteria:

- local fix only
- no shared workflow or architecture implications
- no change to shared patterns or public behavior

Examples:

- isolated CSS tweak
- a small bug fix inside one component
- a local variable rename

### Level 2: Suggest doc review

Criteria:

- shared behavior changed
- a common pattern changed
- a workflow changed in a way that may affect docs

Examples:

- new Make target
- revised quality workflow
- changed admin page behavior that affects operator usage
- new shared frontend abstraction

### Level 3: Strongly recommend doc update

Criteria:

- architecture changed
- runtime workflow changed materially
- schema, auth, permissions, routing, or scope model changed
- intended system design or long-lived repo guidance is now different

Examples:

- branch-aware operational model
- permission model changes
- route structure changes
- migration/schema changes
- major UI design-system shifts
- destructive fresh-start command introduced

---

## Decision Detection Rules

The tool should flag documentation-relevant changes when diffs touch areas like:

- `server/internal/handler/router.go`
- `server/internal/middleware/`
- `server/internal/service/permission_service.go`
- `server/migrations/`
- root `Makefile`
- `docker-compose.yml`
- `web/src/App.tsx`
- shared web component and hook directories
- shared admin composables, layout, and shell structure
- quality tooling configs and scripts

Change families the system should recognize:

- routing
- auth
- permissions
- schema
- runtime workflow
- quality workflow
- design system / shared UI patterns
- branch/store scoping
- public API shape

It should treat a single file change as insufficient evidence by itself. The classifier should consider both file location and the kind of change.

Example:

- editing `README.md` alone is not an architecture change
- editing `Makefile` to add `make fresh` is a runtime workflow change

---

## Document Registry

The system should use a repo-level mapping file to connect change classes to documents.

Recommended location:

- `docs/registry.yml`

Alternative acceptable location:

- `.agents/docs-map.json`

Recommended structure:

```yaml
documents:
  readme:
    path: README.md
    owns:
      - runtime-workflow
      - operator-commands
  agents:
    path: AGENTS.md
    owns:
      - architecture-reflection
      - repo-rules
      - shared-patterns
  quality:
    path: docs/code-quality-tooling.md
    owns:
      - quality-workflow
      - lint-and-build-flow
```

This registry should allow the tool to stay declarative and extensible.

---

## Document Routing Model

The advisory system should use these routing rules.

### Runtime and command changes

Examples:

- `Makefile`
- Docker flow
- startup and reset commands

Primary docs:

- `README.md`
- `AGENTS.md`

### Quality tooling changes

Examples:

- lint rules
- quality scripts
- new quality commands

Primary docs:

- `docs/code-quality-tooling.md`
- `README.md`
- `AGENTS.md`

### Architecture and shared implementation guidance changes

Examples:

- branch scope changes
- permission model changes
- shared frontend architecture changes

Primary docs:

- `AGENTS.md`
- relevant spec file in `docs/superpowers/specs/`

### Intended system or product behavior changes

Examples:

- discovery model changes
- design-system change in intended UX
- order flow changes in intended product behavior

Primary docs:

- relevant spec file in `docs/superpowers/specs/`
- possibly `AGENTS.md` if implementation guidance also changes

### Implementation sequencing changes

Examples:

- revised rollout order
- newly deferred or re-scoped implementation tasks

Primary docs:

- relevant plan file in `docs/superpowers/plans/`

Important constraint:

Plan files should not be rewritten after every implementation detail. They should be updated only when execution sequencing or implementation scope meaningfully changes.

---

## Output Design

The advisory report should be concise and structured.

Minimum fields:

- change classification
- documentation severity
- recommended docs
- why they were selected
- suggested next step

Optional fields:

- likely stale section names
- suggested outline of the update
- confidence level

The output should be optimized for quick decision-making, not exhaustive prose.

---

## Suggested Commands

Initial command set:

- `make docs-check`
- `make docs-refresh`

Later optional command:

- `make docs-check-staged`

The `docs-sync` layer should exist as a custom skill rather than a plain shell command, because it benefits from richer repo context and doc reasoning.

---

## Suggested Workflow

### After a normal implementation task

1. run focused code verification
2. run `make quality`
3. run `make docs-check`
4. review the suggestions
5. decide whether to update docs

### After a significant architecture or workflow change

1. finish the code changes
2. run `make quality`
3. run `make docs-check`
4. run `docs-sync` for deeper analysis
5. update the relevant docs manually or with explicit approval

### Periodic reflection maintenance

1. run `make docs-refresh`
2. inspect suggested reflection updates
3. refresh `AGENTS.md` and related long-lived docs as needed

---

## Initial Implementation Order

The implementation should be sequenced conservatively.

### Phase 1

- add the document registry
- add `make docs-check`
- support diff classification and advisory output only

### Phase 2

- add the `docs-sync` skill
- support richer document analysis and suggested updates

### Phase 3

- add `make docs-refresh`
- support whole-codebase reflection summaries

### Phase 4

- optional advisory Git integration
- for example, a pre-commit reminder that points to `make docs-check`

This phase should remain advisory, not blocking, unless the team explicitly decides otherwise.

---

## Risks

### 1. Over-triggering

If the system suggests doc updates too often, it will be ignored.

Mitigation:

- conservative classifier
- severity model
- strong preference for “suggest review” over “must update”

### 2. Under-triggering

If the classifier is too weak, stale docs will remain stale.

Mitigation:

- maintain a document registry
- refine routing rules as the repo evolves
- use `docs-refresh` periodically

### 3. Confusing specs and plans

If the tool treats all docs alike, it may suggest rewriting plans unnecessarily.

Mitigation:

- explicit document-class rules
- plans are sequence docs, not reflection docs

### 4. False architecture signals from file-path heuristics

A changed file path alone may not imply a major decision.

Mitigation:

- combine path-based heuristics with change family classification
- bias toward advisory language instead of rigid enforcement

---

## Success Criteria

This system is successful when:

- meaningful architecture and workflow changes reliably trigger doc suggestions
- low-level changes usually do not trigger noisy advice
- users can quickly see which docs likely need updates
- `AGENTS.md`, `README.md`, and major specs stay closer to code reality over time
- the workflow remains lightweight enough to use regularly

---

## Recommendation

Implement the system as an advisory documentation-maintenance workflow centered on:

1. a document registry
2. `make docs-check`
3. a repo-aware `docs-sync` skill
4. a periodic `make docs-refresh` reflection pass

This gives the repository a practical mechanism for reducing documentation drift without introducing risky automatic edits or heavy process overhead.
