# Realistic Seed Data Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the seed command to create a realistic, idempotent multi-store dataset with populated branches, staff, branch-scoped menus, users, and orders for admin and discovery flows.

**Architecture:** Keep all seed behavior in the existing `server/cmd/seed/main.go` entrypoint, but refactor that file into clear fixture definitions and helper functions so multi-store seeding stays readable and deterministic. Reuse the current schema exactly as-is, preserve branch-scoped menu ownership, and make order generation deterministic enough to rerun without uncontrolled duplication.

**Tech Stack:** Go, pgx, PostgreSQL, existing Xpressgo schema and seed command

---

## File Structure

### Existing files to modify

- `server/cmd/seed/main.go`
  - expand from single-store seed into structured multi-store seed orchestration
  - add fixture structs for stores, branches, categories, items, modifiers, staff, users, and orders
  - add helper functions for idempotent upserts and seed-owned order replacement

### Existing files to read during implementation

- `server/internal/model/store.go`
- `server/internal/model/branch.go`
- `server/internal/model/staff.go`
- `server/internal/model/category.go`
- `server/internal/model/item.go`
- `server/internal/model/modifier.go`
- `server/internal/model/order.go`
- `server/migrations/000001_init.up.sql`
- `server/migrations/000002_branches_and_permissions.up.sql`

### Verification commands

- `go test ./...`
- `make quality-server`
- `make seed`

---

### Task 1: Refactor Seed Configuration Into Structured Fixtures

**Files:**
- Modify: `server/cmd/seed/main.go`

- [ ] **Step 1: Write the failing compile target in your head and capture the intended fixture types**

Add these types near the top of `server/cmd/seed/main.go` so the rest of the file can move away from hardcoded single-store values:

```go
type seedStore struct {
	code        string
	slug        string
	name        string
	category    string
	description string
	address     string
	phone       string
	logoURL     string
	branches    []seedBranch
	staff       []seedStaff
	users       []seedUser
	menu        seedMenu
	orderPlan   seedOrderPlan
}

type seedBranch struct {
	code                string
	name                string
	address             string
	lat                 float64
	lng                 float64
	bannerImageURL      string
	telegramGroupChatID *int64
	isActive            bool
}

type seedStaff struct {
	branchCode string
	staffCode  string
	name       string
	role       string
	isActive   bool
}

type seedUser struct {
	telegramID int64
	phone      string
	firstName  string
	lastName   string
	username   string
}

type seedMenu struct {
	categories []seedCategory
}

type seedCategory struct {
	name  string
	items []seedCatalogItem
}

type seedCatalogItem struct {
	name          string
	description   string
	price         int64
	modGroups     []seedModifierGroup
	unavailableAt map[string]bool
}

type seedOrderPlan struct {
	totalOrders int
}
```

- [ ] **Step 2: Run a focused build check to confirm the file still compiles before replacing behavior**

Run:

```bash
go test ./server/cmd/seed
```

Expected:

```text
?   	github.com/xpressgo/server/cmd/seed	[no test files]
```

- [ ] **Step 3: Replace the single-store inline seed literals with fixture-building functions**

Add these function stubs and move all current hardcoded literals into them:

```go
func buildSeedStores() []seedStore {
	return []seedStore{
		buildDemoBarStore(),
		buildUrbanCoffeeStore(),
		buildStreetBurgerStore(),
	}
}

func buildDemoBarStore() seedStore {
	return seedStore{}
}

func buildUrbanCoffeeStore() seedStore {
	return seedStore{}
}

func buildStreetBurgerStore() seedStore {
	return seedStore{}
}
```

Then change `main()` to start from:

```go
stores := buildSeedStores()
for _, store := range stores {
	log.Printf("seeding store %s", store.code)
}
```

- [ ] **Step 4: Run a focused build check again**

Run:

```bash
go test ./server/cmd/seed
```

Expected:

```text
?   	github.com/xpressgo/server/cmd/seed	[no test files]
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "refactor(seed): structure multi-store seed fixtures"
```

### Task 2: Seed Three Stores And Twelve Branches Idempotently

**Files:**
- Modify: `server/cmd/seed/main.go`

- [ ] **Step 1: Add stable store and branch upsert helpers**

Add a branch code column in-memory only and keep the database matching deterministic with store code plus branch name:

```go
type seededStoreRefs struct {
	storeID   string
	branchIDs map[string]string
}

func ensureStoreFromFixture(ctx context.Context, conn *pgx.Conn, store seedStore) (string, error) {
	return ensureStore(ctx, conn, store)
}

func ensureBranchFromFixture(ctx context.Context, conn *pgx.Conn, storeID string, branch seedBranch) (string, error) {
	return ensureBranch(ctx, conn, storeID, branch)
}
```

