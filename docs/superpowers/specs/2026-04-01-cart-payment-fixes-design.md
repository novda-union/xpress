# Cart, Payment, and Auth Fixes

**Date:** 2026-04-01
**Scope:** Web mini app (`web/`), Admin panel (`admin/`)
**Status:** Approved for implementation

## Overview

Four focused fixes to the customer mini app and admin panel:

1. Cart icon in home page header
2. Unauthorized error when placing an order
3. Stale cart data showing phantom carts
4. Payment method selector (cash / card) — per cart, no payment processing

---

## 1. Cart Icon in Home Header

### Location

Right side of the greeting block in `web/src/pages/HomePage.tsx`.

Currently the greeting is:
```
h1: "Good evening"      [nothing on the right]
p:  "What are you craving?"
```

After:
```
h1: "Good evening"      [ShoppingBag icon button]
p:  "What are you craving?"
```

### Spec

- Icon: `ShoppingBag` from Lucide React, size 22px
- Button: `w-10 h-10`, rounded-full, `bg-(--xp-card-bg)`, centered icon
- Badge dot: shown when `cart.totalCartsCount() > 0`
  - Style: `absolute top-0 right-0`, `w-2 h-2 rounded-full bg-(--xp-brand)`
- Tap action: `navigate('/cart')`
- Always visible (not conditional on cart count) — gives users a persistent path to cart

### Layout change

The greeting `div` becomes a flex row: `flex items-start justify-between`:
- Left: existing `h1` + `p` stack
- Right: icon button with optional badge

---

## 2. Unauthorized Fix

### Root cause

`useTelegramAuth` skips authentication if `xpressgo_token` exists in localStorage:

```ts
const [isAuthenticated, setIsAuthenticated] = useState(Boolean(localStorage.getItem('xpressgo_token')))
// ...
if (isAuthenticated) return  // ← skips re-auth even if token is expired
```

An expired token stays in localStorage indefinitely. Any authenticated request returns 401, which surfaces as an alert in CartPage.

### Fix

In `web/src/lib/api.ts`, after checking `!res.ok`:

```ts
if (res.status === 401) {
  localStorage.removeItem('xpressgo_token')
}
```

This clears the stale token on any 401 response. On the next render cycle, `useTelegramAuth`'s `isAuthenticated` becomes `false`, the effect re-runs, and the app re-authenticates using Telegram initData or dev fallback. No page reload required.

---

## 3. Stale Cart Data

### Root cause

The Zustand persist migration (version 2 → 3) carries over old branch + items from localStorage. Test/dev sessions with data in `xpressgo-cart-v2` or earlier `v3` keys surface as phantom carts in the UI.

### Fix

Bump `STORAGE_VERSION` from `3` to `4` in `web/src/store/cart.ts`.

In the `migrate` function, when `version < 4`, always return clean empty state:

```ts
if (version < STORAGE_VERSION) {
  return {
    carts: {},
    activeBranchId: null,
    branch: null,
    items: [],
  }
}
```

This wipes stale data for all existing sessions on first load after the update. The localStorage key remains `xpressgo-cart-v3` (no need to change the key, only the version number inside the persisted payload).

---

## 4. Payment Method Selector

### Data model changes

`BranchCart` in `web/src/types/index.ts` gains two new fields:

```ts
interface BranchCart {
  branch: CartMeta
  items: CartItem[]
  paymentMethod: 'cash' | 'card'   // default: 'cash'
  etaMinutes: number                // default: 15 (moved from CartPage local state)
}
```

`etaMinutes` moves from CartPage local `useState` into the cart store so each branch cart remembers its own ETA independently.

### Store changes (`web/src/store/cart.ts`)

New action:

```ts
setCartOptions(branchId: string, options: { paymentMethod?: 'cash' | 'card'; etaMinutes?: number }): void
```

When `addItem` creates a new `BranchCart` entry, initialize with `paymentMethod: 'cash'` and `etaMinutes: 15`.

### CartPage UI

**Payment method section** — placed above the ETA section:

```
Payment
[Cash]  [Card]
```

- Chip style: identical to ETA chips (`xp-pill`, `px-4`, `text-sm font-medium`)
- Active: `bg-(--xp-brand) text-white`
- Inactive: `bg-(--xp-card-bg) text-(--tg-theme-hint-color)`
- Chips: "Cash" and "Card"
- Tapping a chip calls `cart.setCartOptions(branchId, { paymentMethod: value })`

**ETA section** — reads `activeBranch.etaMinutes` from the store instead of local state. Tapping an ETA chip calls `cart.setCartOptions(branchId, { etaMinutes: minutes })`.

**Order payload** — `payment_method` sent as `'cash'` or `'card'` (replaces hardcoded `'pay_at_pickup'`).

### Admin OrderCard (`admin/components/OrderCard.vue`)

Add payment method display to the existing ETA line:

```
ETA ~15 min · Branch abc12345 · Cash
```

- Source: `order.payment_method`
- Format: capitalize first letter (`cash` → `Cash`, `card` → `Card`)
- No icon needed — text label only

The `AdminOrder` type in `admin/types/auth.ts` gains `payment_method: string`.

---

## 5. Files Affected

### Modified

- `web/src/pages/HomePage.tsx` — add cart icon button to greeting header
- `web/src/lib/api.ts` — clear token on 401
- `web/src/store/cart.ts` — bump version to 4, migration clears all, add `setCartOptions`, init `paymentMethod` + `etaMinutes` per cart
- `web/src/types/index.ts` — add `paymentMethod` and `etaMinutes` to `BranchCart`
- `web/src/pages/CartPage.tsx` — payment chip row, ETA reads from store, updated order payload
- `admin/components/OrderCard.vue` — display payment method in order details
- `admin/types/auth.ts` — add `payment_method` to `AdminOrder`
