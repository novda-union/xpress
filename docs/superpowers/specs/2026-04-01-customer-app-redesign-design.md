# Customer Mini App Redesign

**Date:** 2026-04-01
**Scope:** Web mini app (`web/`)
**Status:** Approved for implementation

## Overview

Redesign the Xpressgo customer Telegram mini app from a branch-card discovery model to an item-first feed experience similar to Yandex Food / Delivery Club. The goal is an engaging, appetite-driven UI where food and drinks are the primary visual content from the first screen.

This spec also covers:
- Multi-cart architecture (one cart per branch, placed as separate orders)
- Item detail page redesign with back button and dedicated API endpoint
- Bug fixes for loading states, z-index conflicts, and branch page rendering
- New backend endpoints for item-centric discovery

---

## 1. Home / Discovery Page

### Layout (top to bottom)

**Greeting header** (not sticky, scrolls away)
- Line 1: "Good evening, [first_name]" — 22px bold
- Line 2: "What are you craving?" — 14px muted
- First name sourced from `AuthUser.first_name` via `useTelegramAuth()`
- Falls back to "there" if no name available

**Smart chip row** (sticky on scroll, `z-10`)
- Single horizontally scrollable row, no wrapping
- Chips in order: `[New] [Popular] [Drinks] [Food] [Bites] [Bars] [Cafes] [Coffee]`
- First two (`New`, `Popular`) are sort-intent chips — filter the vertical feed by `created_at desc` or `order_count desc`
- Rest are item-type / store-category filters
- Default active: `New`
- Active chip: filled brand color (`--xp-brand`), white text
- Inactive chip: `bg-(--xp-card-bg)`, muted text
- Chip height: 32px, horizontal padding: 16px, border-radius: full

**"New Arrivals" section**
- Section header: "New Arrivals" (15px semibold) + "See all" link (13px brand color) on the right
- Horizontal scroll row of item cards (160px wide, see Item Card spec)
- Items: sorted by `created_at desc`, limit 10
- Data source: `GET /discover/feed` → section `type: "new"`

**"Popular Right Now" section**
- Same structure as New Arrivals
- Items: sorted by `order_count desc`, limit 10
- Data source: `GET /discover/feed` → section `type: "popular"`

**Vertical item feed**
- 2-column grid, `gap-3`, `px-4`
- Filtered and sorted by active chip selection
- Paginated: 20 items per page, infinite scroll (load next page when last item enters viewport — use IntersectionObserver)
- Data source: `GET /discover/items?type=&sort=new|popular&page=&limit=20`
- Shows skeleton cards while loading first page

### Skeleton loading
- Greeting header: two lines of `animate-pulse` rounded rectangles
- Chip row: 5 pill-shaped skeletons
- Section cards: 2 square card skeletons per section
- Vertical feed: 4 card skeletons (2×2 grid)

### Map view
- Map/list toggle (`ViewToggle`) remains accessible — `z-20`
- Map view unchanged from current implementation (tile already updated)
- `BranchSheet` stays at `z-30` — toggle always renders below sheet

---

## 2. Item Card

Used in both horizontal section rows and the vertical grid feed.

### Dimensions
- Horizontal section card: `w-[160px]`, `flex-shrink-0`
- Vertical grid card: `w-full` (fills 50% column)

### Structure
```
┌─────────────────┐
│                 │  ← square image, aspect-ratio: 1/1
│  [NEW]          │  ← badge, top-left, conditional
│             [+] │  ← action button, bottom-right
├─────────────────┤
│ Item name       │  ← 13px semibold, 2-line clamp
│ 45,000 UZS      │  ← brand color, 13px
│ Skybar · 1.2 km │  ← muted, 11px
└─────────────────┘
```

### NEW badge
- Shown only if item `created_at` is within the last 7 days
- Style: `bg-emerald-500 text-white text-[10px] font-semibold px-2 py-0.5 rounded-full`
- Position: `absolute top-2 left-2`

### `+` / Quantity button behavior
- **If item has no required modifier groups**: tapping `+` adds item to cart directly (no navigation). Button morphs into inline stepper `[− N +]`.
- **If item has required modifier groups**: tapping `+` navigates to `/item/:id?branch=branchId`.
- Determination: check `item.has_required_modifiers` (new boolean field on `DiscoverItem`).

### Stepper morph animation
- `+` circle expands to `[− N +]` pill with `width` transition, 200ms ease
- Wrap stepper in a container with `overflow: hidden` and transition `width` (not `max-width`) for reliable flex behavior: collapsed `w-9`, expanded `w-24`
- `-` at zero: removes item from cart, morphs back to `+`

### Add-to-cart animation
- On tap: `scale(1.0 → 1.2 → 1.0)` on the button, 150ms ease-out
- Use `transform` only (GPU-accelerated, no layout shift)

### Branch conflict behavior
Three cases when tapping `+` on an item:

