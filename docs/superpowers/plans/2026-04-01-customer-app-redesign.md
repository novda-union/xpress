# Customer App Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

---

## ⚡ HANDOFF STATUS — Pick up from Task 6

**Last updated:** 2026-04-01
**Current branch:** `master`
**Base SHA (before this work):** `631c557e4ecd3a4b87f5c22578658d39274b7aa4`

### Completed tasks

| Task | Status | Commit |
|---|---|---|
| Task 1: Fix z-index layering | ✅ DONE | `4a9daf8` |
| Task 2: BranchPage skeleton + activeCategory fix | ✅ DONE | `ed826db` |
| Task 3: OrderPage skeleton | ✅ DONE | `b492627` |
| Task 4: Add CreatedAt to model.Item | ✅ DONE | `73568be` — migration `000005_items_created_at` applied |
| Task 5: DiscoverItem + feed queries to ItemRepo | ✅ DONE | `5a7f6d0` |

### Remaining tasks (start here → Task 6)

| Task | Status |
|---|---|
| Task 6: Add GetByIDForBranch to ItemRepo | ⏳ NOT STARTED |
| Task 7: Create DiscoverHandler | ⏳ NOT STARTED |
| Task 8: Create ItemHandler | ⏳ NOT STARTED |
| Task 9: Wire handlers into router + main | ⏳ NOT STARTED |
| Task 10: Add DiscoverItem + BranchCart to frontend types | ⏳ NOT STARTED |
| Task 11: Rewrite useCartStore to multi-cart v3 | ⏳ NOT STARTED |
| Task 12: BranchConflictSheet component | ⏳ NOT STARTED |
| Task 13: DiscoverItemCard component | ⏳ NOT STARTED |
| Task 14: Update CartBar for multi-cart | ⏳ NOT STARTED |
| Task 15: CartPage multi-cart tabs | ⏳ NOT STARTED |
| Task 16: Redesign ItemPage | ⏳ NOT STARTED |
| Task 17: Create useDiscoveryFeed and useDiscoveryItems hooks | ⏳ NOT STARTED |
| Task 18: Rewrite HomePage | ⏳ NOT STARTED |

### Known forward-reference TypeScript errors (intentional, will resolve as tasks complete)

`web/src/pages/BranchPage.tsx` has 3 TypeScript errors that are **intentional forward references** — do NOT fix them by reverting:
- `cart.activeBranchCount()` — resolved by Task 11
- `cart.activeBranchTotal()` — resolved by Task 11
- `cart.totalCartsCount()` + `CartBar totalCartsCount` prop — resolved by Tasks 11 + 14

`make quality-web` and `make typecheck` will FAIL until Tasks 11 + 14 are complete. This is expected.

### How to continue

1. Read this plan from Task 6 onwards
2. Use `superpowers:subagent-driven-development` skill to execute remaining tasks
3. Tasks 6–9 are backend (Go). Use `docker compose exec -T server go build ./...` to verify — do NOT use local `go`.
4. Tasks 10–18 are frontend (React/TypeScript). Use `make quality-web` and `make typecheck` to verify (expect failures until Task 11+14 complete).
5. After Task 11+14 complete, `make quality-web` should pass cleanly.

---

**Goal:** Redesign the customer Telegram mini app from a branch-card list to an item-first discovery feed with multi-cart support, fixed loading states, and corrected z-index layering.

**Architecture:** New backend endpoints (`/discover/feed`, `/discover/items`, `/items/:id`) power an item-centric home page with curated sections and infinite scroll. The cart store is rewritten to v3 supporting one cart per branch, each placed as a separate order. The item detail page gets a dedicated endpoint eliminating the full-menu load.

**Tech Stack:** Go/Echo backend, pgx v5, React 19, TypeScript, Tailwind CSS 4, Zustand v5, Telegram Mini App SDK, Lucide icons.

---

## File Map

### New files
| File | Responsibility |
|---|---|
| `server/internal/handler/discover_handler.go` | `GET /discover/feed` and `GET /discover/items` endpoints |
| `server/internal/handler/item_handler.go` | `GET /items/:id?branch=` endpoint |
| `web/src/components/discovery/DiscoverItemCard.tsx` | Item card for feed (stepper, NEW badge, conflict routing) |
| `web/src/components/cart/BranchConflictSheet.tsx` | "Add from new branch?" confirmation sheet |
| `web/src/hooks/useDiscoveryFeed.ts` | Fetches `/discover/feed` sections |
| `web/src/hooks/useDiscoveryItems.ts` | Fetches `/discover/items` with pagination |

### Modified files
| File | What changes |
|---|---|
| `server/internal/model/item.go` | Add `CreatedAt time.Time` field |
| `server/internal/repository/item_repo.go` | Add `GetByIDForBranch`, `GetFeedSection`, `ListForFeed` |
| `server/internal/handler/router.go` | Register new routes, add new handlers to `Handlers` struct |
| `server/cmd/server/main.go` | Instantiate `DiscoverHandler` and `ItemHandler` |
| `web/src/types/index.ts` | Add `DiscoverItem`, `BranchCart`, update `CartMeta` |
| `web/src/store/cart.ts` | Full rewrite to multi-cart v3 |
| `web/src/components/discovery/ViewToggle.tsx` | `z-50` → `z-20` |
| `web/src/components/discovery/BranchSheet.tsx` | outer `z-40` → `z-30` |
| `web/src/components/cart/CartBar.tsx` | Multi-cart count + badge |
| `web/src/pages/CartPage.tsx` | Multi-cart branch tabs |
| `web/src/pages/ItemPage.tsx` | Back button, hero image, sticky header, new endpoint |
| `web/src/pages/BranchPage.tsx` | Skeleton loading, category fix |
| `web/src/pages/OrderPage.tsx` | Skeleton loading instead of blank screen |
| `web/src/pages/HomePage.tsx` | Full rewrite — greeting, chips, sections, feed |
| `web/src/index.css` | Add z-index CSS variables |

---

## Task 1: Fix z-index layering (ViewToggle, BranchSheet, CSS vars)

**Files:**
- Modify: `web/src/index.css`
- Modify: `web/src/components/discovery/ViewToggle.tsx`
- Modify: `web/src/components/discovery/BranchSheet.tsx`

- [ ] **Step 1: Add z-index CSS variables to index.css**

Open `web/src/index.css`. Find the `:root` block (around line 10). Add these variables at the end of the `:root` block, before the closing `}`:

```css
  --z-sticky: 10;
  --z-floating: 20;
  --z-sheet: 30;
  --z-modal: 40;
  --z-toast: 50;
```

- [ ] **Step 2: Fix ViewToggle z-index**

In `web/src/components/discovery/ViewToggle.tsx`, change line:
```tsx
className="fixed bottom-5 left-1/2 z-50 -translate-x-1/2"
```
to:
```tsx
className="fixed bottom-5 left-1/2 z-20 -translate-x-1/2"
```

- [ ] **Step 3: Fix BranchSheet z-index**

In `web/src/components/discovery/BranchSheet.tsx`, change the outer wrapper:
```tsx
<div className="fixed inset-0 z-40">
```
to:
```tsx
<div className="fixed inset-0 z-30">
```

The inner `z-10` on the sheet panel stays — it's relative to its own stacking context inside the `z-30` container, which is correct.

- [ ] **Step 4: Verify quality**

```bash
make quality-web
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add web/src/index.css web/src/components/discovery/ViewToggle.tsx web/src/components/discovery/BranchSheet.tsx
git commit -m "fix(web): fix z-index layering so ViewToggle stays below BranchSheet"
```

---

## Task 2: BranchPage skeleton loading + activeCategory fix

**Files:**
- Modify: `web/src/pages/BranchPage.tsx`

- [ ] **Step 1: Replace BranchPage with skeleton-aware version**

Replace the full contents of `web/src/pages/BranchPage.tsx`:

