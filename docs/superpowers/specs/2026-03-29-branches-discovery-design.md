# Xpressgo — Branches, Discovery & Map Design Spec

## Overview

This spec covers three major additions to the Xpressgo platform:

1. **Branch model** — stores now have one or more branches. Orders go to specific branches. Each branch has its own menu, coordinates, and staff.
2. **Platform discovery** — users can browse all stores/branches on the platform via a map view and a section-divided list view, directly within the Telegram Mini App.
3. **Phone authentication gate** — users must verify their phone number via Telegram's native `requestContact()` before accessing the app.

This also includes a full UI polish pass (handled separately by the ui-ux-pro-max skill) and admin panel upgrades for branch management, staff management, and role-based permissions.

---

## Data Model Changes

### New Table: `branches`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| store_id | UUID | FK → stores |
| name | VARCHAR | e.g. "Branch - Chilonzor" |
| address | VARCHAR | Human-readable address |
| lat | DECIMAL(10,8) | Latitude |
| lng | DECIMAL(11,8) | Longitude |
| banner_image_url | VARCHAR | Shown in map card and list card |
| telegram_group_chat_id | BIGINT | For bot order alerts (per branch) |
| is_active | BOOLEAN | |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

### Modified Tables

**`stores`** — add `category` VARCHAR. Values: `bar`, `cafe`, `restaurant`, `coffee`, `fastfood`. Used for category tab filtering in list view.

**`categories`** — add `branch_id` UUID FK → branches. Keep `store_id` (denormalized for security middleware). Remove old store-only constraint.

**`items`** — add `branch_id` UUID FK → branches. Keep `store_id` (denormalized).

**`modifier_groups`** — add `branch_id` UUID FK → branches. Keep `store_id` (denormalized).

**`modifiers`** — add `branch_id` UUID FK → branches. Keep `store_id` (denormalized).

**`orders`** — add `branch_id` UUID FK → branches.

**`store_staff`** — add `branch_id` UUID FK → branches (nullable — null = Director). Update role enum to: `director`, `manager`, `barista`.

### Seed Data Update

Demo Bar gets one branch: "Demo Bar - Main" with coordinates pointing to central Tashkent (41.2995, 69.2401).

---

## Roles & Permissions

### Role Hierarchy

| Role | Scope | Created By |
|------|-------|-----------|
| `director` | Store-level (all branches) | System only (us, via seed/internal endpoint) |
| `manager` | Branch-level | Director |
| `barista` | Branch-level | Director or Manager |

### Permission Matrix

| Action | Director | Manager | Barista |
|--------|----------|---------|---------|
| `branch:create` | ✓ | ✗ | ✗ |
| `branch:edit` | ✓ | ✗ | ✗ |
| `branch:delete` | ✓ | ✗ | ✗ |
| `staff:create:manager` | ✓ | ✗ | ✗ |
| `staff:create:barista` | ✓ | ✓ | ✗ |
| `staff:edit` | ✓ | ✓ (baristas only) | ✗ |
| `menu:manage` | ✓ | ✓ | ✗ |
| `settings:store` | ✓ | ✗ | ✗ |
| `settings:branch` | ✓ | ✓ | ✗ |
| `orders:view:all` | ✓ | ✗ | ✗ |
| `orders:view` | ✓ | ✓ | ✓ |
| `dashboard:all` | ✓ | ✗ | ✗ |
| `dashboard:branch` | ✓ | ✓ | ✓ |

### JWT Changes

- Director JWT: `{ store_id, staff_id, role: "director", branch_id: null }`
- Manager/Barista JWT: `{ store_id, staff_id, role: "manager"|"barista", branch_id: "uuid" }`

Admin middleware: if `branch_id` is null → scope by `store_id` only. If `branch_id` is set → scope all queries to that branch.

---

## API Changes

### New Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/discover` | All active branches with coords, store name, category, logo. Supports `?category=bar` filter. |
| GET | `/branches/:id` | Single branch info + parent store details |
| GET | `/branches/:id/menu` | Full menu for that branch (categories → items → modifiers) |

### Modified Public Endpoints

- `POST /orders` — body now includes `branch_id` (replaces store slug scoping)

### New Admin Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/branches` | List branches. Director sees all, others see their own. |
| POST | `/admin/branches` | Create branch (Director only) |
| PUT | `/admin/branches/:id` | Update branch info + coordinates (Director only) |
| DELETE | `/admin/branches/:id` | Deactivate branch (Director only) |
| GET | `/admin/staff` | List staff. Director sees all grouped by branch, Manager sees baristas in their branch. |
| POST | `/admin/staff` | Create staff with role + branch assignment |
| PUT | `/admin/staff/:id` | Update staff (role, branch, password reset) |
| DELETE | `/admin/staff/:id` | Deactivate staff |

### Updated Admin Menu Endpoints

All existing menu endpoints (categories, items, modifier groups, modifiers) now accept and validate `branch_id` in the request body. Middleware scopes queries to `branch_id` from JWT when set.

---

## Mini App: UI & Flow

### Phone Authentication Gate

Runs before anything else on first open:

1. App checks for existing JWT in localStorage
2. If none → full-screen gate: brief explanation of why phone is needed + "Allow Access" button
3. Tap Allow → Telegram native `requestContact()` popup
4. User approves → phone sent to `POST /auth/telegram` alongside `initData` → JWT issued → proceed
5. User taps Deny → gate stays, shows message: *"Phone number is required to place orders"* with retry button — hard block, no browsing without auth
6. Subsequent opens → auto-auth via `initData`, no repeat prompt

### Home Screen (`/`)

Single route with two views toggled by a top button: **Map** | **List**.

The map component stays mounted in the DOM when List view is active (hidden via CSS), preventing re-initialization and saving MapLibre load counts.

On first render, app requests user location via Telegram SDK. Falls back to Tashkent city center (41.2995, 69.2401) if denied.

### Map View

**Technology:** MapLibre GL JS + Maptiler free tier tiles.

**Markers:**
- Each active branch = circular marker with `stores.logo_url` (falls back to a default icon if null)
- White border, subtle drop shadow
- Scale + fade animation on first render

**Branch card (bottom sheet):**
- Triggered by tapping a marker
- Slides up from bottom with spring animation
- Contains: store name, branch name, address, `branches.banner_image_url`
- Horizontal scrollable carousel of up to 5 menu items (item image, name, price)
- Each carousel item tappable → navigates to `/item/:id`
- "See Full Menu" button → navigates to `/branch/:id`
- Tap outside or drag down → dismisses card

### List View

**Structure:**
- Category tabs at top (horizontally scrollable): All · Bars · Cafes · Coffee · Restaurants · Fast Food
- Branch cards below, sorted by proximity (nearest first within selected category)
- Each card: store logo, store name, branch name, address, category badge, estimated distance

**Tap card** → navigates to `/branch/:id`

### Branch Menu Page (`/branch/:id`)

Same layout as current `StorePage` — category tabs + item grid — but scoped to `branch_id`. Item cards are tappable links to `/item/:id` (replaces the current modal pattern).

### Item Detail Page (`/item/:id`) — New

Full-screen page:
- Hero image at top
- Item name, description, base price
- Modifier groups (same radio/checkbox logic, migrated from modal)
- "Add to Cart" button pinned at bottom
- Back navigation returns to previous screen

### Cart & Order Flow

No UX change. Cart stores `branch_id` instead of store slug. Order submitted with `branch_id`. Status tracking and WebSocket flow unchanged.

### Bot Deep-link Behavior

`/start` with store slug:
- Store has one branch → go directly to that branch's menu
- Store has multiple branches → show branch picker screen before menu

---

## Admin Panel Changes

### Route Guards

**`auth.global.ts` middleware** — extended with role checks:

1. No JWT → redirect to `/login`
2. Insufficient role for route → redirect to `/` with permission toast

| Route | Director | Manager | Barista |
|-------|----------|---------|---------|
| `/` | ✓ | ✓ | ✓ |
| `/orders` | ✓ | ✓ | ✓ |
| `/menu` | ✓ | ✓ | ✗ |
| `/branches` | ✓ | ✗ | ✗ |
| `/staff` | ✓ | ✓ | ✗ |
| `/settings` | ✓ | ✗ | ✗ |
| `/settings/branch` | ✓ | ✓ | ✗ |

### `usePermissions()` Composable

Central permission utility. Reads role from JWT. Exposes `can(action: string): boolean`.

Used in two layers:
- **UI layer** — `v-if="can('branch:create')"` hides unauthorized buttons/forms
- **Operation layer** — `can()` guard before any API call, shows toast on failure

Backend enforces the same rules independently (JWT role validation in Go middleware).

### New Page: Branches (`/branches`)

- Director only
- Lists all branches with name, address, status, staff count
- "Add Branch" form:
  - Name, address fields
  - Embedded MapLibre map with draggable pin — lat/lng auto-filled from pin
  - Telegram group chat ID field
  - Active toggle
- Edit/deactivate existing branches

### New Page: Staff (`/staff`)

- Director: sees all staff grouped by branch
- Manager: sees only baristas in their branch
- "Add Staff" form:
  - Name, staff code, password, role selector, branch selector
  - Director: role options = Manager, Barista; branch = any
  - Manager: role = Barista only; branch = auto-set to their branch (selector hidden)
- Edit staff: update name, reset password, change branch assignment (Director only)

### Updated: Header

Director sees a branch switcher dropdown in the top navigation:
- Options: individual branches + "All Branches"
- Selection persists in localStorage
- Affects orders page, dashboard, and menu page context

### Updated: Dashboard

- Director + "All Branches": aggregate stats (total orders today, total revenue) + per-branch breakdown cards
- Director + specific branch selected: that branch's stats only
- Manager/Barista: their branch stats only

### Updated: Orders Page

- Director + "All Branches": all orders across branches, each card tagged with branch name
- Director + specific branch: that branch's kanban
- Manager/Barista: their branch kanban only

### Updated: Menu Page

- Director must select a specific branch before managing menu (inline gate if "All Branches" is selected: "Select a branch to manage its menu")
- Manager: goes directly to their branch menu

### Updated: Settings

- `/settings` → store-level (name, logo, category, commission rate) — Director only
- `/settings/branch` → branch-level (address, coords, telegram chat ID) — Director + Manager

---

## Deferred

- Branch-level menu overrides (price/availability per branch)
- Branch analytics (per-branch revenue comparison)
- Self-service Director onboarding
