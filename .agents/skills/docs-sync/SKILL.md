---
name: docs-sync
description: Advisory-only documentation maintenance skill for Xpressgo. Analyze a diff or codebase snapshot, identify likely stale docs, and recommend updates without editing files automatically.
---

# docs-sync

## Purpose

Use this skill when a code change may affect the repository's long-lived documentation.

This skill is advisory only. Do not edit docs automatically unless the user explicitly asks for edits.

## What To Inspect

- current `git diff` or staged diff
- `AGENTS.md`
- `README.md`
- `docs/code-quality-tooling.md`
- relevant specs under `docs/superpowers/specs/`
- relevant plans under `docs/superpowers/plans/`
- any local document registry if present

## Workflow

1. classify the change as runtime, architecture, workflow, quality, or local-only
2. identify the likely affected docs
3. explain why each doc is relevant
4. rank the suggestions:
   - no doc action
   - suggest review
   - strongly recommend update
5. if asked, draft a suggested patch for the user to approve

## Routing Rules

- runtime or Makefile changes usually point to `README.md` and `AGENTS.md`
- architecture or permission changes usually point to `AGENTS.md` and a spec
- quality-tooling changes usually point to `docs/code-quality-tooling.md`, `README.md`, and `AGENTS.md`
- product or design-system changes usually point to the matching spec and possibly `AGENTS.md`
- implementation sequencing changes usually point to a plan file

## Output Rules

- keep the report concise
- name the likely docs explicitly
- state whether the recommendation is advisory or strong
- do not claim a doc must change unless the code change clearly invalidates it

## Related Workflow

Related repo commands:

- `make docs-check`
- `make docs-refresh`

These commands should follow the same classification and routing model as this skill.