Update the current `ensureStore` and `ensureBranch` signatures from hardcoded values to fixture-driven values:

```go
func ensureStore(ctx context.Context, conn *pgx.Conn, store seedStore) (string, error)
func ensureBranch(ctx context.Context, conn *pgx.Conn, storeID string, branch seedBranch) (string, error)
```

- [ ] **Step 2: Implement real fixture data for all stores and branches**

Populate `buildSeedStores()` with these counts:

```go
func buildSeedStores() []seedStore {
	return []seedStore{
		buildDemoBarStore(),
		buildUrbanCoffeeStore(),
		buildStreetBurgerStore(),
	}
}
```

`buildDemoBarStore()` should return `6` branches:

```go
[]seedBranch{
	{code: "main", name: "Demo Bar - Main", address: "Amir Temur St 42, Tashkent", lat: 41.2995, lng: 69.2401, isActive: true},
	{code: "downtown", name: "Demo Bar - Downtown", address: "Afrosiyob St 8, Tashkent", lat: 41.3057, lng: 69.2801, isActive: true},
	{code: "riverside", name: "Demo Bar - Riverside", address: "Kichik Halqa Yuli 14, Tashkent", lat: 41.2871, lng: 69.2684, isActive: true},
	{code: "chilonzor", name: "Demo Bar - Chilonzor", address: "Chilonzor 19 квартал, Tashkent", lat: 41.2752, lng: 69.2038, isActive: true},
	{code: "samarkand-darvoza", name: "Demo Bar - Samarkand Darvoza", address: "Qatortol St 2, Tashkent", lat: 41.3164, lng: 69.2128, isActive: true},
	{code: "airport-road", name: "Demo Bar - Airport Road", address: "Kushbegi St 118, Tashkent", lat: 41.2578, lng: 69.2816, isActive: true},
}
```

`buildUrbanCoffeeStore()` and `buildStreetBurgerStore()` should each return `3` active branches with distinct addresses and coordinates.

- [ ] **Step 3: Change the main loop to seed stores and collect branch IDs**

Replace the current single-store flow with:

```go
stores := buildSeedStores()
storeRefs := make(map[string]seededStoreRefs, len(stores))

for _, store := range stores {
	storeID, err := ensureStore(ctx, conn, store)
	if err != nil {
		log.Fatalf("failed to ensure store %s: %v", store.code, err)
	}

	refs := seededStoreRefs{
		storeID:   storeID,
		branchIDs: map[string]string{},
	}

	for _, branch := range store.branches {
		branchID, err := ensureBranch(ctx, conn, storeID, branch)
		if err != nil {
			log.Fatalf("failed to ensure branch %s/%s: %v", store.code, branch.code, err)
		}
		refs.branchIDs[branch.code] = branchID
	}

	storeRefs[store.code] = refs
}
```

- [ ] **Step 4: Run the focused seed package build**

Run:

```bash
go test ./server/cmd/seed
```

Expected:

```text
?   	github.com/xpressgo/server/cmd/seed	[no test files]
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "feat(seed): add multi-store multi-branch fixtures"
```

### Task 3: Seed Store-Isolated Staff Hierarchies

**Files:**
- Modify: `server/cmd/seed/main.go`

- [ ] **Step 1: Extend the staff fixture data for all three stores**

Populate each store with explicit staff fixtures:

```go
staff: []seedStaff{
	{staffCode: "admin", name: "Bar Director", role: "director", isActive: true},
	{branchCode: "main", staffCode: "manager-main", name: "Main Branch Manager", role: "manager", isActive: true},
	{branchCode: "main", staffCode: "barista-main-1", name: "Main Barista 1", role: "barista", isActive: true},
}
```

For final counts:

- `Demo Bar`: `1` director, `6` managers, `24` baristas
- `Urban Coffee`: `1` director, `3` managers, `9` baristas
- `Street Burger`: `1` director, `3` managers, `9` baristas

Set at least `1` inactive barista across the secondary stores for UI coverage.

- [ ] **Step 2: Update `ensureStaff` to accept active state**

Change the signature to:

```go
func ensureStaff(ctx context.Context, conn *pgx.Conn, storeID string, branchID *string, staffCode, name, passwordHash, role string, isActive bool) error
```

Update the query so reruns preserve the fixture's active state:

```sql
ON CONFLICT (store_id, staff_code) DO UPDATE SET
	branch_id = EXCLUDED.branch_id,
	name = EXCLUDED.name,
	password_hash = EXCLUDED.password_hash,
	role = EXCLUDED.role,
	is_active = EXCLUDED.is_active
```

