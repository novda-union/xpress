# Documentation Maintenance and Reflection Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build an advisory-only documentation maintenance workflow that detects significant codebase changes, maps them to relevant docs, and provides both diff-based and full-reflection guidance without auto-editing files.

**Architecture:** Add a declarative document registry, a lightweight repo command for diff-based advisory checks, a reflection command for broader codebase analysis, and a repo-specific advisory skill that can read the diff plus targeted docs and recommend updates. Keep all outputs read-only and human-decision-first.

**Tech Stack:** Makefile, shell scripts, Python 3 CLI helpers, YAML registry, repo docs, custom skill scaffolding

---

## File Structure

### New Files

- `docs/registry.yml`
  - Declarative mapping of document ownership, change classes, and routing hints.
- `scripts/docs/docs_check.py`
  - Diff-based advisory analyzer used by `make docs-check`.
- `scripts/docs/docs_refresh.py`
  - Whole-codebase reflection analyzer used by `make docs-refresh`.
- `scripts/docs/common.py`
  - Shared helpers for registry loading, diff parsing, path classification, and output formatting.
- `docs/superpowers/specs/2026-03-30-documentation-maintenance-and-reflection-design.md`
  - Approved design spec already written; verify links stay correct.
- `.agents/skills/docs-sync/SKILL.md`
  - Advisory skill instructions for deep documentation sync analysis.
- `.agents/skills/docs-sync/examples.md`
  - Small repo-specific examples for how the skill should classify and suggest doc targets.

### Modified Files

- `Makefile`
  - Add `docs-check` and `docs-refresh` targets.
- `README.md`
  - Document documentation-maintenance workflow.
- `AGENTS.md`
  - Add instructions pointing future agents to the advisory doc workflow.
- `.gitignore`
  - Ignore any script-local cache or temp artifacts if introduced.

### Validation Inputs

- `git diff`
- `git diff --cached`
- `docs/code-quality-tooling.md`
- `docs/superpowers/specs/`
- `docs/superpowers/plans/`
- existing runtime and architecture files already listed in `AGENTS.md`

---

### Task 1: Add the Document Registry

**Files:**
- Create: `docs/registry.yml`
- Modify: `AGENTS.md`
- Test: `python3 scripts/docs/docs_check.py --help` after script scaffold exists in Task 2

- [ ] **Step 1: Create the registry file with the initial document ownership map**

```yaml
documents:
  readme:
    path: README.md
    class: runbook
    owns:
      - runtime-workflow
      - operator-commands
      - local-development-flow
  agents:
    path: AGENTS.md
    class: reflection
    owns:
      - architecture-reflection
      - repo-rules
      - shared-patterns
      - implementation-guidance
  quality:
    path: docs/code-quality-tooling.md
    class: runbook
    owns:
      - quality-workflow
      - lint-and-build-flow
  branches_discovery_spec:
    path: docs/superpowers/specs/2026-03-29-branches-discovery-design.md
    class: spec
    owns:
      - branch-discovery
      - branch-scope
      - ordering-flow
  design_system_spec:
    path: docs/superpowers/specs/2026-03-29-ui-ux-design-system.md
    class: spec
    owns:
      - design-system
      - shared-ui-patterns
  doc_maintenance_spec:
    path: docs/superpowers/specs/2026-03-30-documentation-maintenance-and-reflection-design.md
    class: spec
    owns:
      - documentation-maintenance
      - reflection-workflow

change_families:
  runtime:
    paths:
      - Makefile
      - docker-compose.yml
    tags:
      - runtime-workflow
      - operator-commands
  quality:
    paths:
      - scripts/quality/
      - server/.golangci.yml
      - web/package.json
      - admin/package.json
      - admin/eslint.config.mjs
      - docs/code-quality-tooling.md
    tags:
      - quality-workflow
      - lint-and-build-flow
  backend_architecture:
    paths:
      - server/internal/handler/
      - server/internal/middleware/
      - server/internal/service/
      - server/internal/repository/
      - server/migrations/
    tags:
      - architecture-reflection
      - branch-scope
      - public-api
      - permissions
  web_architecture:
    paths:
      - web/src/App.tsx
      - web/src/pages/
      - web/src/components/
      - web/src/store/
      - web/vite.config.ts
    tags:
      - architecture-reflection
      - shared-ui-patterns
      - branch-discovery
  admin_architecture:
    paths:
      - admin/layouts/
      - admin/pages/
      - admin/components/
      - admin/composables/
      - admin/types/
    tags:
      - architecture-reflection
      - shared-ui-patterns
      - permissions
```

