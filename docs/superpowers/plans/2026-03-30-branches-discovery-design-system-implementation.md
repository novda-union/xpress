# Branches, Discovery, and Design System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add branch-aware data and permissions, ship public branch discovery with phone-gated ordering, and roll out the full mini app and admin UI/UX design system without breaking the current order flow.

**Architecture:** Start by introducing `branches` as the new operational scope while keeping `store_id` as the tenant boundary. Once the backend can serve branch-scoped menu, discovery, orders, staff, and permissions, replace the slug-based mini app flow with discovery-first navigation and rebuild the admin shell around role-aware branch management. The design system should be implemented as shared tokens and layout primitives in each frontend before feature pages are rebuilt.

**Tech Stack:** Go + Echo + pgx + PostgreSQL, React 19 + Vite + Tailwind 4 + Zustand, Nuxt 3 + Tailwind, Telegram Mini App SDK, MapLibre GL JS, Lucide icons.

---

## File Structure Map

### Server

- Create: `server/migrations/000002_branches_and_permissions.up.sql`
- Create: `server/migrations/000002_branches_and_permissions.down.sql`
- Create: `server/internal/model/branch.go`
- Create: `server/internal/repository/branch_repo.go`
- Create: `server/internal/service/permission_service.go`
- Create: `server/internal/handler/branch_handler.go`
- Create: `server/internal/handler/staff_handler.go`
- Modify: `server/cmd/seed/main.go`
- Modify: `server/internal/middleware/auth.go`
- Modify: `server/internal/model/store.go`
- Modify: `server/internal/model/staff.go`
- Modify: `server/internal/model/category.go`
- Modify: `server/internal/model/item.go`
- Modify: `server/internal/model/modifier.go`
- Modify: `server/internal/model/order.go`
- Modify: `server/internal/repository/store_repo.go`
- Modify: `server/internal/repository/menu_repo.go`
- Modify: `server/internal/repository/staff_repo.go`
- Modify: `server/internal/repository/order_repo.go`
- Modify: `server/internal/repository/item_repo.go`
- Modify: `server/internal/repository/category_repo.go`
- Modify: `server/internal/repository/modifier_repo.go`
- Modify: `server/internal/service/auth_service.go`
- Modify: `server/internal/service/order_service.go`
- Modify: `server/internal/handler/auth_handler.go`
- Modify: `server/internal/handler/store_handler.go`
- Modify: `server/internal/handler/menu_handler.go`
- Modify: `server/internal/handler/order_handler.go`
- Modify: `server/internal/handler/router.go`
- Modify: `server/internal/telegram/notifications.go`
- Modify: `server/internal/ws/hub.go`

### Mini App

- Create: `web/src/lib/telegram.ts`
- Create: `web/src/lib/theme.ts`
- Create: `web/src/lib/location.ts`
- Create: `web/src/lib/distance.ts`
- Create: `web/src/components/layout/AppShell.tsx`
- Create: `web/src/components/auth/PhoneGate.tsx`
- Create: `web/src/components/discovery/ViewToggle.tsx`
- Create: `web/src/components/discovery/CategoryTabs.tsx`
- Create: `web/src/components/discovery/DiscoveryMap.tsx`
- Create: `web/src/components/discovery/BranchMarker.tsx`
- Create: `web/src/components/discovery/BranchSheet.tsx`
- Create: `web/src/components/discovery/BranchListCard.tsx`
- Create: `web/src/components/menu/MenuHeader.tsx`
- Create: `web/src/components/menu/ItemCard.tsx`
- Create: `web/src/components/menu/ModifierGroupSelector.tsx`
- Create: `web/src/components/cart/CartBar.tsx`
- Create: `web/src/components/common/StatusBadge.tsx`
- Create: `web/src/hooks/useTelegramAuth.ts`
- Create: `web/src/hooks/useTelegramTheme.ts`
- Create: `web/src/hooks/useUserLocation.ts`
- Create: `web/src/hooks/useDiscovery.ts`
- Create: `web/src/pages/HomePage.tsx`
- Create: `web/src/pages/BranchPage.tsx`
- Create: `web/src/pages/ItemPage.tsx`
- Modify: `web/package.json`
- Modify: `web/src/App.tsx`
- Modify: `web/src/index.css`
- Modify: `web/src/lib/api.ts`
- Modify: `web/src/store/cart.ts`
- Modify: `web/src/types/index.ts`
- Modify: `web/src/pages/CartPage.tsx`
- Modify: `web/src/pages/OrderPage.tsx`
- Modify: `web/src/pages/OrdersPage.tsx`
- Delete or retire usage from: `web/src/pages/StorePage.tsx`