```tsx
import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { CartBar } from '../components/cart/CartBar'
import { AppShell } from '../components/layout/AppShell'
import { MenuHeader } from '../components/menu/MenuHeader'
import { ItemCard } from '../components/menu/ItemCard'
import { api } from '../lib/api'
import { useCartStore } from '../store/cart'
import type { BranchDetail, Menu, MenuItem } from '../types'

function BranchPageSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-[200px] w-full bg-[var(--xp-card-bg)]" />
      <div className="flex gap-2 px-4 py-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="h-8 w-20 rounded-full bg-[var(--xp-card-bg)]" />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-3 px-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="aspect-square rounded-xl bg-[var(--xp-card-bg)]" />
        ))}
      </div>
    </div>
  )
}

export default function BranchPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const cart = useCartStore()
  const [detail, setDetail] = useState<BranchDetail | null>(null)
  const [menu, setMenu] = useState<Menu | null>(null)
  const [activeCategory, setActiveCategory] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return

    api<BranchDetail>(`/branches/${id}`).then((nextDetail) => {
      setDetail(nextDetail)
      cart.setBranch({
        branchId: nextDetail.branch.id,
        branchName: nextDetail.branch.name,
        storeName: nextDetail.store.name,
        bannerImageUrl: nextDetail.branch.banner_image_url,
      })
    })

    api<Menu>(`/branches/${id}/menu`).then((nextMenu) => {
      setMenu(nextMenu)
      if (nextMenu.categories.length > 0) {
        setActiveCategory(nextMenu.categories[0].id)
      }
    })
  }, [id]) // eslint-disable-line react-hooks/exhaustive-deps

  const currentCategory = useMemo(() => {
    if (!menu) return null
    if (activeCategory) {
      return menu.categories.find((c) => c.id === activeCategory) ?? menu.categories[0] ?? null
    }
    return menu.categories[0] ?? null
  }, [activeCategory, menu])

  const cartCount = cart.activeBranchCount()
  const cartTotal = cart.activeBranchTotal()

  return (
    <AppShell
      header={<MenuHeader title={detail?.store.name ?? ''} count={cartCount} />}
      bottomBar={
        cartCount > 0 ? (
          <CartBar count={cartCount} total={cartTotal} onOpen={() => navigate('/cart')} />
        ) : null
      }
    >
      {!detail || !menu ? (
        <BranchPageSkeleton />
      ) : (
        <>
          <div className="relative">
            <img
              src={detail.branch.banner_image_url || 'https://placehold.co/900x400?text=Xpressgo'}
              alt={detail.branch.name}
              className="h-[200px] w-full object-cover"
            />
            <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/70 to-transparent px-4 py-6">
              <p className="text-[22px] font-bold text-white">{detail.store.name}</p>
              <p className="text-sm text-white/80">{detail.branch.name}</p>
            </div>
          </div>

          <div className="sticky top-12 z-10 bg-[var(--tg-theme-bg-color)]">
            <div className="scrollbar-none flex gap-2 overflow-x-auto px-4 py-3">
              {menu.categories.map((category) => (
                <button
                  type="button"
                  key={category.id}
                  onClick={() => setActiveCategory(category.id)}
                  className={`xp-pill whitespace-nowrap px-4 text-[13px] font-medium ${
                    activeCategory === category.id
                      ? 'bg-[var(--xp-brand)] text-white'
                      : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                  }`}
                >
                  {category.name}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 px-4 pb-28">
            {(currentCategory?.items ?? []).map((item: MenuItem) => (
              <ItemCard
                key={item.id}
                item={item}
                onSelect={(selected) => navigate(`/item/${selected.id}?branch=${detail.branch.id}`)}
              />
            ))}
          </div>
        </>
      )}
    </AppShell>
  )
}
```

Note: `cart.activeBranchCount()` and `cart.activeBranchTotal()` are new methods from the v3 cart store (Task 11). For now, temporarily keep the old calls — they will be updated when the cart store is rewritten in Task 11.

- [ ] **Step 2: Temporarily keep old cart calls until Task 11**

The `cart.activeBranchCount()` / `cart.activeBranchTotal()` methods don't exist yet. Keep the file as above — it will fail TypeScript until Task 11. That is intentional. Do not run quality checks on this file until after Task 11.

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/BranchPage.tsx
git commit -m "fix(web): add skeleton loading and fix activeCategory init on BranchPage"
```

---

## Task 3: OrderPage skeleton loading

**Files:**
- Modify: `web/src/pages/OrderPage.tsx`

- [ ] **Step 1: Replace the loading fallback with a skeleton**

In `web/src/pages/OrderPage.tsx`, replace:
```tsx
if (!order) {
  return <div className="min-h-screen flex items-center justify-center">Loading...</div>
}
```
with:
```tsx
if (!order) {
  return (
    <AppShell header={<MenuHeader title="Order" count={0} />}>
      <div className="animate-pulse px-4 pt-6 space-y-4">
        <div className="xp-card h-40 p-5" />
        <div className="xp-card h-32 p-5" />
        <div className="xp-card h-20 p-5" />
      </div>
    </AppShell>
  )
}
```

- [ ] **Step 2: Verify quality**

```bash
make quality-web
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/OrderPage.tsx
git commit -m "fix(web): replace blank loading screen with skeleton on OrderPage"
```

---

## Task 4: Add CreatedAt to model.Item + verify items table

**Files:**
- Modify: `server/internal/model/item.go`

- [ ] **Step 1: Check if items table has created_at column**

```bash
docker compose exec -T postgres psql -U postgres -d xpressgo -c "\d items"
```

Look for a `created_at` column. If it exists, proceed to Step 2. If it does NOT exist, create a migration first:

```bash
cat > server/migrations/000004_items_created_at.up.sql << 'EOF'
ALTER TABLE items ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
EOF

cat > server/migrations/000004_items_created_at.down.sql << 'EOF'
ALTER TABLE items DROP COLUMN IF EXISTS created_at;
EOF

docker compose exec -T server ./migrate
```

- [ ] **Step 2: Add CreatedAt to model.Item**

Read the current `server/internal/model/item.go`. Add `CreatedAt time.Time` to the struct:

```go
package model

import "time"

type Item struct {
	ID          string
	CategoryID  string
	StoreID     string
	BranchID    string
	Name        string
	Description string
	BasePrice   int64
	ImageURL    string
	IsAvailable bool
	SortOrder   int
	CreatedAt   time.Time
}
```

- [ ] **Step 3: Update ListByCategory scan in item_repo.go**

In `server/internal/repository/item_repo.go`, update the `ListByCategory` SELECT and Scan to include `created_at`:

```go
func (r *ItemRepo) ListByCategory(ctx context.Context, categoryID, storeID string, branchID *string) ([]model.Item, error) {
	query := `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order, created_at
		FROM items WHERE category_id = $1 AND store_id = $2
	`
	args := []any{categoryID, storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $3`
		args = append(args, *branchID)
	}
	query += ` ORDER BY sort_order`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var i model.Item
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}
```

Also update `GetByID` scan:

```go
func (r *ItemRepo) GetByID(ctx context.Context, id, storeID string) (*model.Item, error) {
	i := &model.Item{}
	err := r.db.QueryRow(ctx, `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order, created_at
		FROM items WHERE id = $1 AND store_id = $2
	`, id, storeID).Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder, &i.CreatedAt)
	if err != nil {
		return nil, err
	}
	return i, nil
}
```

- [ ] **Step 4: Verify build**

```bash
docker compose exec -T server go build ./...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add server/internal/model/item.go server/internal/repository/item_repo.go
git commit -m "feat(server): add created_at to Item model and queries"
```

---

## Task 5: Add DiscoverItem type + feed queries to ItemRepo

**Files:**
- Modify: `server/internal/repository/item_repo.go`

- [ ] **Step 1: Add DiscoverItem struct and new methods to item_repo.go**

Add to the bottom of `server/internal/repository/item_repo.go`:

```go
// DiscoverItem is the item-centric discovery type used by feed and paginated endpoints.
type DiscoverItem struct {
	ID                   string
	Name                 string
	Description          string
	ImageURL             string
	BasePrice            int64
	IsAvailable          bool
	CreatedAt            time.Time
	OrderCount           int
	HasRequiredModifiers bool
	BranchID             string
	BranchName           string
	BranchAddress        string
	Lat                  *float64
	Lng                  *float64
	StoreID              string
	StoreName            string
	StoreCategory        string
}

const discoverItemSelect = `
	SELECT
	  i.id, i.name, i.description, i.image_url, i.base_price, i.is_available, i.created_at,
	  COALESCE((
	    SELECT COUNT(oi2.id)
	    FROM order_items oi2
	    JOIN orders o2 ON o2.id = oi2.order_id
	    WHERE oi2.item_id = i.id
	    AND o2.status NOT IN ('cancelled', 'rejected')
	  ), 0) AS order_count,
	  EXISTS(
	    SELECT 1 FROM modifier_groups mg
	    WHERE mg.item_id = i.id AND mg.is_required = true
	  ) AS has_required_modifiers,
	  b.id, b.name, b.address, b.lat, b.lng,
	  s.id, s.name, s.category
	FROM items i
	JOIN branches b ON b.id = i.branch_id AND b.is_active = true
	JOIN stores s ON s.id = i.store_id AND s.is_active = true
	WHERE i.is_available = true
`

func scanDiscoverItem(rows interface {
	Scan(dest ...any) error
}) (DiscoverItem, error) {
	var d DiscoverItem
	err := rows.Scan(
		&d.ID, &d.Name, &d.Description, &d.ImageURL, &d.BasePrice, &d.IsAvailable, &d.CreatedAt,
		&d.OrderCount, &d.HasRequiredModifiers,
		&d.BranchID, &d.BranchName, &d.BranchAddress, &d.Lat, &d.Lng,
		&d.StoreID, &d.StoreName, &d.StoreCategory,
	)
	return d, err
}

// GetFeedSection returns up to limit items sorted by "new" (created_at desc) or "popular" (order_count desc).
func (r *ItemRepo) GetFeedSection(ctx context.Context, sort string, limit int) ([]DiscoverItem, error) {
	orderClause := "i.created_at DESC"
	if sort == "popular" {
		orderClause = "order_count DESC, i.created_at DESC"
	}

	query := discoverItemSelect + " GROUP BY i.id, b.id, s.id ORDER BY " + orderClause + " LIMIT $1"
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		d, err := scanDiscoverItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, nil
}

// ListForFeed returns paginated discover items with optional store category filter and sort.
func (r *ItemRepo) ListForFeed(ctx context.Context, category, sort string, page, limit int) ([]DiscoverItem, int, error) {
	orderClause := "i.created_at DESC"
	if sort == "popular" {
		orderClause = "order_count DESC, i.created_at DESC"
	}

	args := []any{}
	whereExtra := ""
	if category != "" {
		args = append(args, category)
		whereExtra = " AND s.category = $1"
	}

	countQuery := `
		SELECT COUNT(DISTINCT i.id)
		FROM items i
		JOIN branches b ON b.id = i.branch_id AND b.is_active = true
		JOIN stores s ON s.id = i.store_id AND s.is_active = true
		WHERE i.is_available = true
	` + whereExtra

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	pageArgs := append(args, limit, offset)
	pageArgOffset := len(args)

	query := discoverItemSelect + whereExtra +
		" GROUP BY i.id, b.id, s.id ORDER BY " + orderClause +
		" LIMIT $" + fmt.Sprintf("%d", pageArgOffset+1) +
		" OFFSET $" + fmt.Sprintf("%d", pageArgOffset+2)

	rows, err := r.db.Query(ctx, query, pageArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []DiscoverItem
	for rows.Next() {
		d, err := scanDiscoverItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, d)
	}
	return items, total, nil
}
```

Make sure `"fmt"` is in the import block of `item_repo.go`. The file already imports `"context"` and others — add `"fmt"` and `"time"` if not present.

- [ ] **Step 2: Verify build**

```bash
docker compose exec -T server go build ./...
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add server/internal/repository/item_repo.go
git commit -m "feat(server): add DiscoverItem type and feed query methods to ItemRepo"
```

---

## Task 6: Add GetByIDForBranch to ItemRepo

**Files:**
- Modify: `server/internal/repository/item_repo.go`

- [ ] **Step 1: Add GetByIDForBranch method**

Add to `server/internal/repository/item_repo.go` (after the existing `GetByID` method):

```go
// GetByIDForBranch fetches a single item by its ID scoped to a specific branch.
// Used by the public GET /items/:id endpoint which only knows branchID, not storeID.
func (r *ItemRepo) GetByIDForBranch(ctx context.Context, id, branchID string) (*model.Item, error) {
	i := &model.Item{}
	err := r.db.QueryRow(ctx, `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order, created_at
		FROM items WHERE id = $1 AND branch_id = $2
	`, id, branchID).Scan(
		&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description,
		&i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder, &i.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return i, nil
}
```

- [ ] **Step 2: Verify build**

```bash
docker compose exec -T server go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/repository/item_repo.go
git commit -m "feat(server): add GetByIDForBranch to ItemRepo for single-item endpoint"
```

---

## Task 7: Create DiscoverHandler

**Files:**
- Create: `server/internal/handler/discover_handler.go`

- [ ] **Step 1: Create the handler file**

Create `server/internal/handler/discover_handler.go`:

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/repository"
)

type DiscoverHandler struct {
	itemRepo *repository.ItemRepo
}

func NewDiscoverHandler(itemRepo *repository.ItemRepo) *DiscoverHandler {
	return &DiscoverHandler{itemRepo: itemRepo}
}

type discoverItemResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	ImageURL             string   `json:"image_url"`
	BasePrice            int64    `json:"base_price"`
	IsAvailable          bool     `json:"is_available"`
	CreatedAt            string   `json:"created_at"`
	OrderCount           int      `json:"order_count"`
	HasRequiredModifiers bool     `json:"has_required_modifiers"`
	BranchID             string   `json:"branch_id"`
	BranchName           string   `json:"branch_name"`
	BranchAddress        string   `json:"branch_address"`
	Lat                  *float64 `json:"lat"`
	Lng                  *float64 `json:"lng"`
	StoreID              string   `json:"store_id"`
	StoreName            string   `json:"store_name"`
	StoreCategory        string   `json:"store_category"`
}