1. **Cart is empty or item's branch matches `activeBranchId`**: add directly, no dialog.
2. **Item's branch already has a cart (but is not active)**: add directly to that branch's existing cart, switch `activeBranchId` to it. No dialog — user is continuing an existing cart.
3. **Item's branch is new (not in `carts` at all) and other carts exist**: show `BranchConflictSheet`. On confirm: create new cart for this branch, add item, set `activeBranchId`.

After any add: set `activeBranchId` to item's branch.

### Image
- `object-fit: cover`, `rounded-xl`
- Fallback: `https://placehold.co/300x300?text=Item` (square placeholder)
- Lazy loading: `loading="lazy"` on `<img>`
- Reserve space with `aspect-ratio: 1/1` to prevent layout shift

---

## 3. Item Detail Page

### Back button
- `position: absolute`, `top: 16px`, `left: 16px`, `z-30`
- Style: `bg-black/40 backdrop-blur-sm` pill, white `ArrowLeft` icon (Lucide, 20px)
- Touch target: `min-h-[44px] min-w-[44px]`
- Action: `navigate(-1)`

### Hero image
- `aspect-ratio: 4/3`, `w-full`, `object-fit: cover`
- Gradient overlay at bottom: `bg-gradient-to-t from-(--tg-theme-bg-color) to-transparent`, height 80px
- Fallback placeholder maintains aspect ratio

### Sticky header
- Appears when hero image scrolls out of viewport (use IntersectionObserver on image)
- Contains: back button (left) + item name truncated (center) + cart icon with count (right)
- Fade-in: `opacity 0 → 1`, 150ms ease
- `z-10`, `bg-(--tg-theme-bg-color)/90 backdrop-blur-sm`

### Content below image
- Item name: 22px bold
- Base price: 20px semibold, brand color
- Description: 15px, 1.6 line-height, muted color
- Modifier groups: see below

### Modifier groups
- Group header: group name (15px semibold) + "Required" label in red if `is_required`
- Each modifier: pill/chip
  - Unselected: `border border-(--xp-border) bg-transparent`
  - Selected: `bg-(--xp-brand) text-white border-transparent`
  - Price adjustment shown inline: `+2,000` in smaller text if `price_adjustment > 0`
  - Touch target: `min-h-[44px]`

### Bottom action bar
- Fixed, `inset-x-0 bottom-0`, `z-20`, `pb-safe`
- Left: quantity stepper `[− N +]` in rounded pill
- Right: "Add to Cart · X UZS" button, flex-1, brand color, `rounded-[20px]`

### Loading fix
- Replace full menu load with `GET /items/:id?branch=branchId`
- Show skeleton layout (image placeholder + text lines) while loading instead of "Loading..." text

---

## 4. Multi-Cart Architecture

### Cart store shape

```ts
interface CartStore {
  carts: Record<string, BranchCart>  // keyed by branchId
  activeBranchId: string | null

  // actions
  setActiveBranch(branchId: string): void
  addItem(branchMeta: CartMeta, item: CartItem): void
  removeItem(branchId: string, index: number): void
  updateQuantity(branchId: string, index: number, qty: number): void
  clearCart(branchId: string): void
  clearAll(): void
  total(branchId: string): number
  count(branchId: string): number
  activeCart(): BranchCart | null
  activeBranchTotal(): number
  activeBranchCount(): number
  totalCartsCount(): number  // number of branches with items
}

interface BranchCart {
  branch: CartMeta
  items: CartItem[]
}
```

localStorage key: `xpressgo-cart-v3` (version bump from v2 to avoid stale shape conflicts).

### Branch conflict sheet (`BranchConflictSheet`)

Slides up from bottom (`z-40`) when user taps `+` on an item from a branch not in `carts`.
- Title: "Add from [Branch Name]?"
- Body: "You have items from [existing branch name]. They'll stay in a separate cart."
- Buttons: "Add to new cart" (brand color, primary) + "Cancel" (ghost)
- On confirm: add item, set `activeBranchId` to new branch

### Cart bar

- Fixed bottom, `z-20`, slides up with `translateY(100% → 0)` (300ms ease-out) when `activeBranchCount() > 0`
- Shows: item count + total for `activeBranchId` cart
- If `totalCartsCount() > 1`: small badge top-right — "[N] carts"
- Count change animation: number scale-pulse 150ms

### Cart page (`/cart`)

- Tabs at top: one tab per branch in `carts` — tab label is `branch.branchName`
- Active tab underline with brand color
- Tab content: items list for that branch, quantity controls, ETA selector, Place Order button
- Placing an order only submits the active tab's cart
- After successful order: `clearCart(activeBranchId)`, navigate to `/order/:id`
- If only one cart remains after clearing: `activeBranchId` updates to the remaining one
- If no carts remain: show empty state

---

## 5. Z-Index Scale

Define globally in `index.css`:

```css
:root {
  --z-sticky: 10;
  --z-floating: 20;
  --z-sheet: 30;
  --z-modal: 40;
  --z-toast: 50;
}
```