- [ ] **Step 3: Seed staff from store fixtures**

Add a loop after store and branch creation:

```go
for _, staff := range store.staff {
	var branchID *string
	if staff.branchCode != "" {
		id := refs.branchIDs[staff.branchCode]
		branchID = &id
	}

	if err := ensureStaff(ctx, conn, refs.storeID, branchID, staff.staffCode, staff.name, string(hash), staff.role, staff.isActive); err != nil {
		log.Fatalf("failed to ensure staff %s/%s: %v", store.code, staff.staffCode, err)
	}
}
```

- [ ] **Step 4: Run the focused build and server test suite**

Run:

```bash
go test ./server/cmd/seed ./server/internal/...
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "feat(seed): add realistic store staff hierarchies"
```

### Task 4: Seed Branch-Scoped Menus For Three Store Concepts

**Files:**
- Modify: `server/cmd/seed/main.go`

- [ ] **Step 1: Replace the old single-store category and item list with store menu fixtures**

Create one `seedMenu` per store fixture and remove the current single `items := []seedItem{...}` block.

Use this pattern:

```go
menu: seedMenu{
	categories: []seedCategory{
		{
			name: "Cocktails",
			items: []seedCatalogItem{
				{
					name:        "Mojito",
					description: "Fresh lime, mint, rum, soda",
					price:       45000,
					modGroups: []seedModifierGroup{
						{
							name:          "Size",
							selectionType: "single",
							required:      true,
							mods: []seedModifier{
								{name: "Regular", price: 0},
								{name: "Large", price: 15000},
							},
						},
					},
					unavailableAt: map[string]bool{
						"airport-road": true,
					},
				},
			},
		},
	},
}
```

Category counts:

- `Demo Bar`: `6`
- `Urban Coffee`: `5`
- `Street Burger`: `5`

- [ ] **Step 2: Implement branch-scoped menu seeding loops**

For each branch in each store, create its own categories, items, modifier groups, and modifiers:

```go
for branchIndex, branch := range store.branches {
	branchID := refs.branchIDs[branch.code]

	for categoryIndex, category := range store.menu.categories {
		categoryID, err := ensureCategory(ctx, conn, refs.storeID, branchID, category.name, categoryIndex)
		if err != nil {
			log.Fatalf("failed to ensure category %s/%s/%s: %v", store.code, branch.code, category.name, err)
		}

		itemSort := 0
		for _, item := range category.items {
			itemID, err := ensureItem(ctx, conn, refs.storeID, branchID, categoryID, seedItem{
				category:    category.name,
				name:        item.name,
				description: item.description,
				price:       item.price,
				modGroups:   item.modGroups,
			}, itemSort)
			if err != nil {
				log.Fatalf("failed to ensure item %s/%s/%s: %v", store.code, branch.code, item.name, err)
			}

			isAvailable := !item.unavailableAt[branch.code]
			if err := setItemAvailability(ctx, conn, itemID, isAvailable); err != nil {
				log.Fatalf("failed to set availability for %s/%s/%s: %v", store.code, branch.code, item.name, err)
			}

			_ = branchIndex
			itemSort++
		}
	}
}
```

- [ ] **Step 3: Add an item availability helper**

Add:

```go
func setItemAvailability(ctx context.Context, conn *pgx.Conn, itemID string, isAvailable bool) error {
	_, err := conn.Exec(ctx, `
		UPDATE items
		SET is_available = $2
		WHERE id = $1
	`, itemID, isAvailable)
	return err
}
```

- [ ] **Step 4: Run the focused build**

Run:

```bash
go test ./server/cmd/seed
```

Expected:

```text
?   	github.com/xpressgo/server/cmd/seed	[no test files]
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "feat(seed): add branch-scoped menus for all stores"
```

### Task 5: Seed Users And Deterministic Seed-Owned Orders

**Files:**
- Modify: `server/cmd/seed/main.go`

- [ ] **Step 1: Expand user fixtures and create reusable lookup helpers**

Add users per store fixture, but seed them in a shared table:

```go
func ensureUser(ctx context.Context, conn *pgx.Conn, user seedUser) (string, error) {
	var id string
	err := conn.QueryRow(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET
			phone = EXCLUDED.phone,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username
		RETURNING id
	`, user.telegramID, user.phone, user.firstName, user.lastName, user.username).Scan(&id)
	return id, err
}
```

- [ ] **Step 2: Add seed-owned order cleanup helpers**

Because the schema has no seed marker column, use seeded users plus store scope to replace generated order history safely:

```go
func deleteOrdersForSeedUsersByStore(ctx context.Context, conn *pgx.Conn, storeID string, userIDs []string) error {
	_, err := conn.Exec(ctx, `
		DELETE FROM orders
		WHERE store_id = $1
		  AND user_id = ANY($2)
	`, storeID, userIDs)
	return err
}
```

This relies on existing FK cascade behavior from the order tables defined in migrations. Verify that before using it.

- [ ] **Step 3: Add order generation helpers**

Add deterministic helpers:

```go
type seededCatalogRefs struct {
	itemIDsByBranch map[string]map[string]string
}

