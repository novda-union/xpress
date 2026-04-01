# CI/CD Start Notifier Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Send a Telegram notification immediately when the GitHub Actions CI/CD workflow starts, while preserving the existing end-of-run success and failure notification behavior.

**Architecture:** Extend the existing `.github/workflows/ci-cd.yml` workflow with a dedicated `notify_start` job that runs before `quality`. Reuse the current Telegram notifier secrets and the same `curl`-based API call pattern already used by the final notification job so the change stays minimal and isolated.

**Tech Stack:** GitHub Actions workflow YAML, Telegram Bot HTTP API, `curl`

---

## File Map

- Modify: `.github/workflows/ci-cd.yml`
  - Add a new `notify_start` job that sends the start-of-run Telegram message.
  - Update `quality` job dependencies so the start notification happens first.
  - Preserve the existing `deploy` and final `notify` behavior.
- Reference: `docs/superpowers/specs/2026-04-01-ci-cd-start-notifier-design.md`
  - Source of truth for intended behavior and constraints.

## Task 1: Add workflow start notifier job

**Files:**
- Modify: `.github/workflows/ci-cd.yml`
- Reference: `docs/superpowers/specs/2026-04-01-ci-cd-start-notifier-design.md`

- [ ] **Step 1: Inspect the current workflow before editing**

Run:

```bash
sed -n '1,220p' .github/workflows/ci-cd.yml
```

Expected:

- The workflow contains `quality`, `deploy`, and `notify` jobs.
- `notify` already posts to Telegram using `NOTIFY_BOT_TOKEN` and `NOTIFY_CHAT_ID`.

- [ ] **Step 2: Add the new `notify_start` job and wire `quality` to depend on it**

Update `.github/workflows/ci-cd.yml` so the job structure matches this shape:

```yaml
name: CI/CD

on:
  push:
    branches:
      - master

concurrency:
  group: production-main
  cancel-in-progress: true

env:
  FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true

jobs:
  notify_start:
    runs-on: ubuntu-latest
    steps:
      - name: Send Telegram start notification
        env:
          BOT_TOKEN: ${{ secrets.NOTIFY_BOT_TOKEN }}
          CHAT_ID: ${{ secrets.NOTIFY_CHAT_ID }}
        run: |
          TEXT="xpressgo%0A%F0%9F%9A%80 CI%2FCD started"
          curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
            -d "chat_id=${CHAT_ID}&text=${TEXT}"

  quality:
    runs-on: ubuntu-latest
    needs: notify_start
    permissions:
      contents: read
```

Keep the rest of the `quality`, `deploy`, and `notify` jobs intact except for the new `needs` dependency on `quality`.

- [ ] **Step 3: Verify the final workflow file contains the intended job ordering**

Run:

```bash
sed -n '1,220p' .github/workflows/ci-cd.yml
```

Expected:

- `notify_start` appears before `quality`.
- `quality` includes `needs: notify_start`.
- `deploy` still includes `needs: quality`.
- `notify` still includes `needs: [quality, deploy]`.
- The final `notify` job still uses the existing success/failure logic.

- [ ] **Step 4: Review the diff for scope**

Run:

```bash
git diff -- .github/workflows/ci-cd.yml
```

Expected:

- Only `.github/workflows/ci-cd.yml` changes.
- The diff shows one new start notification job and one dependency update on `quality`.

- [ ] **Step 5: Commit the workflow change**

Run:

```bash
git add .github/workflows/ci-cd.yml
git commit -m "feat(ci): notify when ci cd starts"
```

Expected:

- A single focused commit is created for the workflow change.

## Task 2: Final verification

**Files:**
- Verify: `.github/workflows/ci-cd.yml`

- [ ] **Step 1: Confirm the worktree is clean except for intentional doc files**

Run:

```bash
git status --short
```

Expected:

- No unintended modified files remain.
- If the plan file is intentionally uncommitted, it is the only untracked file.

- [ ] **Step 2: Confirm the latest commits reflect spec plus workflow implementation**

Run:

```bash
git --no-pager log --oneline --decorate -4
```

Expected:

- One recent docs spec commit for the design file.
- One recent feature commit for the CI/CD workflow notifier.

- [ ] **Step 3: Summarize the resulting behavior**

Confirm these outcomes in the final handoff:

- A Telegram message is sent immediately when the workflow starts.
- The existing final success/failure notifier still runs at the end.
- No new secrets or extra workflow files were introduced.

---

## Self-Review

### Spec Coverage

- Immediate start notification: covered in Task 1, Steps 2 and 3.
- Preserve existing final notification behavior: covered in Task 1, Steps 2 and 3.
- Reuse existing secrets: covered in Task 1, Step 2.
- Keep implementation constrained to the workflow file: covered in Task 1, Step 4.

### Placeholder Scan

- No `TODO`, `TBD`, or unresolved placeholders remain.
- Commands, file paths, expected results, and commit message are explicit.

### Type and Naming Consistency

- The job name is consistently `notify_start`.
- The dependency chain is consistently `notify_start` -> `quality` -> `deploy`, with final `notify` depending on `quality` and `deploy`.