- [ ] **Step 2: Add a short reference in `AGENTS.md` telling future agents where the registry lives**

```md
## Documentation Routing

The repo-level document ownership map lives in:

- `docs/registry.yml`

Use it when deciding which docs need review after architecture, workflow, or shared-pattern changes.
```

- [ ] **Step 3: Verify the registry file is valid YAML**

Run:

```bash
python3 - <<'PY'
import yaml
with open("docs/registry.yml", "r", encoding="utf-8") as f:
    data = yaml.safe_load(f)
assert "documents" in data
assert "change_families" in data
print("registry ok")
PY
```

Expected: `registry ok`

- [ ] **Step 4: Commit**

```bash
git add docs/registry.yml AGENTS.md
git commit -m "Add documentation registry"
```

---

### Task 2: Implement the Shared Advisory Engine Helpers

**Files:**
- Create: `scripts/docs/common.py`
- Create: `scripts/docs/__init__.py`
- Test: `python3 -m py_compile scripts/docs/common.py`

- [ ] **Step 1: Create the shared helper module**

```python
from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import subprocess
from typing import Iterable

import yaml


ROOT = Path(__file__).resolve().parents[2]
REGISTRY_PATH = ROOT / "docs" / "registry.yml"


@dataclass
class AdvisoryResult:
    classifications: list[str]
    severity: str
    documents: list[str]
    reasons: list[str]


def load_registry() -> dict:
    with REGISTRY_PATH.open("r", encoding="utf-8") as handle:
        return yaml.safe_load(handle)


def git_diff_names(staged: bool = False) -> list[str]:
    args = ["git", "diff", "--name-only"]
    if staged:
      args.append("--cached")
    proc = subprocess.run(args, cwd=ROOT, capture_output=True, text=True, check=True)
    return [line.strip() for line in proc.stdout.splitlines() if line.strip()]


def match_change_families(paths: Iterable[str], registry: dict) -> tuple[list[str], list[str]]:
    families: list[str] = []
    tags: list[str] = []
    for family_name, config in registry.get("change_families", {}).items():
        family_paths = config.get("paths", [])
        if any(any(path == family_path or path.startswith(family_path) for family_path in family_paths) for path in paths):
            families.append(family_name)
            tags.extend(config.get("tags", []))
    return families, sorted(set(tags))


def select_documents(tags: Iterable[str], registry: dict) -> list[str]:
    tag_set = set(tags)
    results: list[str] = []
    for doc in registry.get("documents", {}).values():
        owns = set(doc.get("owns", []))
        if tag_set & owns:
            results.append(doc["path"])
    return sorted(dict.fromkeys(results))


def classify_severity(paths: list[str], families: list[str]) -> str:
    if not families:
        return "no-doc-action"
    if any(family in {"runtime", "backend_architecture"} for family in families):
        return "strongly-recommend-update"
    if len(paths) <= 2:
        return "suggest-review"
    return "suggest-review"
```

- [ ] **Step 2: Run Python compile validation**

Run:

```bash
python3 -m py_compile scripts/docs/common.py
```

Expected: no output, exit code `0`

- [ ] **Step 3: Commit**

```bash
git add scripts/docs/common.py scripts/docs/__init__.py
git commit -m "Add shared docs advisory helpers"
```

---

### Task 3: Implement `make docs-check`

**Files:**
- Create: `scripts/docs/docs_check.py`
- Modify: `Makefile`
- Modify: `README.md`
- Test: `python3 scripts/docs/docs_check.py`

- [ ] **Step 1: Create the diff-based advisory script**

