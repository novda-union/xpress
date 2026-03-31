# Telegram Order Notifications Design

## Purpose

Define how Xpressgo should send Telegram notifications for branch order activity using the existing customer-facing bot.

This design adds:

- Telegram group notifications for every admin-driven order status transition
- a branch-specific end-of-day order summary
- timezone-aware daily scheduling using hardcoded `Asia/Tashkent`

This design keeps notification logic in the backend so the admin panel remains a thin client.

## Goals

- send a Telegram notification to the branch group chat for each admin-driven order status transition
- send a daily branch summary to each branch group chat at the end of the local day
- keep notifications branch-aware and aligned with the existing branch-scoped order model
- reuse the existing Telegram bot already used for customer auth and Mini App launch
- avoid duplicate daily summary sends for the same branch and local date

## Non-Goals

- introducing a second Telegram bot for operations
- moving notification logic into the admin frontend
- building a general event bus or outbox framework
- adding store-wide daily rollup messages

## Current Behavior

Today the backend sends Telegram group messages only for:

- new customer orders
- customer-initiated order cancellations

The admin panel can configure `telegram_group_chat_id` for branches, but admin-driven order status changes do not currently send Telegram messages.

The code already includes a Telegram helper for direct user status messages, but that helper is not wired into the order status flow.

## Product Decision

Use the existing Telegram bot configured by the server for both customer interactions and operational notifications.

Notifications for admin-driven status changes should be sent to the branch's own `telegram_group_chat_id`.

Daily summaries should be sent per branch to that same branch chat only. If a branch does not have a configured Telegram group chat ID, that branch's daily summary should be skipped and logged.

## Order Status Notification Model

### Triggered statuses

Send a Telegram group message after a successful admin order status update for:

- `accepted`
- `preparing`
- `ready`
- `picked_up`
- `rejected`

These statuses match the existing admin workflow.

### Trigger point

The notification should be sent in the backend after the order status update succeeds in the admin order status endpoint flow.

The send must happen after persistence succeeds, not before.

### Delivery target

For admin-driven status updates:

- use the branch's `telegram_group_chat_id`
- if the branch does not have a chat ID, skip the Telegram send and log the reason

This is intentionally stricter than the current new-order and cancellation fallback behavior. The requested behavior is branch-specific, and each branch should receive only its own operational updates.

### Message content

Each status message should be compact and operationally useful.

Required content:

- order number
- branch name
- new status
- rejection reason when status is `rejected`
- total price
- item summary

Recommended status wording:

- `accepted`: order accepted
- `preparing`: preparation started
- `ready`: order ready for pickup
- `picked_up`: customer picked up the order
- `rejected`: order rejected, including reason when present

## Daily Summary Model

### Timing

The daily summary should use hardcoded `Asia/Tashkent`.

The summary should cover the full local calendar day for each branch and send shortly after the day ends. A practical default is a few minutes after midnight local time so the entire previous day is closed before reporting.

### Delivery target

Each branch summary must go to that branch's own `telegram_group_chat_id`.

No store-level fallback should be used for the daily summary.

### Summary scope

Each summary covers one branch and one local date.

The reporting window is:

- start: local day `00:00:00` in `Asia/Tashkent`
- end: next local day `00:00:00` in `Asia/Tashkent`

The server should convert that local window into UTC-backed query boundaries before reading from the database.

### Summary content

The daily summary should provide a full branch-level operational snapshot for the day.

Required content:

- branch name
- local date being summarized
- total orders
- counts by status:
  - `pending`
  - `accepted`
  - `preparing`
  - `ready`
  - `picked_up`
  - `rejected`
  - `cancelled`
- total value of all created orders for the day
- total value and count of `picked_up` orders
- total rejected count
- total cancelled count

Optional future expansion, but not required for the first implementation:

- top ordered items
- average fulfillment time
- hourly breakdown

## Scheduling Model

Use a backend in-process scheduler started by the server.

Recommended shape:

- load `Asia/Tashkent` once at startup
- run a lightweight ticker every minute
- detect when local time has crossed the configured summary send threshold
- generate summaries for the just-finished local day

This keeps the implementation small and consistent with the current single-server runtime model.

If the runtime later becomes multi-instance, the scheduling mechanism may need to move to a single-runner or job-locking model.

## Duplicate Prevention

Daily summaries must not be sent more than once per branch for the same local date.

To enforce that, persist a delivery record keyed by:

- notification type
- branch ID
- local summary date

This record should be written only after a successful send.

If a send fails, the failed branch can be retried later because the delivery record will not exist yet.

This persistence is required so server restarts do not cause duplicate daily summary messages.

## Failure Handling

Notification sending should not roll back the order status update itself.

For admin status changes:

- persist the status update first
- attempt Telegram delivery after persistence
- if Telegram delivery fails, log the failure with branch and order context
- still return success for the status update if the database change succeeded

For daily summaries:

- process branches independently
- if one branch send fails, continue processing other branches
- log skipped branches with missing chat IDs
- log failed sends with enough context for manual follow-up

## Architecture Changes

### Backend responsibilities

Handlers:

- keep HTTP orchestration in the order handler
- call notification helpers only after successful order updates

Telegram package:

- add branch group status message helpers
- add daily summary message helper

Service or repository layer:

- add a branch/day summary query or service helper that computes daily metrics from orders
- add a persistence mechanism for sent daily summaries

Runtime startup:

- start the summary scheduler from the server process when the Telegram bot is enabled

### Admin responsibilities

No new notification logic should be added to the admin frontend.

The admin app remains responsible only for invoking the order status update endpoint and configuring branch Telegram group chat IDs.

## Data Requirements

The implementation needs:

- branch Telegram group chat ID from branch data
- order details for status-change messages
- daily branch order aggregates
- a stored record of already-sent daily summaries

The exact storage format for summary delivery tracking can be a dedicated table keyed by branch, local date, and notification type.

## Testing Strategy

### Backend tests

Add focused tests for:

- status update flow triggers the correct Telegram helper for each admin-driven status
- `rejected` includes the rejection reason
- missing branch chat ID skips send without failing the request
- summary aggregation returns correct counts and totals for a branch and local date
- duplicate-prevention logic blocks a second send for the same branch and date
- scheduler date-window logic uses `Asia/Tashkent` correctly around day boundaries

### Manual verification

Verify:

- admin order accept sends a branch-group message
- admin reject sends a branch-group message with reason
- admin prepare, ready, and picked-up transitions each send one branch-group message
- branches receive only their own messages
- a branch without a Telegram group chat ID does not receive sends
- end-of-day summary contains the expected branch totals for the prior local day

## Rollout Notes

This change should preserve existing new-order and customer-cancel notification behavior while adding the new admin-status and end-of-day summary behavior.

The implementation should avoid changing tenant boundaries:

- `store_id` remains the tenant boundary
- `branch_id` remains the operational scope for notifications

## Open Decisions Resolved

The following choices are fixed by this design:

- use the existing customer-facing Telegram bot
- send admin-driven status notifications for every current admin transition
- use branch-specific group chat IDs
- use hardcoded `Asia/Tashkent`
- send one branch summary per day per branch group
- skip and log branches that do not have a configured Telegram group chat ID
