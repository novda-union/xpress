# CI/CD Start Notifier Design

## Summary

Add a Telegram notification at the beginning of the GitHub Actions CI/CD workflow so a run announces itself immediately, before quality checks and deployment complete.

## Current State

The repository has a single production workflow in `.github/workflows/ci-cd.yml`.

- The workflow triggers on pushes to `master`.
- `quality` runs first.
- `deploy` depends on `quality`.
- `notify` runs last with `if: always()` and sends the final success or failure result to Telegram using `NOTIFY_BOT_TOKEN` and `NOTIFY_CHAT_ID`.

There is currently no message when the workflow begins.

## Goal

When a CI/CD run starts, Telegram should receive a short "started" notification immediately. The existing end-of-run success/failure notification must remain intact.

## Recommended Approach

Add a dedicated `notify_start` job at the top of the workflow.

This job should:

- run on `ubuntu-latest`
- use the existing `NOTIFY_BOT_TOKEN` and `NOTIFY_CHAT_ID` secrets
- send a Telegram message that clearly indicates CI/CD has started

Then:

- `quality` should depend on `notify_start`
- `deploy` should keep depending on `quality`
- `notify` should keep depending on `quality` and `deploy`

This keeps the start notification separate from the final notification and makes the workflow order explicit.

## Message Format

The start message should follow the existing notifier style and stay concise:

- first line: `xpressgo`
- second line: `CI/CD started`

Encoding can stay aligned with the current `curl`-based Telegram API call pattern already used in the workflow.

## Constraints

- Do not introduce new repository secrets.
- Do not change the meaning of the final `notify` job.
- Do not add a reusable workflow or composite action for this change.
- Keep the implementation limited to the existing workflow file unless validation reveals a syntax issue.

## Verification

Validation for this change should be lightweight and local:

- inspect the rendered YAML shape for correctness
- ensure job dependencies still form the intended order:
  - `notify_start` -> `quality` -> `deploy`
  - final `notify` waits for `quality` and `deploy`
- confirm no unrelated files change

## Expected Outcome

After a push to `master`, Telegram receives:

1. an immediate "CI/CD started" message when the workflow begins
2. the existing final success/failure message when the workflow finishes