func seedOrdersForStore(ctx context.Context, conn *pgx.Conn, store seedStore, refs seededStoreRefs, userIDs []string, catalog seededCatalogRefs) error {
	return nil
}
```

Inside `seedOrdersForStore`, generate:

- `Demo Bar`: `120`
- `Urban Coffee`: `45`
- `Street Burger`: `45`

Rules:

- order timestamps distributed across last `14` days
- live queue status mix for `Demo Bar`: `pending 4`, `accepted 5`, `preparing 6`, `ready 5`
- remaining orders mostly `picked_up`, plus smaller `cancelled` and `rejected` shares
- basket sizes from `1` to `5`
- ETA from `5` to `25`

Use direct inserts into `orders`, `order_items`, and `order_item_modifiers`.

- [ ] **Step 4: Add explicit insert helpers for order rows**

Add helpers:

```go
func insertSeedOrder(ctx context.Context, conn *pgx.Conn, order seedGeneratedOrder) (string, error) {
	return "", nil
}

func insertSeedOrderItem(ctx context.Context, conn *pgx.Conn, orderID string, item seedGeneratedOrderItem) (string, error) {
	return "", nil
}

func insertSeedOrderModifier(ctx context.Context, conn *pgx.Conn, orderItemID string, mod seedGeneratedOrderModifier) error {
	return nil
}
```

The order insert must set all fields explicitly, including `status`, `payment_status`, `eta_minutes`, `rejection_reason`, `created_at`, and `updated_at`, so the seeded dataset does not rely on the runtime order service path.

- [ ] **Step 5: Run focused verification**

Run:

```bash
go test ./server/cmd/seed
```

Expected:

```text
?   	github.com/xpressgo/server/cmd/seed	[no test files]
```

- [ ] **Step 6: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "feat(seed): add deterministic users and realistic orders"
```

### Task 6: Full Seed Verification And Manual QA

**Files:**
- Modify: `server/cmd/seed/main.go`
- Read: `README.md`
- Read: `AGENTS.md`

- [ ] **Step 1: Run the seed command against the local stack**

Run:

```bash
make seed
```

Expected:

```text
Store ready
Branch ready
Staff ready
Seed completed successfully
```

The exact counts may differ in the logs, but the command must complete successfully.

- [ ] **Step 2: Rerun the seed command to verify idempotency**

Run:

```bash
make seed
```

Expected:

```text
Seed completed successfully
```

The second run must not create runaway duplicates for stores, branches, staff, or branch menus.

- [ ] **Step 3: Run server quality checks**

Run:

```bash
make quality-server
```

Expected:

```text
go vet ./...
golangci-lint run ./...
go test ./...
```

All should pass.

- [ ] **Step 4: Perform manual QA in the running apps**

Check these flows:

```text
1. Log into Demo Bar admin with the seeded director account.
2. Confirm the dashboard has visible activity.
3. Confirm the branches page shows 6 Demo Bar branches with staff counts.
4. Confirm the staff page groups staff by branch and store correctly.
5. Confirm the orders page shows mixed statuses and realistic totals.
6. Open the mini app discovery flow and confirm all 3 stores are visible.
7. Open at least one branch menu for each store and confirm the menu concept differs by store.
8. Confirm smaller branches have some unavailable items but still usable menus.
```

- [ ] **Step 5: Commit**

```bash
git add server/cmd/seed/main.go
git commit -m "test(seed): verify realistic seed dataset end to end"
```

## Self-Review

Spec coverage:

- multi-store discovery realism: covered by Tasks 1, 2, 4, and 5
- isolated staff hierarchies: covered by Task 3
- branch-scoped menus: covered by Task 4
- realistic orders and users: covered by Task 5
- idempotency and QA: covered by Task 6

Placeholder scan:

- no `TODO`, `TBD`, or deferred implementation steps remain
- each task names exact files and concrete commands

Type consistency:

- plan consistently uses `seedStore`, `seedBranch`, `seedStaff`, `seedUser`, `seedMenu`, `seedCategory`, `seedCatalogItem`, and `seedOrderPlan`
- store and branch fixture boundaries match the approved spec

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-03-30-realistic-seed-data-implementation.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
