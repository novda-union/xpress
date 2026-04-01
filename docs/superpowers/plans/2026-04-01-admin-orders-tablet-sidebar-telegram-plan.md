# Admin Orders Tablet, Sidebar, and Telegram Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Improve the admin orders experience for tablet use, collapse the sidebar on navigation in both mobile and desktop modes, and include customer phone numbers in every branch/store Telegram order status notification.

**Architecture:** The admin work stays split between sidebar navigation behavior, order-board page state, and presentational order-card changes. Backend notification work extends the existing order status notification path by plumbing customer phone onto loaded orders and formatting it in Telegram message builders, rather than creating a parallel notification system.

**Tech Stack:** Nuxt 3, Vue 3, TypeScript, lucide-vue-next, reka-ui sidebar primitives, Go, Echo, pgx v5, Telegram Bot API

---

## File Map

| File | Responsibility |
|---|---|
| `admin/components/layout/AppSidebar.vue` | Close/collapse sidebar when a navigation item is pressed |
| `admin/pages/orders/index.vue` | Per-order loading state, horizontal kanban layout, fixed-width columns |
| `admin/components/OrderCard.vue` | Compact tablet-friendly card layout, icon-only actions, neutral disabled/loading state |
| `server/internal/model/order.go` | Add customer phone field to loaded order shape |
| `server/internal/repository/order_repo.go` | Load customer phone with order queries used by status notifications and admin listings |
| `server/internal/telegram/notifications.go` | Include phone line in branch/store order status messages |

## Task 1: Sidebar close-on-navigation

**Files:**
- Modify: `admin/components/layout/AppSidebar.vue`

- [ ] **Step 1: Import the sidebar context**

In `admin/components/layout/AppSidebar.vue`, add `useSidebar` to the sidebar imports and create the local context handle in `<script setup>`:

```ts
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
  useSidebar,
} from '@/components/ui/sidebar'
```

```ts
const sidebar = useSidebar()
```

- [ ] **Step 2: Add a nav click handler that closes both modes correctly**

In the same file, add a helper below `isActive`:

```ts
function closeSidebarAfterNavigation() {
  if (sidebar.isMobile.value) {
    sidebar.setOpenMobile(false)
    return
  }

  sidebar.setOpen(false)
}
```

- [ ] **Step 3: Wire the handler to sidebar navigation links**

Update the `NuxtLink` inside the navigation menu:

```vue
<NuxtLink :to="item.to" @click="closeSidebarAfterNavigation">
  <component :is="item.icon" />
  <span>{{ item.label }}</span>
</NuxtLink>
```

Do not add this handler to branch selection or logout.

- [ ] **Step 4: Run admin quality**

Run:

```bash
make quality-admin
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add admin/components/layout/AppSidebar.vue
git commit -m "feat(admin): close sidebar after navigation"
```

## Task 2: Per-order loading state on the orders board

**Files:**
- Modify: `admin/pages/orders/index.vue`

- [ ] **Step 1: Add loading state refs keyed by order id**

In `admin/pages/orders/index.vue`, add loading refs in `<script setup>`:

```ts
const loadingByOrderId = ref<Record<string, string | null>>({})
```

The value should store the in-flight action name, for example:

- `'accept'`
- `'reject'`
- `'mark-ready'`
- `'picked-up'`

- [ ] **Step 2: Add helper functions for card loading/disabled state**

Add these helpers below the existing computed order groups:

```ts
function isOrderLoading(orderId: string) {
  return Boolean(loadingByOrderId.value[orderId])
}

function orderLoadingAction(orderId: string) {
  return loadingByOrderId.value[orderId] ?? null
}
```

- [ ] **Step 3: Wrap status mutation with per-order loading guards**

Replace `updateStatus` with:

```ts
async function updateStatus(orderId: string, status: string, reason = '', action = 'update') {
  if (loadingByOrderId.value[orderId]) {
    return
  }

  loadingByOrderId.value = {
    ...loadingByOrderId.value,
    [orderId]: action,
  }

  try {
    await api(`/admin/orders/${orderId}/status`, {
      method: 'PUT',
      body: { status, reason },
    })
    await loadOrders()
  } finally {
    const next = { ...loadingByOrderId.value }
    delete next[orderId]
    loadingByOrderId.value = next
  }
}
```