```python
from __future__ import annotations

import argparse
import sys

from common import classify_severity, git_diff_names, load_registry, match_change_families, select_documents


def main() -> int:
    parser = argparse.ArgumentParser(description="Advisory documentation relevance check")
    parser.add_argument("--staged", action="store_true", help="Inspect staged diff instead of working tree diff")
    args = parser.parse_args()

    registry = load_registry()
    paths = git_diff_names(staged=args.staged)

    if not paths:
        print("Change classification: none")
        print("Severity: no-doc-action")
        print("Recommended docs: none")
        print("Why: no changed files detected")
        return 0

    families, tags = match_change_families(paths, registry)
    docs = select_documents(tags, registry)
    severity = classify_severity(paths, families)

    print(f"Change classification: {', '.join(families) if families else 'local-only'}")
    print(f"Severity: {severity}")
    print("Recommended docs:")
    if docs:
        for doc in docs:
            print(f"- {doc}")
    else:
        print("- none")
    print("Why:")
    if families:
        for family in families:
            print(f"- matched change family: {family}")
    else:
        print("- no configured architecture or workflow family matched")
    print("Suggested next step:")
    if docs:
        print("- review the listed docs and decide whether they need updates")
        print("- use docs-refresh or docs-sync for deeper analysis")
    else:
        print("- no documentation action is likely required")
    return 0


if __name__ == "__main__":
    sys.exit(main())
```

- [ ] **Step 2: Add the new Make target**

```make
.PHONY: docs-check docs-refresh

docs-check:
	python3 scripts/docs/docs_check.py
```

- [ ] **Step 3: Document the new command in `README.md`**

```md
### Documentation Advisory Targets

- `make docs-check`
  - Analyze the current diff and suggest which docs may need review.
- `make docs-refresh`
  - Run a broader architecture reflection advisory pass.
```

- [ ] **Step 4: Run the new command**

Run:

```bash
make docs-check
```

Expected:

- exit code `0`
- printed classification, severity, and recommended docs

- [ ] **Step 5: Commit**

```bash
git add scripts/docs/docs_check.py Makefile README.md
git commit -m "Add docs-check advisory command"
```

---

### Task 4: Implement `make docs-refresh`

**Files:**
- Create: `scripts/docs/docs_refresh.py`
- Modify: `Makefile`
- Modify: `README.md`
- Test: `python3 scripts/docs/docs_refresh.py`

- [ ] **Step 1: Create the reflection script**

```python
from __future__ import annotations

from pathlib import Path

from common import ROOT, load_registry


REFLECTION_PATHS = [
    "server/internal/handler/router.go",
    "server/internal/middleware/auth.go",
    "server/internal/service/permission_service.go",
    "web/src/App.tsx",
    "web/src/store/cart.ts",
    "web/vite.config.ts",
    "admin/layouts/default.vue",
    "admin/composables/useBranchContext.ts",
]


def main() -> int:
    _ = load_registry()
    print("Reflection scope:")
    for rel in REFLECTION_PATHS:
        path = ROOT / rel
        state = "present" if path.exists() else "missing"
        print(f"- {rel}: {state}")
    print("Recommended docs:")
    print("- AGENTS.md")
    print("- README.md")
    print("- relevant specs under docs/superpowers/specs/")
    print("Why:")
    print("- docs-refresh is intended for broader architecture reflection, not just the current diff")
    print("Suggested next step:")
    print("- inspect architecture changes and decide whether long-lived docs need updates")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
```

- [ ] **Step 2: Add the Make target**

```make
docs-refresh:
	python3 scripts/docs/docs_refresh.py
```

- [ ] **Step 3: Document the reflection command in `README.md`**

```md
- `make docs-refresh`
  - Inspect the current codebase shape and suggest reflection-doc reviews for files like `AGENTS.md`.
```

- [ ] **Step 4: Run the reflection command**

Run:

```bash
make docs-refresh
```

Expected:

- exit code `0`
- printed reflection scope and recommended docs

- [ ] **Step 5: Commit**

```bash
git add scripts/docs/docs_refresh.py Makefile README.md
git commit -m "Add docs-refresh reflection command"
```

---

### Task 5: Add the Advisory `docs-sync` Skill

**Files:**
- Create: `.agents/skills/docs-sync/SKILL.md`
- Create: `.agents/skills/docs-sync/examples.md`
- Modify: `AGENTS.md`
- Test: read skill file and verify paths only

- [ ] **Step 1: Create the skill definition**

```md
---
name: docs-sync
description: Advisory-only documentation sync skill for Xpressgo. Reads the current diff, consults the document registry, and suggests which docs likely need review or updates.
---

# docs-sync

## Purpose

Use this skill after implementation work when the codebase may have changed in ways that affect long-lived documentation.

This skill is advisory only.

Never edit docs automatically unless the user explicitly asks you to apply the recommendations.

## Required Inputs

- current `git diff` or staged diff
- `docs/registry.yml`
- relevant candidate docs selected from the registry

## Workflow

1. Inspect changed files.
2. Classify the change family.
3. Identify recommended docs from `docs/registry.yml`.
4. Read only the likely affected docs.
5. Produce:
   - change classification
   - severity
   - docs to review
   - why each doc is relevant
   - optional suggested update outline

## Output Rules

- advisory only
- concise
- do not claim a doc must change unless the diff clearly invalidates it
```