### Admin

- Create: `admin/assets/css/main.css`
- Create: `admin/types/auth.ts`
- Create: `admin/composables/usePermissions.ts`
- Create: `admin/composables/useBranchContext.ts`
- Create: `admin/components/layout/AppSidebar.vue`
- Create: `admin/components/layout/AppHeader.vue`
- Create: `admin/components/ui/StatCard.vue`
- Create: `admin/components/ui/EmptyState.vue`
- Create: `admin/components/branches/BranchForm.vue`
- Create: `admin/components/branches/BranchTable.vue`
- Create: `admin/components/staff/StaffForm.vue`
- Create: `admin/components/staff/StaffGroupList.vue`
- Create: `admin/pages/branches.vue`
- Create: `admin/pages/staff.vue`
- Create: `admin/pages/settings/branch.vue`
- Modify: `admin/package.json`
- Modify: `admin/nuxt.config.ts`
- Modify: `admin/tailwind.config.ts`
- Modify: `admin/app.vue`
- Modify: `admin/layouts/default.vue`
- Modify: `admin/middleware/auth.global.ts`
- Modify: `admin/composables/useAuth.ts`
- Modify: `admin/composables/useApi.ts`
- Modify: `admin/pages/index.vue`
- Modify: `admin/pages/orders/index.vue`
- Modify: `admin/pages/menu/index.vue`
- Modify: `admin/pages/settings.vue`

## Task 1: Branch-Aware Schema and Seed Foundation

**Files:**
- Create: `server/migrations/000002_branches_and_permissions.up.sql`
- Create: `server/migrations/000002_branches_and_permissions.down.sql`
- Create: `server/internal/model/branch.go`
- Modify: `server/internal/model/store.go`
- Modify: `server/internal/model/staff.go`
- Modify: `server/internal/model/category.go`
- Modify: `server/internal/model/item.go`
- Modify: `server/internal/model/modifier.go`
- Modify: `server/internal/model/order.go`
- Modify: `server/cmd/seed/main.go`

- [ ] Add the `branches` table and new `branch_id` or `category` columns exactly as defined in the spec, keeping `store_id` on menu entities and orders for tenant scoping and commission calculations.
- [ ] Backfill existing seeded store/menu data into a single default branch so current data remains usable after migration.
- [ ] Update Go models so branch-aware JSON fields are available to both admin and mini app clients.
- [ ] Extend the seed command to create `Demo Bar - Main` with the Tashkent coordinates from the spec and assign seeded menu and staff records to that branch.
- [ ] Run: `go test ./...`
- [ ] Run: `go run ./cmd/migrate`
- [ ] Commit: `git commit -m "feat: add branch-aware schema foundation"`

## Task 2: JWT, Middleware, and Permission Boundaries

**Files:**
- Create: `server/internal/service/permission_service.go`
- Modify: `server/internal/middleware/auth.go`
- Modify: `server/internal/service/auth_service.go`
- Modify: `server/internal/handler/auth_handler.go`
- Modify: `server/internal/repository/staff_repo.go`
- Modify: `server/internal/model/staff.go`

- [ ] Extend admin JWT claims to include `branch_id` and keep `store_id` for all admin users.
- [ ] Treat `director` as store-wide and `manager` or `barista` as branch-scoped in middleware helpers rather than scattering role checks across handlers.
- [ ] Add reusable permission checks for all spec actions such as `branch:create`, `staff:create:manager`, `menu:manage`, and `dashboard:all`.
- [ ] Update admin login response payload so the frontend receives role, branch assignment, and enough metadata to render route guards immediately.
- [ ] Preserve user JWT behavior for mini app users, but extend Telegram auth to accept phone data and persist it on first successful contact share.
- [ ] Run: `go test ./...`
- [ ] Commit: `git commit -m "feat: add branch-aware auth and permissions"`

## Task 3: Repositories and Public Branch Discovery API

**Files:**
- Create: `server/internal/repository/branch_repo.go`
- Create: `server/internal/handler/branch_handler.go`
- Modify: `server/internal/repository/store_repo.go`
- Modify: `server/internal/repository/menu_repo.go`
- Modify: `server/internal/handler/store_handler.go`
- Modify: `server/internal/handler/router.go`