func toDiscoverItemResponse(d repository.DiscoverItem) discoverItemResponse {
	return discoverItemResponse{
		ID:                   d.ID,
		Name:                 d.Name,
		Description:          d.Description,
		ImageURL:             d.ImageURL,
		BasePrice:            d.BasePrice,
		IsAvailable:          d.IsAvailable,
		CreatedAt:            d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		OrderCount:           d.OrderCount,
		HasRequiredModifiers: d.HasRequiredModifiers,
		BranchID:             d.BranchID,
		BranchName:           d.BranchName,
		BranchAddress:        d.BranchAddress,
		Lat:                  d.Lat,
		Lng:                  d.Lng,
		StoreID:              d.StoreID,
		StoreName:            d.StoreName,
		StoreCategory:        d.StoreCategory,
	}
}

// Feed returns pre-built curated sections: new arrivals and popular.
func (h *DiscoverHandler) Feed(c echo.Context) error {
	ctx := c.Request().Context()

	newItems, err := h.itemRepo.GetFeedSection(ctx, "new", 10)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	popularItems, err := h.itemRepo.GetFeedSection(ctx, "popular", 10)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	type sectionResponse struct {
		Title string                 `json:"title"`
		Type  string                 `json:"type"`
		Items []discoverItemResponse `json:"items"`
	}

	type feedResponse struct {
		Sections []sectionResponse `json:"sections"`
	}

	toResponses := func(items []repository.DiscoverItem) []discoverItemResponse {
		out := make([]discoverItemResponse, len(items))
		for i, item := range items {
			out[i] = toDiscoverItemResponse(item)
		}
		return out
	}

	return c.JSON(http.StatusOK, feedResponse{
		Sections: []sectionResponse{
			{Title: "New Arrivals", Type: "new", Items: toResponses(newItems)},
			{Title: "Popular Right Now", Type: "popular", Items: toResponses(popularItems)},
		},
	})
}