- [ ] **Step 4: Update reject flow to use the same loading path**

Replace `rejectOrder` with:

```ts
async function rejectOrder(orderId: string) {
  if (loadingByOrderId.value[orderId]) {
    return
  }

  const reason = window.prompt('Rejection reason:')
  if (reason === null) {
    return
  }

  await updateStatus(orderId, 'rejected', reason, 'reject')
}
```

- [ ] **Step 5: Pass loading props into each `OrderCard`**

Update the three `OrderCard` usages so each receives:

```vue
:loading="isOrderLoading(order.id)"
:loading-action="orderLoadingAction(order.id)"
```

For example, the first card block becomes:

```vue
<OrderCard
  v-for="order in newOrders"
  :key="order.id"
  :order="order"
  :loading="isOrderLoading(order.id)"
  :loading-action="orderLoadingAction(order.id)"
  @accept="updateStatus(order.id, order.status === 'pending' ? 'accepted' : 'preparing', '', 'accept')"
  @reject="rejectOrder(order.id)"
/>
```

Use matching action labels for the other columns:

- preparing card: `'mark-ready'`
- ready card: `'picked-up'`

- [ ] **Step 6: Run admin quality**

Run:

```bash
make quality-admin
```

Expected: PASS or template/type failures only in `OrderCard.vue` because the new props are not added yet. If it already passes, continue.

- [ ] **Step 7: Commit**

```bash
git add admin/pages/orders/index.vue
git commit -m "feat(admin): add per-order loading state to orders board"
```

## Task 3: Compact icon-only order card actions

**Files:**
- Modify: `admin/components/OrderCard.vue`

- [ ] **Step 1: Add loading props and action icons**

In `admin/components/OrderCard.vue`, update the script imports:

```ts
import {
  Check,
  ChefHat,
  LoaderCircle,
  PackageCheck,
  X,
} from 'lucide-vue-next'
```

Replace the props declaration with:

```ts
const props = defineProps<{
  order: AdminOrder
  loading?: boolean
  loadingAction?: string | null
}>()
```

- [ ] **Step 2: Add helpers for neutral disabled button styling**

In `<script setup>`, add:

```ts
function actionIsLoading(action: string) {
  return props.loading && props.loadingAction === action
}

const neutralDisabledClass =
  'disabled:border-muted disabled:bg-muted disabled:text-muted-foreground disabled:opacity-100'
```

- [ ] **Step 3: Make the card more compact for tablet use**

Adjust the key layout classes in the template:

- card content: `p-4` -> `p-3`
- first row bottom margin: `mb-3` -> `mb-2`
- metadata line: keep on one compact line with `text-xs`
- items section margin: `mb-4` -> `mb-3`
- total section margin: `mb-4` -> `mb-3`

Update the ETA/branch/payment line to stay compact:

```vue
<p class="mb-2 text-xs text-muted-foreground">
  ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }} · {{ formatPayment(order.payment_method) }}
</p>
```

- [ ] **Step 4: Replace text buttons with icon-only buttons in the same row**

Replace the current action row with icon-only buttons that stay in the same order and same location:

```vue
<div class="flex gap-2">
  <Button
    v-if="order.status === 'pending'"
    size="icon"
    class="h-9 w-9"
    :class="neutralDisabledClass"
    :disabled="loading"
    :title="actionIsLoading('accept') ? 'Loading' : 'Accept order'"
    @click="$emit('accept')"
  >
    <LoaderCircle v-if="actionIsLoading('accept')" class="h-4 w-4 animate-spin" />
    <Check v-else class="h-4 w-4" />
    <span class="sr-only">Accept order</span>
  </Button>

  <Button
    v-if="order.status === 'pending'"
    size="icon"
    variant="destructive"
    class="h-9 w-9"
    :class="neutralDisabledClass"
    :disabled="loading"
    :title="actionIsLoading('reject') ? 'Loading' : 'Reject order'"
    @click="$emit('reject')"
  >
    <LoaderCircle v-if="actionIsLoading('reject')" class="h-4 w-4 animate-spin" />
    <X v-else class="h-4 w-4" />
    <span class="sr-only">Reject order</span>
  </Button>

  <Button
    v-if="order.status === 'accepted'"
    size="icon"
    class="h-9 w-9"
    :class="neutralDisabledClass"
    :disabled="loading"
    :title="actionIsLoading('accept') ? 'Loading' : 'Start preparing'"
    @click="$emit('accept')"
  >
    <LoaderCircle v-if="actionIsLoading('accept')" class="h-4 w-4 animate-spin" />
    <ChefHat v-else class="h-4 w-4" />
    <span class="sr-only">Start preparing</span>
  </Button>

  <Button
    v-if="order.status === 'preparing'"
    size="icon"
    class="h-9 w-9"
    :class="neutralDisabledClass"
    :disabled="loading"
    :title="actionIsLoading('mark-ready') ? 'Loading' : 'Mark ready'"
    @click="$emit('mark-ready')"
  >
    <LoaderCircle v-if="actionIsLoading('mark-ready')" class="h-4 w-4 animate-spin" />
    <PackageCheck v-else class="h-4 w-4" />
    <span class="sr-only">Mark ready</span>
  </Button>

  <Button
    v-if="order.status === 'ready'"
    size="icon"
    variant="outline"
    class="h-9 w-9"
    :class="neutralDisabledClass"
    :disabled="loading"
    :title="actionIsLoading('picked-up') ? 'Loading' : 'Mark picked up'"
    @click="$emit('picked-up')"
  >
    <LoaderCircle v-if="actionIsLoading('picked-up')" class="h-4 w-4 animate-spin" />
    <Check v-else class="h-4 w-4" />
    <span class="sr-only">Mark picked up</span>
  </Button>
  </div>
```

This preserves action placement while making the controls compact.

- [ ] **Step 5: Run admin quality**

Run:

```bash
make quality-admin
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add admin/components/OrderCard.vue
git commit -m "feat(admin): compact order card actions for tablet use"
```

## Task 4: Single-row horizontally scrollable kanban

**Files:**
- Modify: `admin/pages/orders/index.vue`

- [ ] **Step 1: Replace the grid wrapper with horizontal scroll**

In `admin/pages/orders/index.vue`, replace:

```vue
<div class="grid gap-4 xl:grid-cols-3">
```

with:

```vue
<div class="-mx-4 overflow-x-auto px-4 pb-2 lg:mx-0 lg:px-0">
  <div class="flex min-w-max gap-4">
```

And add the matching closing `</div></div>` around the three sections.

- [ ] **Step 2: Give each status column a fixed tablet-friendly width**

Update each `<section>` class from:

```vue
class="space-y-3 rounded-xl border bg-muted/30 p-4 border-t-4 ..."
```

to:

```vue
class="w-[20rem] shrink-0 space-y-3 rounded-xl border border-t-4 bg-muted/30 p-3 ..."
```

Keep the existing top border color per status.

- [ ] **Step 3: Tighten empty states and stack spacing**

Update the empty-state copy containers from `text-sm` to `text-xs` where needed and keep card stacks at:

```vue
<div class="space-y-2.5">
```

for denser tablet scanning.

- [ ] **Step 4: Run admin quality**

Run:

```bash
make quality-admin
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add admin/pages/orders/index.vue
git commit -m "feat(admin): make orders kanban horizontally scrollable"
```

## Task 5: Add customer phone to order status notification payload

**Files:**
- Modify: `server/internal/model/order.go`
- Modify: `server/internal/repository/order_repo.go`

- [ ] **Step 1: Add customer phone to the order model**

In `server/internal/model/order.go`, add the new field after `UserID`:

```go
CustomerPhone   string      `json:"customer_phone,omitempty"`
```

So the top of the struct becomes:

```go
type Order struct {
	ID              string      `json:"id"`
	OrderNumber     int         `json:"order_number"`
	UserID          string      `json:"user_id"`
	CustomerPhone   string      `json:"customer_phone,omitempty"`
	StoreID         string      `json:"store_id"`
	BranchID        string      `json:"branch_id"`
	Status          string      `json:"status"`
	TotalPrice      int64       `json:"total_price"`
```

- [ ] **Step 2: Load customer phone in order queries**

In `server/internal/repository/order_repo.go`, update the `orders` select statements used by `GetByID`, `ListByScope`, and `ListByUser` to join users and scan the phone:

For `GetByID`, replace the query header:

```go
	SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
	       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
	FROM orders o
	LEFT JOIN users u ON u.id = o.user_id
	WHERE o.id = $1
```