- [ ] Add repository methods for active-branch discovery, branch detail lookup, branch lookup by store slug, and branch-scoped menu loading.
- [ ] Keep current store-slug endpoints working temporarily, but implement `/discover`, `/branches/:id`, and `/branches/:id/menu` as the canonical public read API.
- [ ] Make `/discover` return the exact fields the mini app needs in one payload: store name, store logo, store category, branch id, branch name, branch address, coordinates, banner image, and preview menu items.
- [ ] Add store-level helper logic for bot deep links so one-branch stores can resolve directly and multi-branch stores can return a picker payload.
- [ ] Run: `go test ./...`
- [ ] Smoke-check with `curl` against `/discover`, `/branches/:id`, and `/branches/:id/menu`.
- [ ] Commit: `git commit -m "feat: add public branch discovery endpoints"`

## Task 4: Orders, Notifications, and Branch Scope Propagation

**Files:**
- Modify: `server/internal/model/order.go`
- Modify: `server/internal/repository/order_repo.go`
- Modify: `server/internal/service/order_service.go`
- Modify: `server/internal/handler/order_handler.go`
- Modify: `server/internal/ws/hub.go`
- Modify: `server/internal/telegram/notifications.go`

- [ ] Change order creation to require `branch_id`, validate that all items and modifiers belong to that branch, and derive `store_id` server-side from the branch instead of trusting the client.
- [ ] Keep WebSocket subscriptions store-aware for director aggregate views, but add branch-targeted fan-out for manager and barista screens.
- [ ] Route Telegram order notifications to `branches.telegram_group_chat_id` with fallback behavior defined for missing branch chat ids.
- [ ] Update order list queries so directors can see all store orders, while managers and baristas only see their branch unless an explicit director filter is selected.
- [ ] Keep order status transitions unchanged.
- [ ] Run: `go test ./...`
- [ ] Run manual flow: create order -> update status -> confirm user and admin WebSocket updates still arrive.
- [ ] Commit: `git commit -m "feat: scope orders and notifications by branch"`

## Task 5: Admin Branch and Staff APIs

**Files:**
- Create: `server/internal/handler/staff_handler.go`
- Modify: `server/internal/repository/staff_repo.go`
- Modify: `server/internal/repository/branch_repo.go`
- Modify: `server/internal/handler/router.go`
- Modify: `server/internal/handler/menu_handler.go`
- Modify: `server/internal/repository/category_repo.go`
- Modify: `server/internal/repository/item_repo.go`
- Modify: `server/internal/repository/modifier_repo.go`

- [ ] Implement `/admin/branches` CRUD with director-only create, edit, and deactivate semantics.
- [ ] Implement `/admin/staff` list, create, update, and deactivate endpoints with role matrix enforcement from the spec.
- [ ] Update all admin menu endpoints to accept `branch_id`, validate branch ownership, and automatically pin non-directors to their assigned branch.
- [ ] Add list responses that are already shaped for the admin UI: grouped staff for director view, branch summaries with staff counts, and branch-scoped menu payloads.
- [ ] Run: `go test ./...`
- [ ] Smoke-check the new admin endpoints with authenticated director and manager tokens.
- [ ] Commit: `git commit -m "feat: add admin branch and staff management api"`

## Task 6: Mini App Design System and Telegram Runtime Foundation

**Files:**
- Modify: `web/package.json`
- Modify: `web/src/index.css`
- Modify: `web/src/lib/api.ts`
- Create: `web/src/lib/telegram.ts`
- Create: `web/src/lib/theme.ts`
- Create: `web/src/components/layout/AppShell.tsx`
- Create: `web/src/hooks/useTelegramTheme.ts`

- [ ] Install the missing runtime dependencies: Telegram Mini App SDK, `maplibre-gl`, and `lucide-react`. Keep the dependency set minimal; do not add a component framework.
- [ ] Replace the current bare `index.css` with the spec token layer for Telegram colors, brand colors, radii, shadows, spacing, safe-area handling, reduced-motion rules, and utility classes for cards, pills, and bottom bars.
- [ ] Move API base URL and WebSocket URL configuration out of hard-coded localhost values so both frontends can run in dev and production without file edits.
- [ ] Add Telegram bootstrapping utilities for theme sync, viewport expansion, contact request, location fallback handling, and init data access.
- [ ] Build a thin `AppShell` that owns safe-area padding, background color, loading skeletons, and route-level page transitions.
- [ ] Run: `npm --prefix web run build`
- [ ] Commit: `git commit -m "feat: add mini app design system foundation"`

## Task 7: Phone Gate and Discovery-First Mini App Home

