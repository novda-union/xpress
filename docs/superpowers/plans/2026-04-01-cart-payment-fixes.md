# Cart, Payment, and Auth Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

---

## HANDOFF STATUS (as of 2026-04-01)

**Tasks 1–5 DONE and committed. Task 6 is the only remaining task.**

| Task | Status | Commit |
|---|---|---|
| 1: Fix 401 token clear in api.ts | ✅ DONE | 930d38a → 92e20e3 |
| 2: Update BranchCart type | ✅ DONE | 50e0fcc |
| 3: Update cart store | ✅ DONE | 73b0b97 → c79b010 |
| 4: Update CartPage | ✅ DONE | 8c708e6 |
| 5: Cart icon in home header | ✅ DONE | eba29a3 |
| 6: Admin payment method in OrderCard | ⏳ NOT STARTED | — |

**HEAD SHA:** eba29a3

**Pick up at:** Task 6 only. Two files: `admin/types/auth.ts` and `admin/components/OrderCard.vue`.

**Known notes from code review:**
- Task 3: the `migrate` fallback branch for `version >= STORAGE_VERSION` is unreachable dead code — not a bug, acceptable as-is
- Task 4: `resolvedActiveBranchId` is typed `string | null` but passed to `setCartOptions` which expects `string` — safe at runtime due to the early-return guard, acceptable as-is
- Task 5: cart icon only shows in list view, not map view — intentional per spec

---

**Goal:** Fix four issues in the customer mini app and admin panel: cart icon in home header, 401 unauthorized when placing orders, phantom carts from stale localStorage, and a per-cart payment method selector (cash/card) visible in both web and admin.

**Architecture:** All changes are frontend-only. Types are updated first, then the Zustand store, then page components. The auth fix is in the shared `api.ts` utility. Admin changes are isolated to one type file and one component.

**Tech Stack:** React 19, TypeScript, Tailwind CSS 4, Zustand v5 (web); Nuxt 3, Vue 3, TypeScript (admin)

---

## File Map

| File | Change |
|---|---|
| `web/src/lib/api.ts` | Clear token on 401 response |
| `web/src/types/index.ts` | Add `paymentMethod` and `etaMinutes` to `BranchCart` |
| `web/src/store/cart.ts` | Bump version to 4, clear migration, add `setCartOptions`, init defaults |
| `web/src/pages/CartPage.tsx` | Payment chip row, ETA reads from store, updated order payload |
| `web/src/pages/HomePage.tsx` | Cart icon button in greeting header |
| `admin/types/auth.ts` | Add `payment_method` to `AdminOrder` |
| `admin/components/OrderCard.vue` | Display payment method in order details row |

---

## Task 1: Fix 401 — clear token in api.ts

**Files:**
- Modify: `web/src/lib/api.ts`

- [ ] **Step 1: Read current api.ts**

File is at `web/src/lib/api.ts`. The relevant block is the `!res.ok` check (lines 40–43).

- [ ] **Step 2: Add 401 token clear**

In `web/src/lib/api.ts`, update the error block inside the `api` function:

```ts
if (!res.ok) {
  if (res.status === 401) {
    localStorage.removeItem('xpressgo_token')
  }
  const error = await res.json().catch(() => ({ error: 'Request failed' }))
  throw new Error(error.error || 'Request failed')
}
```

The full updated `api` function:

```ts
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const headers = new Headers(options.headers)

  if (!headers.has('Content-Type') && options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const res = await fetch(resolveApiUrl(path), {
    ...options,
    headers,
  })

  if (!res.ok) {
    if (res.status === 401) {
      localStorage.removeItem('xpressgo_token')
    }
    const error = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new Error(error.error || 'Request failed')
  }

  if (res.status === 204) {
    return undefined as T
  }

  return res.json()
}
```

- [ ] **Step 3: Run quality check**

```bash
make quality-web
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/src/lib/api.ts
git commit -m "fix(api): clear token on 401 to trigger re-auth"
```

---

## Task 2: Update BranchCart type

**Files:**
- Modify: `web/src/types/index.ts`

- [ ] **Step 1: Update BranchCart interface**

In `web/src/types/index.ts`, find the `BranchCart` interface (currently at the bottom, after `DiscoverItem` and related types) and update it:

```ts
export interface BranchCart {
  branch: CartMeta
  items: CartItem[]
  paymentMethod: 'cash' | 'card'
  etaMinutes: number
}
```

- [ ] **Step 2: Run quality check**

```bash
make quality-web
```

Expected: TypeScript errors in `cart.ts` and `CartPage.tsx` because the new required fields don't exist yet — that is correct. Those will be fixed in the next two tasks.

- [ ] **Step 3: Commit**

```bash
git add web/src/types/index.ts
git commit -m "feat(types): add paymentMethod and etaMinutes to BranchCart"
```

---

## Task 3: Update cart store

**Files:**
- Modify: `web/src/store/cart.ts`

This task:
- Bumps `STORAGE_VERSION` from `3` to `4` (clears stale data on first load)
- Updates migration to always return empty state for versions below 4
- Adds `setCartOptions` action to the `CartStore` interface
- Initializes `paymentMethod: 'cash'` and `etaMinutes: 15` in every new `BranchCart` entry

- [ ] **Step 1: Bump version and update migration**

In `web/src/store/cart.ts`, change `STORAGE_VERSION`:

```ts
const STORAGE_VERSION = 4
```

Update the `migrate` function inside `persist(...)`:

```ts
migrate: (persistedState, version) => {
  if (version < STORAGE_VERSION) {
    return {
      carts: {},
      activeBranchId: null,
      branch: null,
      items: [],
    }
  }

  const nextState = persistedState as Partial<CartStore>
  const carts = nextState.carts ?? {}
  const activeBranchId = nextState.activeBranchId ?? Object.keys(carts)[0] ?? null

  return {
    ...nextState,
    ...withCompatState(carts, activeBranchId),
  }
},
```

- [ ] **Step 2: Add setCartOptions to CartStore interface**

In the `CartStore` interface (lines 20–39), add the new action:

```ts
interface CartStore {
  carts: Record<string, BranchCart>
  activeBranchId: string | null
  branch: CartMeta | null
  items: CartItem[]
  setActiveBranch: (branchId: string) => void
  setBranch: (branch: CartMeta) => void
  addItem: AddItem
  removeItem: RemoveItem
  updateQuantity: UpdateQuantity
  setCartOptions: (branchId: string, options: { paymentMethod?: 'cash' | 'card'; etaMinutes?: number }) => void
  clear: () => void
  clearCart: (branchId: string) => void
  clearAll: () => void
  activeBranchTotal: () => number
  activeBranchCount: () => number
  activeCart: () => BranchCart | null
  totalCartsCount: () => number
  total: () => number
  count: () => number
}
```

- [ ] **Step 3: Implement setCartOptions in the store**

After the `setBranch` action in the `create(...)` block, add `setCartOptions`:

```ts
setCartOptions: (branchId, options) => {
  set((state) => {
    const cart = state.carts[branchId]
    if (!cart) return state

    const nextCart: BranchCart = {
      ...cart,
      paymentMethod: options.paymentMethod ?? cart.paymentMethod,
      etaMinutes: options.etaMinutes ?? cart.etaMinutes,
    }

    const nextCarts = { ...state.carts, [branchId]: nextCart }
    return {
      carts: nextCarts,
      ...withCompatState(nextCarts, state.activeBranchId),
    }
  })
},
```

- [ ] **Step 4: Initialize paymentMethod and etaMinutes in addItem**

In the `addItem` implementation, when creating a new cart entry, the `BranchCart` object must include the defaults. Find the line that builds `BranchCart` inside `addItem`:

```ts
const existing = state.carts[branchMeta.branchId]
const nextItems = [...(existing?.items ?? []), item]
const nextCarts = {
  ...state.carts,
  [branchMeta.branchId]: {
    branch: branchMeta,
    items: nextItems,
    paymentMethod: existing?.paymentMethod ?? 'cash',
    etaMinutes: existing?.etaMinutes ?? 15,
  },
}
```

- [ ] **Step 5: Initialize paymentMethod and etaMinutes in setBranch**

In the `setBranch` action, the `BranchCart` object must also include the defaults:

```ts
setBranch: (branch) => {
  set((state) => {
    const existing = state.carts[branch.branchId]
    const nextCarts = {
      ...state.carts,
      [branch.branchId]: {
        branch,
        items: existing?.items ?? [],
        paymentMethod: existing?.paymentMethod ?? 'cash',
        etaMinutes: existing?.etaMinutes ?? 15,
      },
    }

    return {
      carts: nextCarts,
      activeBranchId: branch.branchId,
      ...withCompatState(nextCarts, branch.branchId),
    }
  })
},
```

- [ ] **Step 6: Run quality check**

```bash
make quality-web
```

Expected: no errors. TypeScript will confirm that `BranchCart` now satisfies the updated interface everywhere.

- [ ] **Step 7: Commit**

```bash
git add web/src/store/cart.ts
git commit -m "feat(cart): bump to v4, clear stale data, add setCartOptions, init payment+eta defaults"
```

---

## Task 4: Update CartPage — payment chips + ETA from store

**Files:**
- Modify: `web/src/pages/CartPage.tsx`

This task:
- Removes the local `eta` state (was `useState<number>(15)`)
- Reads `paymentMethod` and `etaMinutes` from the active branch cart in the store
- Adds a "Payment" chip row above the ETA section
- Updates ETA chip onClick to call `cart.setCartOptions`
- Updates the order payload to use `activeBranch.paymentMethod`

- [ ] **Step 1: Remove local eta state and read from store**

Remove the line:

```ts
const [eta, setEta] = useState<number>(15)
```

Add two derived values after `activeTotal`:

```ts
const activePaymentMethod = activeBranch?.paymentMethod ?? 'cash'
const activeEtaMinutes = activeBranch?.etaMinutes ?? 15
```

- [ ] **Step 2: Add payment chip row**

Add a payment section above the ETA section. Find the `<div className="mt-6">` block that starts with `<p className="mb-3 font-semibold">Arrive in</p>` and insert the payment section directly above it:

```tsx
<div className="mt-6">
  <p className="mb-3 font-semibold">Payment</p>
  <div className="flex gap-2">
    {(['cash', 'card'] as const).map((method) => (
      <button
        type="button"
        key={method}
        onClick={() => cart.setCartOptions(resolvedActiveBranchId, { paymentMethod: method })}
        className={`xp-pill flex shrink-0 items-center px-4 text-sm font-medium capitalize ${
          activePaymentMethod === method
            ? 'bg-[var(--xp-brand)] text-white'
            : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
        }`}
      >
        {method}
      </button>
    ))}
  </div>
</div>
```

- [ ] **Step 3: Update ETA chip row to use store**

Replace the ETA section's `onClick` and active check to use the store value:

```tsx
<div className="mt-6">
  <p className="mb-3 font-semibold">Arrive in</p>
  <div className="scrollbar-none flex gap-2 overflow-x-auto">
    {ETA_OPTIONS.map((minutes) => (
      <button
        type="button"
        key={minutes}
        onClick={() => cart.setCartOptions(resolvedActiveBranchId, { etaMinutes: minutes })}
        className={`xp-pill flex shrink-0 items-center gap-2 px-4 text-sm font-medium ${
          activeEtaMinutes === minutes
            ? 'bg-[var(--xp-brand)] text-white'
            : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
        }`}
      >
        <Clock3 className="h-4 w-4" />
        {minutes} min
      </button>
    ))}
  </div>
</div>
```

- [ ] **Step 4: Update order payload**

In the `placeOrder` function, replace the hardcoded `payment_method: 'pay_at_pickup'` with the store value, and replace `eta_minutes: eta` with `eta_minutes: activeEtaMinutes`:

```ts
body: JSON.stringify({
  branch_id: resolvedActiveBranchId,
  payment_method: activePaymentMethod,
  eta_minutes: activeEtaMinutes,
  items: activeItems.map((item) => ({
    item_id: item.itemId,
    item_name: item.name,
    item_price: item.price,
    quantity: item.quantity,
    modifiers: item.modifiers.map((modifier) => ({
      modifier_id: modifier.id,
      modifier_name: modifier.name,
      price_adjustment: modifier.price,
    })),
  })),
}),
```

- [ ] **Step 5: Run quality check**