| Layer | Value | Elements |
|---|---|---|
| sticky | 10 | Chip row, category tabs, sticky item page header |
| floating | 20 | CartBar, ViewToggle |
| sheet | 30 | BranchSheet, item detail back button overlay |
| modal | 40 | BranchConflictSheet, confirmation dialogs |
| toast | 50 | Toast notifications |

---

## 6. Bug Fixes

### ViewToggle over BranchSheet
- `ViewToggle`: change to `z-20`
- `BranchSheet` outer container: change to `z-30`
- Sheet overlay backdrop: `z-30` (same stacking context)

### ItemPage stuck on loading
- Add `GET /items/:id?branch=branchId` backend endpoint
- Returns: `MenuItem` with `modifier_groups` + `BranchDetail` in one response (or two separate fields)
- Remove the full menu load from `ItemPage.tsx`

### OrderPage stuck on loading
- Replace `if (!order) return <Loading>` with a skeleton layout
- Investigate WebSocket auth: if token is missing, `useWebSocket` silently fails — add error logging

### BranchPage category tabs / items not visible
- Investigate: `activeCategory` defaults to `''` before menu loads; verify `currentCategory` fallback reaches `menu.categories[0]`
- Check sticky `top-12` value — if `MenuHeader` height changes, tabs may be hidden behind header
- Add skeleton cards while menu loads

---

## 7. New Backend Endpoints

### `GET /discover/feed`

Returns pre-built curated sections.

```json
{
  "sections": [
    {
      "title": "New Arrivals",
      "type": "new",
      "items": [DiscoverItem]
    },
    {
      "title": "Popular Right Now",
      "type": "popular",
      "items": [DiscoverItem]
    }
  ]
}
```

Limit: 10 items per section. No pagination on this endpoint.

### `GET /discover/items`

Query params: `type` (item category name, optional), `sort` (`new` | `popular`, default `new`), `page` (default 1), `limit` (default 20).

Returns:
```json
{
  "items": [DiscoverItem],
  "total": 120,
  "page": 1,
  "limit": 20
}
```

### `GET /items/:id?branch=branchId`

Returns single item with modifier groups + branch context.

```json
{
  "item": MenuItem,
  "branch": BranchDetail
}
```

### `DiscoverItem` type

Extends `BranchPreviewItem` with:

```ts
interface DiscoverItem {
  id: string
  name: string
  description: string
  image_url: string
  base_price: number
  is_available: boolean
  created_at: string
  order_count: number
  has_required_modifiers: boolean  // determines + button behavior
  branch_id: string
  branch_name: string
  branch_address: string
  store_id: string
  store_name: string
  store_category: string
  lat?: number
  lng?: number
}
```

---

## 8. Animation Summary

| Interaction | Spec |
|---|---|
| `+` button tap | `scale(1 → 1.2 → 1)`, 150ms ease-out, `transform` only |
| Stepper morph | `max-width` expand, 200ms ease |
| Cart bar appear | `translateY(100% → 0)`, 300ms ease-out |
| Cart bar count update | scale-pulse on number, 150ms |
| Bottom sheet open | `translateY(100% → 0)`, 300ms ease-out |
| Sticky header fade-in | `opacity 0 → 1`, 150ms |
| Skeleton screens | `animate-pulse` on all loading states |
| Page-level transitions | Telegram SDK handles; no custom page transitions |

All animations use `transform` or `opacity` only — no `width`, `height`, or `top/left` transitions.
Respect `prefers-reduced-motion`: wrap all non-essential animations in `@media (prefers-reduced-motion: no-preference)`.

---

## 9. Files Affected

### New files
- `web/src/components/discovery/ItemCard.tsx` (replaces `BranchListCard`)
- `web/src/components/cart/BranchConflictSheet.tsx`
- `web/src/hooks/useDiscoveryFeed.ts`
- `web/src/hooks/useDiscoveryItems.ts`
- `server/internal/handler/item_handler.go`

### Modified files
- `web/src/pages/HomePage.tsx` — full rewrite
- `web/src/pages/ItemPage.tsx` — back button, hero image, sticky header, new endpoint
- `web/src/pages/CartPage.tsx` — multi-cart tabs
- `web/src/pages/BranchPage.tsx` — loading skeleton, category fix
- `web/src/pages/OrderPage.tsx` — loading skeleton
- `web/src/store/cart.ts` — multi-cart shape, localStorage v3
- `web/src/components/discovery/ViewToggle.tsx` — z-index fix
- `web/src/components/discovery/BranchSheet.tsx` — z-index fix
- `web/src/components/cart/CartBar.tsx` — multi-cart count + badge
- `web/src/types/index.ts` — add `DiscoverItem`, update `CartStore`
- `server/internal/handler/router.go` — new routes
- `server/internal/repository/item_repo.go` — new queries
- `server/internal/handler/discover_handler.go` — feed + items endpoints