**Files:**
- Create: `web/src/hooks/useTelegramAuth.ts`
- Create: `web/src/hooks/useUserLocation.ts`
- Create: `web/src/hooks/useDiscovery.ts`
- Create: `web/src/components/auth/PhoneGate.tsx`
- Create: `web/src/components/discovery/ViewToggle.tsx`
- Create: `web/src/components/discovery/CategoryTabs.tsx`
- Create: `web/src/components/discovery/DiscoveryMap.tsx`
- Create: `web/src/components/discovery/BranchMarker.tsx`
- Create: `web/src/components/discovery/BranchSheet.tsx`
- Create: `web/src/components/discovery/BranchListCard.tsx`
- Create: `web/src/pages/HomePage.tsx`
- Modify: `web/src/App.tsx`
- Modify: `web/src/types/index.ts`

- [ ] Make `/` the real home screen and gate all content behind the phone verification flow from the spec.
- [ ] Persist user auth token locally, auto-auth on repeat opens with `initData`, and only show the hard-block phone gate when the user has not yet shared contact info.
- [ ] Keep the map mounted when list mode is active, with view toggling handled by CSS and component state rather than remounting MapLibre.
- [ ] Implement discovery sorting and filtering client-side using current user location, with Tashkent center as the fallback if location permission is denied.
- [ ] Match the list and bottom-sheet layouts to the design system spec, including preview carousel items and CTA to the branch menu page.
- [ ] Run: `npm --prefix web run build`
- [ ] Manual QA: phone deny, phone retry, location deny, category filter, marker selection, branch card dismissal.
- [ ] Commit: `git commit -m "feat: add mini app auth gate and discovery home"`

## Task 8: Branch Menu, Item Detail, Cart, and Order Tracking Redesign

**Files:**
- Create: `web/src/components/menu/MenuHeader.tsx`
- Create: `web/src/components/menu/ItemCard.tsx`
- Create: `web/src/components/menu/ModifierGroupSelector.tsx`
- Create: `web/src/components/cart/CartBar.tsx`
- Create: `web/src/components/common/StatusBadge.tsx`
- Create: `web/src/pages/BranchPage.tsx`
- Create: `web/src/pages/ItemPage.tsx`
- Modify: `web/src/store/cart.ts`
- Modify: `web/src/pages/CartPage.tsx`
- Modify: `web/src/pages/OrderPage.tsx`
- Modify: `web/src/pages/OrdersPage.tsx`
- Modify: `web/src/types/index.ts`
- Modify: `web/src/App.tsx`
- Retire usage from: `web/src/pages/StorePage.tsx`

- [ ] Replace slug-based routing with `/branch/:id` and `/item/:id`, keeping a small compatibility redirect for old store deep links until the bot flow is updated.
- [ ] Move modifier selection out of the modal and into the new full-page item detail screen.
- [ ] Change cart persistence from `storeId` and `storeSlug` to `branchId` plus enough branch/store display metadata for cart and order history screens.
- [ ] Rebuild cart, order tracking, and order history pages to use the new status badge, progress bar, and branch-aware copy from the spec.
- [ ] Keep order submission and live tracking behavior unchanged from the user’s perspective except for the new branch scope.
- [ ] Run: `npm --prefix web run build`
- [ ] Manual QA: add item with modifiers, switch branches clears cart, place order with `branch_id`, track status, cancel pending order.
- [ ] Commit: `git commit -m "feat: redesign branch menu cart and order flow"`

## Task 9: Admin Design System and Navigation Shell

**Files:**
- Create: `admin/assets/css/main.css`
- Create: `admin/types/auth.ts`
- Create: `admin/composables/usePermissions.ts`
- Create: `admin/composables/useBranchContext.ts`
- Create: `admin/components/layout/AppSidebar.vue`
- Create: `admin/components/layout/AppHeader.vue`
- Create: `admin/components/ui/StatCard.vue`
- Create: `admin/components/ui/EmptyState.vue`
- Modify: `admin/package.json`
- Modify: `admin/nuxt.config.ts`
- Modify: `admin/tailwind.config.ts`
- Modify: `admin/app.vue`
- Modify: `admin/layouts/default.vue`
- Modify: `admin/middleware/auth.global.ts`
- Modify: `admin/composables/useAuth.ts`
- Modify: `admin/composables/useApi.ts`