```bash
make quality-web
```

Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add web/src/pages/CartPage.tsx
git commit -m "feat(cart): per-cart payment method selector and ETA from store"
```

---

## Task 5: Cart icon in home header

**Files:**
- Modify: `web/src/pages/HomePage.tsx`

- [ ] **Step 1: Add ShoppingBag import**

In `web/src/pages/HomePage.tsx`, add `ShoppingBag` to the lucide-react import. Currently there is no lucide import in this file — add it at the top of the imports:

```ts
import { ShoppingBag } from 'lucide-react'
```

- [ ] **Step 2: Replace greeting block with flex row**

Find the current greeting `div` (lines 157–162):

```tsx
<div className="px-4 pb-4 pt-6">
  <h1 className="text-[22px] font-bold">{greeting}</h1>
  <p className="mt-1 text-[14px] text-[var(--tg-theme-hint-color)]">
    What are you craving?
  </p>
</div>
```

Replace with:

```tsx
<div className="flex items-start justify-between px-4 pb-4 pt-6">
  <div>
    <h1 className="text-[22px] font-bold">{greeting}</h1>
    <p className="mt-1 text-[14px] text-[var(--tg-theme-hint-color)]">
      What are you craving?
    </p>
  </div>
  <button
    type="button"
    onClick={() => navigate('/cart')}
    className="relative flex h-10 w-10 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
  >
    <ShoppingBag size={22} />
    {cart.totalCartsCount() > 0 ? (
      <span className="absolute right-0 top-0 h-2.5 w-2.5 rounded-full bg-[var(--xp-brand)]" />
    ) : null}
  </button>
</div>
```

- [ ] **Step 3: Run quality check**

```bash
make quality-web
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/src/pages/HomePage.tsx
git commit -m "feat(home): add cart icon button to greeting header"
```

---

## Task 6: Admin — payment method in OrderCard

**Files:**
- Modify: `admin/types/auth.ts`
- Modify: `admin/components/OrderCard.vue`

- [ ] **Step 1: Add payment_method to AdminOrder**

In `admin/types/auth.ts`, add `payment_method` to the `AdminOrder` interface:

```ts
export interface AdminOrder {
  id: string
  order_number: number
  store_id: string
  branch_id: string
  status: string
  total_price: number
  payment_method: string
  eta_minutes: number
  created_at: string
  items: AdminOrderItem[]
}
```

- [ ] **Step 2: Display payment method in OrderCard**

In `admin/components/OrderCard.vue`, update the ETA/branch line to include payment method. Find:

```html
<p class="mb-3 text-sm text-muted-foreground">
  ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }}
</p>
```

Replace with:

```html
<p class="mb-3 text-sm text-muted-foreground">
  ETA ~{{ order.eta_minutes }} min · Branch {{ order.branch_id.slice(0, 8) }} · {{ formatPayment(order.payment_method) }}
</p>
```

Add the `formatPayment` helper to the `<script setup>` block (after the existing `formatTime` function):

```ts
function formatPayment(method: string) {
  if (!method) return ''
  if (method === 'cash') return 'Cash'
  if (method === 'card') return 'Card'
  return method.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}
```

- [ ] **Step 3: Run quality check**

```bash
make quality-admin
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add admin/types/auth.ts admin/components/OrderCard.vue
git commit -m "feat(admin): show payment method on order card"
```

---

## Self-Review

**Spec coverage:**
- [x] Cart icon in home header (Task 5)
- [x] 401 unauthorized fix (Task 1)
- [x] Stale cart data cleared (Task 3 — version bump + migration reset)
- [x] Payment method selector per cart on CartPage (Task 4)
- [x] ETA moved to per-cart in store (Task 3 + 4)
- [x] Admin OrderCard shows payment method (Task 6)
- [x] `BranchCart` type updated with new fields (Task 2)
- [x] `AdminOrder` type updated (Task 6)

**Type consistency:**
- `BranchCart.paymentMethod: 'cash' | 'card'` — defined in Task 2, used in Tasks 3 and 4 consistently
- `BranchCart.etaMinutes: number` — defined in Task 2, used in Tasks 3 and 4 consistently
- `setCartOptions(branchId, { paymentMethod?, etaMinutes? })` — defined in Task 3 interface, implemented in Task 3, called in Task 4 with matching signature
- `AdminOrder.payment_method: string` — defined in Task 6, used in Task 6 `OrderCard.vue`

**No placeholders found.**