Update the scan call to include `&o.CustomerPhone` immediately after `&o.UserID`.

For `ListByScope`, update the base query:

```go
		SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
		       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.store_id = $1
```

Update the row scan in the same order, including `&o.CustomerPhone`.

For `ListByUser`, update similarly:

```go
		SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
		       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.user_id = $1 ORDER BY o.created_at DESC
```

- [ ] **Step 3: Run focused server quality**

Run:

```bash
make quality-server
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/model/order.go server/internal/repository/order_repo.go
git commit -m "feat(server): load customer phone on orders for notifications"
```

## Task 6: Add phone line to branch/store Telegram order status messages

**Files:**
- Modify: `server/internal/telegram/notifications.go`

- [ ] **Step 1: Add a phone formatter helper**

In `server/internal/telegram/notifications.go`, add:

```go
func formatCustomerPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "not provided"
	}
	return phone
}
```

Place it near the other formatting helpers.

- [ ] **Step 2: Include phone in branch/store status messages**

Update `formatBranchOrderStatusMessage` so the lines include phone after status:

```go
func formatBranchOrderStatusMessage(branchName string, order *model.Order) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("Order #%d at %s", order.OrderNumber, branchName))
	lines = append(lines, fmt.Sprintf("Status: %s", formatOrderStatusLabel(order.Status)))
	lines = append(lines, fmt.Sprintf("Phone: %s", formatCustomerPhone(order.CustomerPhone)))

	if order.RejectionReason != "" {
		lines = append(lines, fmt.Sprintf("Reason: %s", order.RejectionReason))
	}

	if len(order.Items) > 0 {
		lines = append(lines, "Items:")
		for _, item := range order.Items {
			entry := fmt.Sprintf("%dx %s", item.Quantity, item.ItemName)
			if len(item.Modifiers) > 0 {
				var mods []string
				for _, mod := range item.Modifiers {
					mods = append(mods, mod.ModifierName)
				}
				entry += fmt.Sprintf(" (%s)", strings.Join(mods, ", "))
			}
			lines = append(lines, entry)
		}
	}

	lines = append(lines, fmt.Sprintf("Total: %s UZS", formatPrice(order.TotalPrice)))
	return strings.Join(lines, "\n")
}
```

- [ ] **Step 3: Run focused server quality**

Run:

```bash
make quality-server
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/telegram/notifications.go
git commit -m "feat(telegram): include customer phone in branch status notifications"
```

## Task 7: Final cross-app verification

**Files:**
- No code changes expected

- [ ] **Step 1: Run full quality**

Run:

```bash
make quality
```

Expected: PASS across server, web, and admin.

- [ ] **Step 2: Inspect worktree**

Run:

```bash
git status --short
```

Expected: clean working tree

- [ ] **Step 3: Manual verification checklist**

Verify locally:

1. Open admin on a tablet-sized viewport
2. Open the sidebar and click a navigation link
3. Confirm the sidebar closes on both mobile-sheet and desktop-offcanvas modes
4. Open the orders board and confirm:

- columns stay in one row
- board scrolls horizontally
- cards are denser than before
- action buttons are icon-only

5. Trigger an order action and confirm:

- only that card goes into loading
- sibling buttons on that card are disabled
- other cards remain enabled

6. Trigger a branch/store order status notification and confirm the Telegram message contains:

- status
- phone line
- item list
- total

## Self-Review

### Spec coverage check

| Spec requirement | Task |
|---|---|
| close sidebar on nav click in mobile and desktop | Task 1 |
| per-card loading owned by orders page | Task 2 |
| icon-only actions, same row/position, neutral disabled style | Task 3 |
| smaller cards for tablet use | Tasks 3 and 4 |
| single-row horizontally scrolled kanban | Task 4 |
| phone on every branch/store Telegram order status notification | Tasks 5 and 6 |

### Placeholder scan

No `TODO`, `TBD`, or deferred implementation language remains. Each task has explicit file paths, code, commands, and expected outcomes.

### Type consistency

- `loading` and `loadingAction` are introduced in Task 2 and consumed in Task 3
- `CustomerPhone` is added to `model.Order` in Task 5 before being formatted in Task 6
- status action names used in Task 2 match the checks used in Task 3