- [ ] Add the admin token layer, typography, spacing, borders, and color variables from the design spec in a single global stylesheet rather than sprinkling raw classes across pages.
- [ ] Install the small missing UI dependencies needed for the admin redesign, including Lucide icons and the specific shadcn-vue primitives actually used.
- [ ] Add route-level role checks in `auth.global.ts` and a centralized `usePermissions()` composable so the same permission map drives both nav visibility and action guards.
- [ ] Add a persisted branch switcher state for directors and expose it through `useBranchContext()` so dashboard, orders, and menu pages can switch between aggregate and branch views.
- [ ] Replace the current sidebar shell with the new professional B2B layout and header treatment before rebuilding feature pages.
- [ ] Run: `npm --prefix admin run build`
- [ ] Commit: `git commit -m "feat: add admin design system and permission shell"`

## Task 10: Admin Branches, Staff, Dashboard, Orders, Menu, and Settings

**Files:**
- Create: `admin/components/branches/BranchForm.vue`
- Create: `admin/components/branches/BranchTable.vue`
- Create: `admin/components/staff/StaffForm.vue`
- Create: `admin/components/staff/StaffGroupList.vue`
- Create: `admin/pages/branches.vue`
- Create: `admin/pages/staff.vue`
- Create: `admin/pages/settings/branch.vue`
- Modify: `admin/pages/index.vue`
- Modify: `admin/pages/orders/index.vue`
- Modify: `admin/pages/menu/index.vue`
- Modify: `admin/pages/settings.vue`

- [ ] Build the director-only branches page with list, create, edit, deactivate, and draggable map pin flows.
- [ ] Build the staff page with grouped director view, manager-limited branch view, and correct role or branch form behavior.
- [ ] Rework dashboard cards so director aggregate mode shows totals and per-branch breakdowns, while manager and barista only see branch stats.
- [ ] Update orders and menu pages to respect branch context and permission guards, including director “All Branches” behavior and non-director branch pinning.
- [ ] Split store-level settings and branch-level settings into the routes defined in the spec.
- [ ] Run: `npm --prefix admin run build`
- [ ] Manual QA: director branch switching, manager restricted views, barista blocked routes, branch CRUD, staff CRUD, branch-scoped menu changes.
- [ ] Commit: `git commit -m "feat: add branch-aware admin management screens"`

## Task 11: Bot Deep Links, Compatibility Cleanup, and End-to-End Verification

**Files:**
- Modify: `server/internal/telegram/bot.go`
- Modify: `server/internal/handler/router.go`
- Modify: `web/src/App.tsx`
- Modify: `admin/pages/login.vue`
- Modify: `web/README.md`
- Modify: `design-system/xpressgo/pages/miniapp.md`
- Modify: `design-system/xpressgo/pages/admin.md`

- [ ] Update Telegram bot deep-link behavior so store slug entry resolves to a direct branch page for single-branch stores and a branch picker flow for multi-branch stores.
- [ ] Remove or redirect obsolete store-slug-only UX paths once branch-aware routing is stable.
- [ ] Update local developer docs for new environment variables, required dependencies, and seed behavior.
- [ ] Run: `go test ./...`
- [ ] Run: `npm --prefix web run build`
- [ ] Run: `npm --prefix admin run build`
- [ ] Perform final end-to-end smoke test: Telegram auth, discovery, branch menu, order placement, admin role login, branch switcher, order status updates.
- [ ] Commit: `git commit -m "chore: finalize branch discovery rollout"`

## Cross-Cutting Notes

- Keep `store_id` on orders and menu entities even after `branch_id` is introduced. It simplifies tenant isolation, commission logic, and director aggregate queries.
- Do not attempt a shared component library between `web` and `admin` in this rollout. Shared tokens and patterns are enough; shared implementation would slow delivery.
- Add MapLibre in both frontends only where it is required: mini app discovery map and admin branch form map.
- The current repo has no frontend test harness. Do not invent a large UI testing stack in this feature branch unless repeated regressions force it. Use build checks plus explicit manual QA checklists.
- Prefer compatibility redirects and transitional handlers over a flag-day migration. The backend and bot can support both old and new entry points briefly while the mini app is replaced.

## Self-Review

- Spec coverage check:
  - Branch model, permissions, admin CRUD, and branch-scoped menu or order flows are covered by Tasks 1 through 5 and Task 10.
  - Phone auth gate, discovery map or list, branch menu page, item page, cart, and order tracking redesign are covered by Tasks 6 through 8.
  - Admin design system, route guards, branch switcher, dashboard, and settings split are covered by Tasks 9 and 10.
  - Bot deep-link behavior and compatibility cleanup are covered by Task 11.
- Placeholder scan: no `TODO`, `TBD`, or “implement later” markers remain.
- Consistency check: branch scope is introduced once in schema and then reused consistently in auth, repositories, orders, mini app routing, and admin context.