// Items returns paginated discover items with optional category and sort filters.
func (h *DiscoverHandler) Items(c echo.Context) error {
	ctx := c.Request().Context()

	category := c.QueryParam("category")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "new"
	}

	page := 1
	limit := 20
	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 50 {
			limit = v
		}
	}

	items, total, err := h.itemRepo.ListForFeed(ctx, category, sort, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	out := make([]discoverItemResponse, len(items))
	for i, item := range items {
		out[i] = toDiscoverItemResponse(item)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"items": out,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
```

- [ ] **Step 2: Verify build**

```bash
docker compose exec -T server go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/handler/discover_handler.go
git commit -m "feat(server): add DiscoverHandler with feed and items endpoints"
```

---

## Task 8: Create ItemHandler (GET /items/:id)

**Files:**
- Create: `server/internal/handler/item_handler.go`

- [ ] **Step 1: Create the handler file**

Create `server/internal/handler/item_handler.go`:

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/repository"
)

type ItemHandler struct {
	itemRepo          *repository.ItemRepo
	modifierGroupRepo *repository.ModifierGroupRepo
	branchRepo        *repository.BranchRepo
}

func NewItemHandler(
	itemRepo *repository.ItemRepo,
	modifierGroupRepo *repository.ModifierGroupRepo,
	branchRepo *repository.BranchRepo,
) *ItemHandler {
	return &ItemHandler{
		itemRepo:          itemRepo,
		modifierGroupRepo: modifierGroupRepo,
		branchRepo:        branchRepo,
	}
}

// GetByID handles GET /items/:id?branch=branchId
// Returns the item with modifier groups and branch context.
// This replaces the previous pattern of loading the full branch menu just to find one item.
func (h *ItemHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")
	branchID := c.QueryParam("branch")

	if branchID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "branch query param required")
	}

	detail, err := h.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "branch not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	item, err := h.itemRepo.GetByIDForBranch(ctx, itemID, branchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "item not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	groups, err := h.modifierGroupRepo.ListByItem(ctx, item.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	type modifierResponse struct {
		ID              string `json:"id"`
		ModifierGroupID string `json:"modifier_group_id"`
		Name            string `json:"name"`
		PriceAdjustment int64  `json:"price_adjustment"`
		IsAvailable     bool   `json:"is_available"`
		SortOrder       int    `json:"sort_order"`
	}
	type modifierGroupResponse struct {
		ID            string             `json:"id"`
		Name          string             `json:"name"`
		SelectionType string             `json:"selection_type"`
		IsRequired    bool               `json:"is_required"`
		MinSelections int                `json:"min_selections"`
		MaxSelections int                `json:"max_selections"`
		Modifiers     []modifierResponse `json:"modifiers"`
	}
	type itemResponse struct {
		ID             string                  `json:"id"`
		CategoryID     string                  `json:"category_id"`
		StoreID        string                  `json:"store_id"`
		BranchID       string                  `json:"branch_id"`
		Name           string                  `json:"name"`
		Description    string                  `json:"description"`
		BasePrice      int64                   `json:"base_price"`
		ImageURL       string                  `json:"image_url"`
		IsAvailable    bool                    `json:"is_available"`
		ModifierGroups []modifierGroupResponse `json:"modifier_groups"`
	}

	groupResponses := make([]modifierGroupResponse, len(groups))
	for gi, g := range groups {
		mods := make([]modifierResponse, len(g.Modifiers))
		for mi, m := range g.Modifiers {
			mods[mi] = modifierResponse{
				ID:              m.ID,
				ModifierGroupID: m.ModifierGroupID,
				Name:            m.Name,
				PriceAdjustment: m.PriceAdjustment,
				IsAvailable:     m.IsAvailable,
				SortOrder:       m.SortOrder,
			}
		}
		groupResponses[gi] = modifierGroupResponse{
			ID:            g.ID,
			Name:          g.Name,
			SelectionType: g.SelectionType,
			IsRequired:    g.IsRequired,
			MinSelections: g.MinSelections,
			MaxSelections: g.MaxSelections,
			Modifiers:     mods,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"item": itemResponse{
			ID:             item.ID,
			CategoryID:     item.CategoryID,
			StoreID:        item.StoreID,
			BranchID:       item.BranchID,
			Name:           item.Name,
			Description:    item.Description,
			BasePrice:      item.BasePrice,
			ImageURL:       item.ImageURL,
			IsAvailable:    item.IsAvailable,
			ModifierGroups: groupResponses,
		},
		"branch": detail,
	})
}
```

- [ ] **Step 2: Check ModifierGroupRepo.ListByItem signature**

Open `server/internal/repository/modifier_repo.go` and confirm `ListByItem(ctx context.Context, itemID string)` exists and returns `([]model.ModifierGroup, error)`. If the method signature differs, adjust the call in `item_handler.go` accordingly.

- [ ] **Step 3: Verify build**

```bash
docker compose exec -T server go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add server/internal/handler/item_handler.go
git commit -m "feat(server): add ItemHandler with GET /items/:id endpoint"
```

---

## Task 9: Wire new handlers into router and main

**Files:**
- Modify: `server/internal/handler/router.go`
- Modify: `server/cmd/server/main.go`

- [ ] **Step 1: Add new handlers to Handlers struct in router.go**

In `server/internal/handler/router.go`, update the `Handlers` struct:

```go
type Handlers struct {
	Auth     *AuthHandler
	Branch   *BranchHandler
	Staff    *StaffHandler
	Store    *StoreHandler
	Menu     *MenuHandler
	Order    *OrderHandler
	Discover *DiscoverHandler
	Item     *ItemHandler
}
```

- [ ] **Step 2: Register new routes in SetupRoutes**

In `SetupRoutes`, after the existing public routes block, add:

```go
e.GET("/discover/feed", h.Discover.Feed)
e.GET("/discover/items", h.Discover.Items)
e.GET("/items/:id", h.Item.GetByID)
```

These go in the public (unauthenticated) section, after:
```go
e.GET("/stores/:slug/menu", h.Store.GetMenu)
```

- [ ] **Step 3: Instantiate new handlers in main.go**

In `server/cmd/server/main.go`, find where handlers are created (the `h := &handler.Handlers{...}` block). Add:

```go
Discover: handler.NewDiscoverHandler(itemRepo),
Item:     handler.NewItemHandler(itemRepo, modifierGroupRepo, branchRepo),
```

`modifierGroupRepo` and `branchRepo` are already instantiated in main.go (they're used by existing handlers). Confirm their variable names by reading the relevant section of `main.go` and matching them.

- [ ] **Step 4: Verify build**

```bash
docker compose exec -T server go build ./...
```

- [ ] **Step 5: Test the new endpoints**

```bash
# Should return two sections with items
curl -s http://localhost:8080/discover/feed | head -100

# Should return paginated items
curl -s "http://localhost:8080/discover/items?sort=new&page=1&limit=5" | head -100

# Replace ITEM_ID and BRANCH_ID with real values from the feed response
curl -s "http://localhost:8080/items/ITEM_ID?branch=BRANCH_ID" | head -100
```

Expected: JSON responses with no errors.

- [ ] **Step 6: Commit**

```bash
git add server/internal/handler/router.go server/cmd/server/main.go
git commit -m "feat(server): wire DiscoverHandler and ItemHandler into router"
```

---

## Task 10: Add DiscoverItem + BranchCart to frontend types

**Files:**
- Modify: `web/src/types/index.ts`

- [ ] **Step 1: Add new types to types/index.ts**

Append to the end of `web/src/types/index.ts`:

```typescript
export interface DiscoverItem {
  id: string
  name: string
  description: string
  image_url: string
  base_price: number
  is_available: boolean
  created_at: string
  order_count: number
  has_required_modifiers: boolean
  branch_id: string
  branch_name: string
  branch_address: string
  lat?: number | null
  lng?: number | null
  store_id: string
  store_name: string
  store_category: StoreCategory
}

export interface FeedSection {
  title: string
  type: 'new' | 'popular'
  items: DiscoverItem[]
}

export interface FeedResponse {
  sections: FeedSection[]
}

export interface ItemsPageResponse {
  items: DiscoverItem[]
  total: number
  page: number
  limit: number
}

export interface BranchCart {
  branch: CartMeta
  items: CartItem[]
}

export interface ItemDetailResponse {
  item: MenuItem
  branch: BranchDetail
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
make typecheck
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/src/types/index.ts
git commit -m "feat(web): add DiscoverItem, BranchCart, FeedSection types"
```

---

## Task 11: Rewrite useCartStore to multi-cart v3

**Files:**
- Modify: `web/src/store/cart.ts`

- [ ] **Step 1: Replace cart.ts with v3 multi-cart implementation**

Replace the full contents of `web/src/store/cart.ts`:

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { BranchCart, CartItem, CartMeta } from '../types'

interface CartStore {
  carts: Record<string, BranchCart>  // keyed by branchId
  activeBranchId: string | null

  // Branch management
  setActiveBranch: (branchId: string) => void
  setBranch: (meta: CartMeta) => void  // kept for BranchPage compatibility

  // Item operations (scoped to active branch)
  addItem: (branchMeta: CartMeta, item: CartItem) => void
  removeItem: (branchId: string, index: number) => void
  updateQuantity: (branchId: string, index: number, quantity: number) => void
  clearCart: (branchId: string) => void
  clearAll: () => void

  // Computed (active branch)
  activeBranchTotal: () => number
  activeBranchCount: () => number
  activeCart: () => BranchCart | null

  // Computed (all carts)
  totalCartsCount: () => number

  // Legacy compat — used by BranchPage; sets active branch without clearing others
  total: () => number
  count: () => number
}

function recalculate(item: CartItem, quantity: number): CartItem {
  const modifierTotal = item.modifiers.reduce((sum, m) => sum + m.price, 0)
  return {
    ...item,
    quantity,
    totalPrice: (item.price + modifierTotal) * quantity,
  }
}

export const useCartStore = create<CartStore>()(
  persist(
    (set, get) => ({
      carts: {},
      activeBranchId: null,

      setActiveBranch: (branchId) => set({ activeBranchId: branchId }),

      setBranch: (meta) => {
        const { carts } = get()
        set({
          activeBranchId: meta.branchId,
          carts: carts[meta.branchId]
            ? {
                ...carts,
                [meta.branchId]: { ...carts[meta.branchId], branch: meta },
              }
            : {
                ...carts,
                [meta.branchId]: { branch: meta, items: [] },
              },
        })
      },

      addItem: (branchMeta, item) => {
        const { carts } = get()
        const existing = carts[branchMeta.branchId] ?? { branch: branchMeta, items: [] }
        set({
          activeBranchId: branchMeta.branchId,
          carts: {
            ...carts,
            [branchMeta.branchId]: {
              branch: branchMeta,
              items: [...existing.items, item],
            },
          },
        })
      },

      removeItem: (branchId, index) => {
        const { carts } = get()
        const cart = carts[branchId]
        if (!cart) return
        const items = cart.items.filter((_, i) => i !== index)
        if (items.length === 0) {
          const nextCarts = { ...carts }
          delete nextCarts[branchId]
          const nextActiveBranchId = get().activeBranchId === branchId
            ? (Object.keys(nextCarts)[0] ?? null)
            : get().activeBranchId
          set({ carts: nextCarts, activeBranchId: nextActiveBranchId })
        } else {
          set({ carts: { ...carts, [branchId]: { ...cart, items } } })
        }
      },

      updateQuantity: (branchId, index, quantity) => {
        const { carts } = get()
        const cart = carts[branchId]
        if (!cart) return
        set({
          carts: {
            ...carts,
            [branchId]: {
              ...cart,
              items: cart.items.map((item, i) =>
                i === index ? recalculate(item, quantity) : item,
              ),
            },
          },
        })
      },

      clearCart: (branchId) => {
        const { carts, activeBranchId } = get()
        const nextCarts = { ...carts }
        delete nextCarts[branchId]
        const nextActiveBranchId = activeBranchId === branchId
          ? (Object.keys(nextCarts)[0] ?? null)
          : activeBranchId
        set({ carts: nextCarts, activeBranchId: nextActiveBranchId })
      },

      clearAll: () => set({ carts: {}, activeBranchId: null }),

      activeBranchTotal: () => {
        const { carts, activeBranchId } = get()
        if (!activeBranchId) return 0
        return (carts[activeBranchId]?.items ?? []).reduce((sum, item) => sum + item.totalPrice, 0)
      },

      activeBranchCount: () => {
        const { carts, activeBranchId } = get()
        if (!activeBranchId) return 0
        return (carts[activeBranchId]?.items ?? []).reduce((sum, item) => sum + item.quantity, 0)
      },

      activeCart: () => {
        const { carts, activeBranchId } = get()
        if (!activeBranchId) return null
        return carts[activeBranchId] ?? null
      },

      totalCartsCount: () => Object.keys(get().carts).length,

      // Legacy compat for BranchPage (uses activeBranchId context)
      total: () => get().activeBranchTotal(),
      count: () => get().activeBranchCount(),
    }),
    { name: 'xpressgo-cart-v3' },
  ),
)
```

- [ ] **Step 2: Fix BranchPage cart calls**

Now that the store is updated, open `web/src/pages/BranchPage.tsx` and update:
- `cart.count()` → `cart.activeBranchCount()`
- `cart.total()` → `cart.activeBranchTotal()`

These are already in the Task 2 version of BranchPage — verify they match. The `setBranch` call at line where BranchDetail loads stays as-is (it now updates the carts map).

- [ ] **Step 3: Verify TypeScript**

```bash
make typecheck
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/src/store/cart.ts web/src/pages/BranchPage.tsx
git commit -m "feat(web): rewrite cart store to v3 multi-cart (one cart per branch)"
```

---

## Task 12: BranchConflictSheet component

**Files:**
- Create: `web/src/components/cart/BranchConflictSheet.tsx`

- [ ] **Step 1: Create the component**

Create `web/src/components/cart/BranchConflictSheet.tsx`:

```tsx
interface BranchConflictSheetProps {
  newBranchName: string
  onConfirm: () => void
  onCancel: () => void
}

export function BranchConflictSheet({ newBranchName, onConfirm, onCancel }: BranchConflictSheetProps) {
  return (
    <div className="fixed inset-0 z-40">
      <button
        type="button"
        aria-label="Cancel"
        onClick={onCancel}
        className="absolute inset-0 bg-[var(--xp-overlay)]"
      />
      <div className="absolute inset-x-0 bottom-0 rounded-t-[20px] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-5">
        <div className="mx-auto mb-5 h-1 w-8 rounded-full bg-[var(--tg-theme-hint-color)]/40" />
        <h2 className="text-[18px] font-bold">Add from {newBranchName}?</h2>
        <p className="mt-2 text-[14px] leading-6 text-[var(--tg-theme-hint-color)]">
          You already have items from another branch. They'll stay in a separate cart — you can place each order independently.
        </p>
        <div className="mt-6 flex flex-col gap-3 pb-4">
          <button
            type="button"
            onClick={onConfirm}
            className="flex h-[52px] w-full items-center justify-center rounded-[20px] bg-[var(--xp-brand)] text-sm font-semibold text-white"
          >
            Add to new cart
          </button>
          <button
            type="button"
            onClick={onCancel}
            className="flex h-[52px] w-full items-center justify-center rounded-[20px] border border-[var(--xp-border)] text-sm font-semibold"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/cart/BranchConflictSheet.tsx
git commit -m "feat(web): add BranchConflictSheet for multi-cart branch switching"
```

---

## Task 13: DiscoverItemCard component

**Files:**
- Create: `web/src/components/discovery/DiscoverItemCard.tsx`

- [ ] **Step 1: Create the component**

Create `web/src/components/discovery/DiscoverItemCard.tsx`:

```tsx
import { Minus, Plus } from 'lucide-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { formatPrice } from '../../lib/format'
import { calculateDistanceInKm, formatDistance } from '../../lib/distance'
import { useCartStore } from '../../store/cart'
import { BranchConflictSheet } from '../cart/BranchConflictSheet'
import type { CartItem, DiscoverItem } from '../../types'

interface DiscoverItemCardProps {
  item: DiscoverItem
  userLat: number
  userLng: number
}

function isNewItem(createdAt: string): boolean {
  const created = new Date(createdAt)
  const sevenDaysAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
  return created > sevenDaysAgo
}

export function DiscoverItemCard({ item, userLat, userLng }: DiscoverItemCardProps) {
  const navigate = useNavigate()
  const cart = useCartStore()
  const [showConflict, setShowConflict] = useState(false)
  const [pendingAdd, setPendingAdd] = useState(false)

  const branchMeta = {
    branchId: item.branch_id,
    branchName: item.branch_name,
    storeName: item.store_name,
    bannerImageUrl: '',
  }

  const activeCart = cart.activeCart()
  const itemCart = cart.carts[item.branch_id]
  const existingItems = itemCart?.items ?? []
  const itemCount = existingItems
    .filter((ci) => ci.itemId === item.id)
    .reduce((sum, ci) => sum + ci.quantity, 0)

  const distanceKm = calculateDistanceInKm(userLat, userLng, item.lat, item.lng)

  function buildCartItem(): CartItem {
    return {
      itemId: item.id,
      imageUrl: item.image_url,
      name: item.name,
      price: item.base_price,
      quantity: 1,
      modifiers: [],
      totalPrice: item.base_price,
    }
  }

  function handleAdd() {
    if (item.has_required_modifiers) {
      navigate(`/item/${item.id}?branch=${item.branch_id}`)
      return
    }

    // Case 1: Cart is empty or same branch is active
    if (!activeCart || activeCart.branch.branchId === item.branch_id) {
      cart.addItem(branchMeta, buildCartItem())
      return
    }

    // Case 2: Branch already has an existing cart (not active)
    if (cart.carts[item.branch_id]) {
      cart.addItem(branchMeta, buildCartItem())
      return
    }

    // Case 3: New branch — show conflict sheet
    setShowConflict(true)
    setPendingAdd(true)
  }

  function handleConflictConfirm() {
    setShowConflict(false)
    setPendingAdd(false)
    cart.addItem(branchMeta, buildCartItem())
  }

  function handleDecrement() {
    const idx = (itemCart?.items ?? []).findLastIndex((ci) => ci.itemId === item.id)
    if (idx !== -1) {
      cart.removeItem(item.branch_id, idx)
    }
  }

  const showStepper = itemCount > 0 && !item.has_required_modifiers

  return (
    <>
      <div
        className="xp-card cursor-pointer overflow-hidden"
        onClick={() => navigate(`/item/${item.id}?branch=${item.branch_id}`)}
      >
        <div className="relative aspect-square">
          <img
            src={item.image_url || 'https://placehold.co/300x300?text=Item'}
            alt={item.name}
            loading="lazy"
            className="h-full w-full object-cover"
          />
          {isNewItem(item.created_at) && (
            <span className="absolute left-2 top-2 rounded-full bg-emerald-500 px-2 py-0.5 text-[10px] font-semibold text-white">
              NEW
            </span>
          )}
          <div
            className="absolute bottom-2 right-2"
            onClick={(e) => e.stopPropagation()}
          >
            {showStepper ? (
              <div
                className="flex items-center overflow-hidden rounded-full bg-[var(--xp-brand)]"
                style={{ width: '88px', transition: 'width 200ms ease' }}
              >
                <button
                  type="button"
                  onClick={handleDecrement}
                  className="flex h-9 w-9 shrink-0 items-center justify-center text-white"
                  aria-label="Remove one"
                >
                  <Minus className="h-3.5 w-3.5" />
                </button>
                <span className="flex-1 text-center text-sm font-semibold text-white">
                  {itemCount}
                </span>
                <button
                  type="button"
                  onClick={handleAdd}
                  className="flex h-9 w-9 shrink-0 items-center justify-center text-white"
                  aria-label="Add one more"
                >
                  <Plus className="h-3.5 w-3.5" />
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={handleAdd}
                className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-brand)] text-white active:scale-110 transition-transform duration-150"
                aria-label={`Add ${item.name} to cart`}
              >
                <Plus className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>
        <div className="p-2 pb-3">
          <p className="line-clamp-2 text-[13px] font-semibold leading-5">{item.name}</p>
          <p className="mt-0.5 text-[13px] font-semibold text-[var(--xp-brand)]">
            {formatPrice(item.base_price)} UZS
          </p>
          <p className="mt-0.5 truncate text-[11px] text-[var(--tg-theme-hint-color)]">
            {item.store_name} · {formatDistance(distanceKm)}
          </p>
        </div>
      </div>

      {showConflict && pendingAdd && (
        <BranchConflictSheet
          newBranchName={item.branch_name}
          onConfirm={handleConflictConfirm}
          onCancel={() => { setShowConflict(false); setPendingAdd(false) }}
        />
      )}
    </>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/discovery/DiscoverItemCard.tsx
git commit -m "feat(web): add DiscoverItemCard with stepper morph and conflict handling"
```

---

## Task 14: Update CartBar for multi-cart

**Files:**
- Modify: `web/src/components/cart/CartBar.tsx`

- [ ] **Step 1: Update CartBar to show multi-cart badge**

Replace the full contents of `web/src/components/cart/CartBar.tsx`:

```tsx
import { formatPrice } from '../../lib/format'
import { Button } from '@/components/ui/button'

interface CartBarProps {
  count: number
  total: number
  totalCartsCount: number
  onOpen: () => void
}

export function CartBar({ count, total, totalCartsCount, onOpen }: CartBarProps) {
  return (
    <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
      <Button
        type="button"
        onClick={onOpen}
        className="mx-auto flex h-14 w-full max-w-[32rem] items-center justify-between rounded-[20px] px-5 text-left"
      >
        <span className="flex items-center gap-2 text-sm font-semibold">
          Cart ({count})
          {totalCartsCount > 1 && (
            <span className="rounded-full bg-white/20 px-2 py-0.5 text-[11px] font-semibold">
              {totalCartsCount} carts
            </span>
          )}
        </span>
        <span className="text-base font-semibold">{formatPrice(total)} UZS</span>
      </Button>
    </div>
  )
}
```

- [ ] **Step 2: Update CartBar usage in BranchPage**

In `web/src/pages/BranchPage.tsx`, update the `CartBar` usage to pass `totalCartsCount`:

```tsx
bottomBar={
  cartCount > 0 ? (
    <CartBar
      count={cartCount}
      total={cartTotal}
      totalCartsCount={cart.totalCartsCount()}
      onOpen={() => navigate('/cart')}
    />
  ) : null
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/cart/CartBar.tsx web/src/pages/BranchPage.tsx
git commit -m "feat(web): add multi-cart badge to CartBar"
```

---

## Task 15: CartPage multi-cart tabs

**Files:**
- Modify: `web/src/pages/CartPage.tsx`

- [ ] **Step 1: Replace CartPage with multi-cart tabs version**

Replace the full contents of `web/src/pages/CartPage.tsx`:

```tsx
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Minus, Plus, Trash2 } from 'lucide-react'
import { AppShell } from '../components/layout/AppShell'
import { MenuHeader } from '../components/menu/MenuHeader'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { Order } from '../types'

const ETA_OPTIONS = [5, 10, 15, 20, 30]

export default function CartPage() {
  const navigate = useNavigate()
  const cart = useCartStore()
  const [eta, setEta] = useState(15)
  const [loading, setLoading] = useState(false)

  const { carts, activeBranchId } = cart
  const branchIds = Object.keys(carts)

  const activeBranch = activeBranchId && carts[activeBranchId] ? carts[activeBranchId] : null
  const activeId = activeBranchId ?? branchIds[0] ?? null

  async function placeOrder() {
    if (!activeBranch || !activeId) return
    setLoading(true)
    try {
      const items = activeBranch.items.map((item) => ({
        item_id: item.itemId,
        item_name: item.name,
        item_price: item.price,
        quantity: item.quantity,
        modifiers: item.modifiers.map((m) => ({
          modifier_id: m.id,
          modifier_name: m.name,
          price_adjustment: m.price,
        })),
      }))
      const order = await api<Order>('/orders', {
        method: 'POST',
        body: JSON.stringify({
          branch_id: activeId,
          payment_method: 'cash',
          eta_minutes: eta,
          items,
        }),
      })
      cart.clearCart(activeId)
      navigate(`/order/${order.id}`)
    } finally {
      setLoading(false)
    }
  }

  if (branchIds.length === 0) {
    return (
      <AppShell header={<MenuHeader title="Cart" count={0} />}>
        <div className="flex min-h-[60vh] flex-col items-center justify-center gap-4 px-4 text-center">
          <p className="text-[18px] font-semibold">Your cart is empty</p>
          <p className="text-[14px] text-[var(--tg-theme-hint-color)]">
            Browse the menu and add something delicious
          </p>
          <button
            type="button"
            onClick={() => navigate('/')}
            className="mt-2 flex h-12 items-center justify-center rounded-[20px] bg-[var(--xp-brand)] px-8 text-sm font-semibold text-white"
          >
            Browse Menu
          </button>
        </div>
      </AppShell>
    )
  }

  return (
    <AppShell header={<MenuHeader title="Cart" count={0} />}>
      {/* Branch tabs — only shown when multiple carts */}
      {branchIds.length > 1 && (
        <div className="scrollbar-none flex gap-2 overflow-x-auto border-b border-[var(--xp-border)] px-4 py-3">
          {branchIds.map((branchId) => {
            const branchCart = carts[branchId]
            const isActive = branchId === activeId
            return (
              <button
                key={branchId}
                type="button"
                onClick={() => cart.setActiveBranch(branchId)}
                className={`xp-pill whitespace-nowrap px-4 text-[13px] font-medium ${
                  isActive
                    ? 'bg-[var(--xp-brand)] text-white'
                    : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                }`}
              >
                {branchCart.branch.branchName}
              </button>
            )
          })}
        </div>
      )}

      {activeBranch && activeId ? (
        <div className="px-4 pb-36 pt-4 space-y-3">
          {/* Items list */}
          <div className="xp-card divide-y divide-[var(--xp-border)] overflow-hidden">
            {activeBranch.items.map((item, index) => (
              <div key={`${item.itemId}-${index}`} className="flex gap-3 p-4">
                <img
                  src={item.imageUrl || 'https://placehold.co/80x80?text=Item'}
                  alt={item.name}
                  className="h-16 w-16 rounded-xl object-cover"
                />
                <div className="flex flex-1 flex-col gap-1">
                  <p className="text-[14px] font-semibold">{item.name}</p>
                  {item.modifiers.length > 0 && (
                    <p className="text-[12px] text-[var(--tg-theme-hint-color)]">
                      {item.modifiers.map((m) => m.name).join(', ')}
                    </p>
                  )}
                  <p className="text-[13px] font-semibold text-[var(--xp-brand)]">
                    {formatPrice(item.totalPrice)} UZS
                  </p>
                </div>
                <div className="flex flex-col items-end gap-2">
                  <button
                    type="button"
                    onClick={() => cart.removeItem(activeId, index)}
                    className="flex h-8 w-8 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
                  >
                    <Trash2 className="h-3.5 w-3.5 text-[var(--tg-theme-hint-color)]" />
                  </button>
                  <div className="flex items-center gap-1 rounded-full bg-[var(--xp-card-bg)] px-1">
                    <button
                      type="button"
                      onClick={() => {
                        if (item.quantity <= 1) {
                          cart.removeItem(activeId, index)
                        } else {
                          cart.updateQuantity(activeId, index, item.quantity - 1)
                        }
                      }}
                      className="flex h-7 w-7 items-center justify-center"
                    >
                      <Minus className="h-3 w-3" />
                    </button>
                    <span className="w-5 text-center text-[13px] font-semibold">{item.quantity}</span>
                    <button
                      type="button"
                      onClick={() => cart.updateQuantity(activeId, index, item.quantity + 1)}
                      className="flex h-7 w-7 items-center justify-center"
                    >
                      <Plus className="h-3 w-3" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* ETA selector */}
          <div className="xp-card p-4">
            <p className="mb-3 text-[14px] font-semibold">Ready in</p>
            <div className="flex gap-2 flex-wrap">
              {ETA_OPTIONS.map((option) => (
                <button
                  key={option}
                  type="button"
                  onClick={() => setEta(option)}
                  className={`xp-pill px-4 text-[13px] font-medium ${
                    eta === option
                      ? 'bg-[var(--xp-brand)] text-white'
                      : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                  }`}
                >
                  {option} min
                </button>
              ))}
            </div>
          </div>

          {/* Total */}
          <div className="xp-card flex items-center justify-between p-4">
            <span className="text-[14px] font-semibold">Total</span>
            <span className="text-[16px] font-bold text-[var(--xp-brand)]">
              {formatPrice(cart.activeBranchTotal())} UZS
            </span>
          </div>
        </div>
      ) : null}

      {/* Place order bar */}
      {activeBranch && activeId && (
        <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
          <button
            type="button"
            onClick={() => void placeOrder()}
            disabled={loading}
            className="flex h-[52px] w-full items-center justify-center rounded-[20px] bg-[var(--xp-brand)] text-sm font-semibold text-white disabled:opacity-60"
          >
            {loading ? 'Placing order…' : `Place Order · ${formatPrice(cart.activeBranchTotal())} UZS`}
          </button>
        </div>
      )}
    </AppShell>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/CartPage.tsx
git commit -m "feat(web): rewrite CartPage with multi-cart branch tabs"
```

---

## Task 16: Redesign ItemPage

**Files:**
- Modify: `web/src/pages/ItemPage.tsx`

- [ ] **Step 1: Replace ItemPage with redesigned version**

Replace the full contents of `web/src/pages/ItemPage.tsx`:

```tsx
import { ArrowLeft, Minus, Plus, ShoppingCart } from 'lucide-react'
import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ModifierGroupSelector } from '../components/menu/ModifierGroupSelector'
import { api } from '../lib/api'
import { formatPrice } from '../lib/format'
import { useCartStore } from '../store/cart'
import type { ItemDetailResponse, ModifierGroup } from '../types'

function ItemPageSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="aspect-[4/3] w-full bg-[var(--xp-card-bg)]" />
      <div className="px-4 pt-4 space-y-3">
        <div className="h-6 w-3/4 rounded-lg bg-[var(--xp-card-bg)]" />
        <div className="h-5 w-1/3 rounded-lg bg-[var(--xp-card-bg)]" />
        <div className="space-y-2">
          <div className="h-4 w-full rounded bg-[var(--xp-card-bg)]" />
          <div className="h-4 w-5/6 rounded bg-[var(--xp-card-bg)]" />
        </div>
      </div>
    </div>
  )
}

export default function ItemPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const branchId = searchParams.get('branch')
  const cart = useCartStore()
  const navigate = useNavigate()

  const [data, setData] = useState<ItemDetailResponse | null>(null)
  const [selectedModifiers, setSelectedModifiers] = useState<Record<string, string[]>>({})
  const [quantity, setQuantity] = useState(1)
  const [stickyVisible, setStickyVisible] = useState(false)
  const heroRef = useRef<HTMLImageElement>(null)

  useEffect(() => {
    if (!branchId || !id) return

    api<ItemDetailResponse>(`/items/${id}?branch=${branchId}`).then((nextData) => {
      setData(nextData)
      const nextSelections: Record<string, string[]> = {}
      nextData.item.modifier_groups.forEach((group) => {
        if (group.is_required && group.selection_type === 'single' && group.modifiers[0]) {
          nextSelections[group.id] = [group.modifiers[0].id]
        }
      })
      setSelectedModifiers(nextSelections)
    })
  }, [branchId, id])

  // Show sticky header when hero image scrolls out of view
  useEffect(() => {
    const hero = heroRef.current
    if (!hero) return
    const observer = new IntersectionObserver(
      ([entry]) => setStickyVisible(!entry.isIntersecting),
      { threshold: 0 },
    )
    observer.observe(hero)
    return () => observer.disconnect()
  }, [data])

  const total = useMemo(() => {
    if (!data) return 0
    const modifierTotal = Object.entries(selectedModifiers).reduce((sum, [, ids]) => {
      return sum + ids.reduce((innerSum, modifierId) => {
        const modifier = data.item.modifier_groups
          .flatMap((g) => g.modifiers)
          .find((m) => m.id === modifierId)
        return innerSum + (modifier?.price_adjustment ?? 0)
      }, 0)
    }, 0)
    return (data.item.base_price + modifierTotal) * quantity
  }, [data, quantity, selectedModifiers])

  function toggleModifier(group: ModifierGroup, modifierId: string) {
    setSelectedModifiers((current) => {
      const selected = current[group.id] ?? []
      if (group.selection_type === 'single') {
        return { ...current, [group.id]: [modifierId] }
      }
      return {
        ...current,
        [group.id]: selected.includes(modifierId)
          ? selected.filter((entry) => entry !== modifierId)
          : [...selected, modifierId],
      }
    })
  }

  function addToCart() {
    if (!data) return

    const branchMeta = {
      branchId: data.branch.branch.id,
      branchName: data.branch.branch.name,
      storeName: data.branch.store.name,
      bannerImageUrl: data.branch.branch.banner_image_url,
    }

    const modifiers = Object.entries(selectedModifiers).flatMap(([, ids]) =>
      ids
        .map((modifierId) =>
          data.item.modifier_groups
            .flatMap((g) => g.modifiers)
            .find((m) => m.id === modifierId),
        )
        .filter(Boolean)
        .map((modifier) => ({
          id: modifier!.id,
          name: modifier!.name,
          price: modifier!.price_adjustment,
        })),
    )

    cart.addItem(branchMeta, {
      itemId: data.item.id,
      imageUrl: data.item.image_url,
      name: data.item.name,
      price: data.item.base_price,
      quantity,
      modifiers,
      totalPrice: total,
    })

    navigate('/cart')
  }

  const cartCount = cart.activeBranchCount()

  return (
    <div className="min-h-dvh bg-[var(--tg-theme-bg-color)]">
      {/* Sticky header (appears when hero scrolls away) */}
      <div
        className="fixed inset-x-0 top-0 z-10 flex items-center gap-3 border-b border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)]/90 px-4 py-3 backdrop-blur-sm"
        style={{ opacity: stickyVisible ? 1 : 0, transition: 'opacity 150ms ease', pointerEvents: stickyVisible ? 'auto' : 'none' }}
      >
        <button
          type="button"
          onClick={() => navigate(-1)}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <p className="flex-1 truncate text-[15px] font-semibold">
          {data?.item.name ?? ''}
        </p>
        <button
          type="button"
          onClick={() => navigate('/cart')}
          className="relative flex h-9 w-9 items-center justify-center rounded-full bg-[var(--xp-card-bg)]"
        >
          <ShoppingCart className="h-4 w-4" />
          {cartCount > 0 && (
            <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-[var(--xp-brand)] text-[9px] font-bold text-white">
              {cartCount}
            </span>
          )}
        </button>
      </div>

      {!data ? (
        <ItemPageSkeleton />
      ) : (
        <>
          {/* Hero image with back button overlay */}
          <div className="relative">
            <img
              ref={heroRef}
              src={data.item.image_url || 'https://placehold.co/800x600?text=Item'}
              alt={data.item.name}
              className="aspect-[4/3] w-full object-cover"
            />
            {/* Gradient overlay */}
            <div className="absolute inset-x-0 bottom-0 h-20 bg-gradient-to-t from-[var(--tg-theme-bg-color)] to-transparent" />
            {/* Floating back button */}
            <button
              type="button"
              onClick={() => navigate(-1)}
              className="absolute left-4 top-4 z-30 flex h-11 w-11 items-center justify-center rounded-full bg-black/40 text-white backdrop-blur-sm"
            >
              <ArrowLeft className="h-5 w-5" />
            </button>
          </div>

          {/* Content */}
          <div className="px-4 pb-32 pt-4">
            <h1 className="text-[22px] font-bold">{data.item.name}</h1>
            <p className="mt-1 text-xl font-semibold text-[var(--xp-brand)]">
              {formatPrice(data.item.base_price)} UZS
            </p>
            {data.item.description && (
              <p className="mt-3 text-[15px] leading-6 text-[var(--tg-theme-hint-color)]">
                {data.item.description}
              </p>
            )}

            {data.item.modifier_groups.map((group) => (
              <ModifierGroupSelector
                key={group.id}
                group={group}
                selected={selectedModifiers[group.id] ?? []}
                onToggle={toggleModifier}
              />
            ))}
          </div>
        </>
      )}

      {/* Bottom action bar */}
      {data && (
        <div className="fixed inset-x-0 bottom-0 z-20 border-t border-[var(--xp-border)] bg-[var(--tg-theme-bg-color)] px-4 pb-safe pt-3">
          <div className="mx-auto flex max-w-[32rem] items-center gap-4">
            <div className="flex items-center gap-2 rounded-[20px] bg-[var(--xp-card-bg)] px-3 py-2">
              <button
                type="button"
                className="flex h-9 w-9 items-center justify-center rounded-full"
                onClick={() => setQuantity((v) => Math.max(1, v - 1))}
              >
                <Minus className="h-4 w-4" />
              </button>
              <span className="w-6 text-center text-xl font-semibold">{quantity}</span>
              <button
                type="button"
                className="flex h-9 w-9 items-center justify-center rounded-full"
                onClick={() => setQuantity((v) => v + 1)}
              >
                <Plus className="h-4 w-4" />
              </button>
            </div>
            <button
              type="button"
              onClick={addToCart}
              className="flex h-[52px] flex-1 items-center justify-center rounded-[20px] bg-[var(--xp-brand)] px-4 text-sm font-semibold text-white"
            >
              Add to Cart · {formatPrice(total)} UZS
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/ItemPage.tsx
git commit -m "feat(web): redesign ItemPage with back button, hero image, and dedicated endpoint"
```

---

## Task 17: Create useDiscoveryFeed and useDiscoveryItems hooks

**Files:**
- Create: `web/src/hooks/useDiscoveryFeed.ts`
- Create: `web/src/hooks/useDiscoveryItems.ts`

- [ ] **Step 1: Create useDiscoveryFeed.ts**

Create `web/src/hooks/useDiscoveryFeed.ts`:

```typescript
import { useEffect, useState } from 'react'
import { api } from '../lib/api'
import type { FeedResponse, FeedSection } from '../types'

export function useDiscoveryFeed() {
  const [sections, setSections] = useState<FeedSection[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let active = true
    api<FeedResponse>('/discover/feed')
      .then((data) => { if (active) setSections(data.sections) })
      .finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [])

  return { sections, loading }
}
```

- [ ] **Step 2: Create useDiscoveryItems.ts**

Create `web/src/hooks/useDiscoveryItems.ts`:

```typescript
import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../lib/api'
import type { DiscoverItem, ItemsPageResponse } from '../types'

export function useDiscoveryItems(category: string, sort: string) {
  const [items, setItems] = useState<DiscoverItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [hasMore, setHasMore] = useState(true)
  const [total, setTotal] = useState(0)
  const pageRef = useRef(1)
  const activeRef = useRef(true)

  // Reset on filter change
  useEffect(() => {
    activeRef.current = true
    pageRef.current = 1
    setItems([])
    setHasMore(true)
    setLoading(true)

    const params = new URLSearchParams({ page: '1', limit: '20', sort })
    if (category) params.set('category', category)

    api<ItemsPageResponse>(`/discover/items?${params}`)
      .then((data) => {
        if (!activeRef.current) return
        setItems(data.items)
        setTotal(data.total)
        setHasMore(data.items.length < data.total)
        pageRef.current = 2
      })
      .finally(() => { if (activeRef.current) setLoading(false) })

    return () => { activeRef.current = false }
  }, [category, sort])

  const loadMore = useCallback(() => {
    if (loadingMore || !hasMore) return

    setLoadingMore(true)
    const params = new URLSearchParams({ page: String(pageRef.current), limit: '20', sort })
    if (category) params.set('category', category)

    api<ItemsPageResponse>(`/discover/items?${params}`)
      .then((data) => {
        setItems((prev) => [...prev, ...data.items])
        setHasMore(items.length + data.items.length < data.total)
        pageRef.current += 1
      })
      .finally(() => setLoadingMore(false))
  }, [loadingMore, hasMore, category, sort, items.length])

  return { items, loading, loadingMore, hasMore, total, loadMore }
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 4: Commit**

```bash
git add web/src/hooks/useDiscoveryFeed.ts web/src/hooks/useDiscoveryItems.ts
git commit -m "feat(web): add useDiscoveryFeed and useDiscoveryItems hooks"
```

---

## Task 18: Rewrite HomePage

**Files:**
- Modify: `web/src/pages/HomePage.tsx`

- [ ] **Step 1: Read the current HomePage**

Read `web/src/pages/HomePage.tsx` fully before replacing — note any state or logic that should be preserved (map/list toggle, user location, BranchSheet, DiscoveryMap).

- [ ] **Step 2: Replace HomePage**

Replace the full contents of `web/src/pages/HomePage.tsx`:

```tsx
import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BranchSheet } from '../components/discovery/BranchSheet'
import { DiscoveryMap } from '../components/discovery/DiscoveryMap'
import { ViewToggle } from '../components/discovery/ViewToggle'
import { DiscoverItemCard } from '../components/discovery/DiscoverItemCard'
import { CartBar } from '../components/cart/CartBar'
import { useDiscoveryFeed } from '../hooks/useDiscoveryFeed'
import { useDiscoveryItems } from '../hooks/useDiscoveryItems'
import { useDiscovery } from '../hooks/useDiscovery'
import { useUserLocation } from '../hooks/useUserLocation'
import { useTelegramTheme } from '../hooks/useTelegramTheme'
import { useCartStore } from '../store/cart'
import type { DiscoverBranch } from '../types'

type ChipId = 'new' | 'popular' | 'bar' | 'cafe' | 'coffee' | 'restaurant' | 'fastfood'

const CHIPS: { id: ChipId; label: string }[] = [
  { id: 'new', label: 'New' },
  { id: 'popular', label: 'Popular' },
  { id: 'bar', label: 'Bars' },
  { id: 'cafe', label: 'Cafes' },
  { id: 'coffee', label: 'Coffee' },
  { id: 'restaurant', label: 'Restaurants' },
  { id: 'fastfood', label: 'Fast Food' },
]

function getSortAndCategory(chip: ChipId): { sort: string; category: string } {
  if (chip === 'new') return { sort: 'new', category: '' }
  if (chip === 'popular') return { sort: 'popular', category: '' }
  return { sort: 'new', category: chip }
}

function FeedSectionSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="mb-3 flex items-center justify-between px-4">
        <div className="h-5 w-28 rounded bg-[var(--xp-card-bg)]" />
        <div className="h-4 w-12 rounded bg-[var(--xp-card-bg)]" />
      </div>
      <div className="flex gap-3 overflow-x-hidden px-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="w-[160px] shrink-0">
            <div className="aspect-square w-full rounded-xl bg-[var(--xp-card-bg)]" />
            <div className="mt-2 h-4 w-3/4 rounded bg-[var(--xp-card-bg)]" />
            <div className="mt-1 h-3 w-1/2 rounded bg-[var(--xp-card-bg)]" />
          </div>
        ))}
      </div>
    </div>
  )
}

function GridSkeleton() {
  return (
    <div className="grid grid-cols-2 gap-3 px-4 animate-pulse">
      {[1, 2, 3, 4].map((i) => (
        <div key={i} className="aspect-square rounded-xl bg-[var(--xp-card-bg)]" />
      ))}
    </div>
  )
}

export default function HomePage() {
  useTelegramTheme()
  const navigate = useNavigate()
  const { lat, lng } = useUserLocation()
  const cart = useCartStore()

  const [activeChip, setActiveChip] = useState<ChipId>('new')
  const [view, setView] = useState<'list' | 'map'>('list')
  const [selectedBranch, setSelectedBranch] = useState<(DiscoverBranch & { distanceKm: number }) | null>(null)

  const { sort, category } = getSortAndCategory(activeChip)
  const { sections, loading: feedLoading } = useDiscoveryFeed()
  const { items, loading: itemsLoading, loadingMore, hasMore, loadMore } = useDiscoveryItems(category, sort)
  const { branches } = useDiscovery(lat, lng)

  // IntersectionObserver for infinite scroll
  const sentinelRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      ([entry]) => { if (entry.isIntersecting) loadMore() },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [loadMore])

  const cartCount = cart.activeBranchCount()
  const cartTotal = cart.activeBranchTotal()

  // Greeting
  const hour = new Date().getHours()
  const greeting = hour < 12 ? 'Good morning' : hour < 18 ? 'Good afternoon' : 'Good evening'

  return (
    <div className="min-h-dvh bg-[var(--tg-theme-bg-color)]">
      {/* Map view */}
      {view === 'map' && (
        <div className="fixed inset-0 z-0">
          <DiscoveryMap
            branches={branches}
            center={{ lat, lng }}
            selectedBranchId={selectedBranch?.branch_id ?? null}
            visible={true}
            onSelect={(branch) => setSelectedBranch(branch)}
          />
        </div>
      )}

      {/* List view */}
      {view === 'list' && (
        <div className="flex flex-col pb-28">
          {/* Greeting header */}
          <div className="px-4 pb-4 pt-6">
            <h1 className="text-[22px] font-bold">{greeting} 👋</h1>
            <p className="mt-1 text-[14px] text-[var(--tg-theme-hint-color)]">What are you craving?</p>
          </div>

          {/* Sticky chip row */}
          <div className="sticky top-0 z-10 bg-[var(--tg-theme-bg-color)]">
            <div className="scrollbar-none flex gap-2 overflow-x-auto px-4 py-2">
              {CHIPS.map((chip) => (
                <button
                  key={chip.id}
                  type="button"
                  onClick={() => setActiveChip(chip.id)}
                  className={`xp-pill shrink-0 whitespace-nowrap px-4 text-[13px] font-medium ${
                    activeChip === chip.id
                      ? 'bg-[var(--xp-brand)] text-white'
                      : 'bg-[var(--xp-card-bg)] text-[var(--tg-theme-hint-color)]'
                  }`}
                >
                  {chip.label}
                </button>
              ))}
            </div>
          </div>

          {/* Curated sections — only shown when "new" or "popular" chip active */}
          {(activeChip === 'new' || activeChip === 'popular') && (
            <div className="mt-2 space-y-6">
              {feedLoading
                ? [1, 2].map((i) => <FeedSectionSkeleton key={i} />)
                : sections.map((section) => (
                    <div key={section.type}>
                      <div className="mb-3 flex items-center justify-between px-4">
                        <h2 className="text-[15px] font-semibold">{section.title}</h2>
                        <button
                          type="button"
                          className="text-[13px] text-[var(--xp-brand)]"
                          onClick={() => {}}
                        >
                          See all
                        </button>
                      </div>
                      <div className="scrollbar-none flex gap-3 overflow-x-auto px-4 pb-1">
                        {section.items.map((item) => (
                          <div key={item.id} className="w-[160px] shrink-0">
                            <DiscoverItemCard item={item} userLat={lat} userLng={lng} />
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}

              {/* Divider before vertical feed */}
              {!feedLoading && <div className="h-px bg-[var(--xp-border)] mx-4" />}
            </div>
          )}

          {/* Vertical item grid */}
          <div className="mt-4 px-4">
            <h2 className="mb-3 text-[15px] font-semibold">
              {activeChip === 'new' && 'All Items'}
              {activeChip === 'popular' && 'All Popular'}
              {!['new', 'popular'].includes(activeChip) && CHIPS.find((c) => c.id === activeChip)?.label}
            </h2>
            {itemsLoading ? (
              <GridSkeleton />
            ) : (
              <div className="grid grid-cols-2 gap-3">
                {items.map((item) => (
                  <DiscoverItemCard key={item.id} item={item} userLat={lat} userLng={lng} />
                ))}
              </div>
            )}
          </div>

          {/* Infinite scroll sentinel */}
          <div ref={sentinelRef} className="h-4" />
          {loadingMore && (
            <div className="flex justify-center py-4">
              <div className="h-5 w-5 animate-spin rounded-full border-2 border-[var(--xp-brand)] border-t-transparent" />
            </div>
          )}
          {!hasMore && items.length > 0 && (
            <p className="py-4 text-center text-[13px] text-[var(--tg-theme-hint-color)]">
              You've seen everything
            </p>
          )}
        </div>
      )}

      {/* BranchSheet for map selections */}
      {view === 'map' && selectedBranch && (
        <BranchSheet
          branch={selectedBranch}
          onClose={() => setSelectedBranch(null)}
        />
      )}

      {/* View toggle */}
      <ViewToggle value={view} onChange={setView} />

      {/* Cart bar */}
      {cartCount > 0 && (
        <CartBar
          count={cartCount}
          total={cartTotal}
          totalCartsCount={cart.totalCartsCount()}
          onOpen={() => navigate('/cart')}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
make typecheck
```

- [ ] **Step 4: Run full quality check**

```bash
make quality
```

Expected: all checks pass. Fix any lint errors before committing.

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/HomePage.tsx
git commit -m "feat(web): rewrite HomePage with item-first feed, curated sections, and infinite scroll"
```

---

## Self-Review

**Spec coverage check:**

| Spec requirement | Task |
|---|---|
| z-index scale + ViewToggle/BranchSheet fix | Task 1 |
| BranchPage skeleton + category fix | Task 2 |
| OrderPage skeleton | Task 3 |
| `created_at` on items | Task 4 |
| `GET /discover/feed` endpoint | Tasks 5, 7, 9 |
| `GET /discover/items` endpoint | Tasks 5, 7, 9 |
| `GET /items/:id?branch=` endpoint | Tasks 6, 8, 9 |
| DiscoverItem + BranchCart types | Task 10 |
| Multi-cart store v3 | Task 11 |
| BranchConflictSheet | Task 12 |
| DiscoverItemCard (stepper, NEW badge, conflict routing) | Task 13 |
| CartBar multi-cart badge | Task 14 |
| CartPage multi-cart tabs | Task 15 |
| ItemPage back button + hero + sticky header + new endpoint | Task 16 |
| useDiscoveryFeed + useDiscoveryItems hooks | Task 17 |
| HomePage greeting + chips + sections + vertical feed + infinite scroll | Task 18 |
| Animation: scale pulse on add | Task 13 (active:scale-110 on button) |
| Animation: stepper morph width | Task 13 (width transition on stepper div) |
| Animation: cart bar slide-up | Not implemented — CartBar is already fixed-position; slide-in can be added as a follow-up |
| `prefers-reduced-motion` | Not implemented — add `@media (prefers-reduced-motion: no-preference)` wrapper around transitions as a follow-up |

**Notes:**
- Tasks 2 and 11 are coupled: BranchPage uses `activeBranchCount()`/`activeBranchTotal()` which only exist after Task 11. TypeScript will error until both are complete.
- Task 18's `DiscoveryMap` props interface must match — verify the component accepts the same props as before. If `DiscoveryMap` uses `DiscoverBranch[]` for `branches`, keep using `useDiscovery` for the map view (which it does).
- `ModifierGroupRepo` constructor: confirm the variable name in `main.go` before Task 9 Step 3.