- [ ] **Step 2: Add concrete repository examples**

```md
# docs-sync Examples

## Example: Makefile runtime change

Changed files:

- `Makefile`

Expected recommendation:

- `README.md`
- `AGENTS.md`

Why:

- run workflow changed
- operator-facing commands changed

## Example: permission model change

Changed files:

- `server/internal/middleware/auth.go`
- `server/internal/service/permission_service.go`

Expected recommendation:

- `AGENTS.md`
- relevant spec under `docs/superpowers/specs/`

Why:

- architecture reflection and access model changed
```

- [ ] **Step 3: Reference the skill in `AGENTS.md`**

```md
For significant architecture or workflow changes, run the advisory documentation flow:

1. `make docs-check`
2. `make docs-refresh` when reflection-level review is needed
3. `docs-sync` skill for deeper advisory routing and update suggestions
```

- [ ] **Step 4: Commit**

```bash
git add .agents/skills/docs-sync/SKILL.md .agents/skills/docs-sync/examples.md AGENTS.md
git commit -m "Add advisory docs-sync skill"
```

---

### Task 6: Integrate the Workflow Into Root Docs

**Files:**
- Modify: `README.md`
- Modify: `docs/code-quality-tooling.md`
- Modify: `AGENTS.md`
- Test: manual read-through

- [ ] **Step 1: Add a documentation maintenance section to `README.md`**

```md
## Documentation Maintenance

When a change affects architecture, workflow, shared patterns, or long-lived behavior:

1. run `make docs-check`
2. if the change is substantial, run `make docs-refresh`
3. decide whether to update `README.md`, `AGENTS.md`, specs, or workflow docs

This workflow is advisory only.
```

- [ ] **Step 2: Add a short reminder to `docs/code-quality-tooling.md`**

```md
## Documentation Follow-Up

After a substantial architecture, workflow, or quality-tooling change:

1. run `make quality`
2. run `make docs-check`
3. review suggested documentation targets before commit
```

- [ ] **Step 3: Make sure `AGENTS.md` points to both `docs/registry.yml` and the new commands**

```md
Primary advisory commands:

- `make docs-check`
- `make docs-refresh`
```

- [ ] **Step 4: Review the docs together for consistency**

Run:

```bash
sed -n '1,240p' README.md
sed -n '1,240p' AGENTS.md
sed -n '1,240p' docs/code-quality-tooling.md
```

Expected:

- command names match exactly
- advisory-only language is consistent
- no document claims auto-editing

- [ ] **Step 5: Commit**

```bash
git add README.md docs/code-quality-tooling.md AGENTS.md
git commit -m "Document advisory documentation workflow"
```

---

### Task 7: Validate the End-to-End Advisory Flow

**Files:**
- Test only

- [ ] **Step 1: Run repository quality checks**

Run:

```bash
make quality
```

Expected: pass

- [ ] **Step 2: Run the docs-check flow on the working tree**

Run:

```bash
make docs-check
```

Expected:

- prints severity and doc recommendations
- exits `0`

- [ ] **Step 3: Run the reflection flow**

Run:

```bash
make docs-refresh
```

Expected:

- prints reflection scope and recommended docs
- exits `0`

- [ ] **Step 4: Stage a sample change and verify staged mode works**

Run:

```bash
git add README.md
python3 scripts/docs/docs_check.py --staged
```

Expected:

- staged diff is read
- advisory output prints normally

- [ ] **Step 5: Final commit**

```bash
git add docs/registry.yml scripts/docs Makefile README.md AGENTS.md docs/code-quality-tooling.md .agents/skills/docs-sync
git commit -m "Implement advisory documentation maintenance workflow"
```

---

## Spec Coverage Check

- Registry and explicit routing logic: covered by Tasks 1 and 2.
- `make docs-check`: covered by Task 3.
- `docs-sync` skill: covered by Task 5.
- `make docs-refresh`: covered by Task 4.
- Advisory-only behavior: reinforced in Tasks 3, 4, 5, and 6.
- Workflow integration and document routing: covered by Tasks 1 and 6.

No uncovered spec requirements remain.
