# Admin Orders Tablet, Sidebar, and Telegram Design

## Summary

This change improves the admin panel for tablet-first branch operations and extends branch Telegram notifications with customer phone visibility.

The work has four parts:

1. close the sidebar when any navigation link is pressed, on both mobile and desktop
2. make order card actions icon-only with per-card loading and neutral disabled treatment
3. reshape the orders kanban into a single horizontally scrollable row of compact columns for tablet use
4. include customer phone number in every branch/store Telegram order status notification

The goal is to make the orders board more usable on tablets without removing operational context, and to make branch Telegram notifications more actionable when staff need to contact a customer.

## Current Context

### Admin sidebar

The sidebar is implemented through the shared sidebar primitives in:

- `admin/components/ui/sidebar/SidebarProvider.vue`
- `admin/components/ui/sidebar/Sidebar.vue`
- `admin/components/layout/AppSidebar.vue`

Navigation currently routes correctly, but clicking a link does not close the mobile sheet or collapse the desktop offcanvas sidebar.

### Orders board

The orders board is currently implemented in:

- `admin/pages/orders/index.vue`
- `admin/components/OrderCard.vue`

It uses a responsive grid for three status sections:

- new orders
- preparing
- ready

Cards are vertically stacked inside each section. Action buttons are text-based and take substantial horizontal space. There is no independent loading state per order card action.

### Telegram notifications

Order notifications for branch/store group chats currently flow through:

- `server/internal/handler/order_handler.go`
- `server/internal/service/order_notification_service.go`
- `server/internal/telegram/notifications.go`

Branch/store group messages already include order status information, but do not include the customer phone number. The customer phone exists on the user model and should be surfaced into all order status notifications sent to branch/store Telegram chats.

## Design Decisions

### 1. Sidebar close-on-navigation

#### Decision

Any sidebar navigation click should close the sidebar immediately.

#### Behavior

- on mobile, close the sheet sidebar
- on desktop, collapse the offcanvas sidebar
- this applies only to navigation links in the sidebar menu
- branch context selection and logout keep current behavior

#### Implementation direction

`AppSidebar.vue` should use the existing sidebar context (`useSidebar`) and wrap navigation clicks with a close handler.

This should not require changes to route definitions or layout structure. The behavior should be attached at the link/button level in the sidebar navigation menu.

## 2. Order card actions

### Decision

Keep the action controls in the same row and same status-specific positions, but convert them to icon-only buttons with accessible labels and per-card loading state.

### Behavior

- `pending` cards still show accept and reject
- `accepted` cards still show start preparing
- `preparing` cards still show mark ready
- `ready` cards still show picked up
- actions do not move to another part of the card
- actions become icon-only controls
- each icon button has an `sr-only` label and tooltip/title text

### Loading model

Loading state is owned by `admin/pages/orders/index.vue`, keyed by `order.id`.

Rules:

- when an action starts for one order, all action buttons on that same card become disabled
- the active button shows loading
- sibling action buttons on the same card use a neutral disabled style
- actions for every other order remain interactive
- websocket refreshes and reloads should clear or reconcile loading state safely after the async action completes

### Visual treatment

- loading buttons use a neutral disabled treatment rather than keeping the original semantic color
- icon buttons should remain large enough for tablet tapping
- recommended target size: about `36px` to `40px`
- spacing inside the action row should be tightened, not removed

## 3. Tablet-first orders kanban layout

### Decision

Replace the responsive multi-row grid with a single horizontally scrollable row of fixed-width status columns.

### Behavior

- all statuses appear in one row
- the board scrolls horizontally when the viewport is narrower than the total board width
- columns do not wrap to a second row
- each column remains vertically scrollable as part of the page content, not as an isolated internal scroller unless current UX requires it

### Column sizing

Each status column should use a stable minimum width suitable for tablets. The exact width can be tuned in implementation, but the intended result is:

- enough width for compact cards without crowding
- three columns visible partially or fully depending on tablet width
- predictable horizontal scan behavior

Recommended starting point:

- min width around `320px` per column

### Card density

Order cards should become smaller but still informational.

Keep visible:

- order number
- created time
- item count
- ETA
- branch short identifier
- payment method
- item lines
- total
- action controls

Compactness changes:

- reduce outer padding
- tighten vertical gaps
- keep metadata in shorter lines
- avoid large empty areas caused by wide text buttons

## 4. Telegram notifications with customer phone

### Decision

Include customer phone number in every order status notification sent to branch/store Telegram group chats.

### Scope

This includes all branch/store group order status notifications that currently mention an order lifecycle change.

It should not be limited only to new-order messages.

### Content rule

If the order user has a phone number:

- include a `Phone:` line in the Telegram message

If the phone number is absent:

- include a fallback such as `Phone: not provided`

This keeps the notification shape consistent for branch operators.

### Implementation direction

Extend the existing Telegram notification formatting path rather than creating a parallel notification builder.

The change should be made where order status messages are formatted for branch/store group chats so that all relevant statuses stay consistent.

## Component and State Boundaries

### `admin/components/layout/AppSidebar.vue`

Responsible for:

- rendering navigation items
- invoking sidebar close/collapse on nav click

Not responsible for:

- changing route permissions
- changing sidebar primitives

### `admin/pages/orders/index.vue`

Responsible for:

- loading orders
- storing per-order loading state
- disabling only the active card during async updates
- passing loading/disabled props to `OrderCard`
- rendering the horizontally scrollable board layout

Not responsible for:

- detailed button visuals
- formatting card internals

### `admin/components/OrderCard.vue`

Responsible for:

- compact display of order info
- icon-only action presentation
- neutral disabled visuals
- preserving action layout order

Not responsible for:

- owning async state
- fetching or mutating orders directly

### Telegram notification formatter path

Responsible for:

- adding customer phone line to all relevant branch/store group order status notifications

Not responsible for:

- changing user-facing mini app notifications unless explicitly required

## Error Handling

### Admin order actions

- if an action fails, only that order card should recover from loading
- other cards should remain unaffected
- existing error handling patterns on the page should be preserved

### Telegram phone formatting

- missing phone should not block notification delivery
- use fallback text instead of returning an error for absent phone data

## Testing and Verification

### Admin

Verify:

- clicking any sidebar nav link closes mobile sheet sidebar
- clicking any sidebar nav link collapses desktop sidebar
- order card buttons are icon-only
- one card entering loading disables only that card
- sibling buttons on the same card are neutral disabled while loading
- other cards remain active
- kanban columns stay in one row and scroll horizontally on tablet-sized widths
- cards remain readable and operationally sufficient at tablet width

### Backend

Verify:

- every branch/store Telegram order status notification includes phone information
- missing phone values render fallback text and still send successfully

### Quality checks

Run at minimum:

- `make quality-admin`
- backend-focused quality or tests for notification changes
- final `make quality`

## Out of Scope

- redesigning the sidebar information architecture
- changing order statuses or workflow semantics
- reordering action controls
- changing Telegram bot auth or customer-facing direct messages
- adding richer customer contact flows beyond displaying the phone number

## Recommended Implementation Order

1. sidebar close-on-navigation
2. per-order loading state in orders page
3. compact icon-only order card actions
4. horizontal kanban layout and card density adjustments
5. Telegram status notification phone line

This order keeps admin interaction changes isolated and makes the backend notification change independent from the UI adjustments.
