package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type MenuRepo struct {
	db *pgxpool.Pool
}

func NewMenuRepo(db *pgxpool.Pool) *MenuRepo {
	return &MenuRepo{db: db}
}

// GetFullMenu loads the complete nested menu for a store in a few queries.
// This keeps the existing admin store-scoped path working.
func (r *MenuRepo) GetFullMenu(ctx context.Context, storeID string) (*model.Menu, error) {
	// 1. Categories
	catRows, err := r.db.Query(ctx, `
		SELECT id, store_id, name, sort_order, is_active
		FROM categories WHERE store_id = $1 AND is_active = true
		ORDER BY sort_order
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	catMap := make(map[string]*model.MenuCategory)
	var menu model.Menu

	for catRows.Next() {
		var c model.Category
		if err := catRows.Scan(&c.ID, &c.StoreID, &c.Name, &c.SortOrder, &c.IsActive); err != nil {
			return nil, err
		}
		mc := model.MenuCategory{Category: c}
		menu.Categories = append(menu.Categories, mc)
		catMap[c.ID] = &menu.Categories[len(menu.Categories)-1]
	}

	if len(menu.Categories) == 0 {
		menu.Categories = []model.MenuCategory{}
		return &menu, nil
	}

	// 2. Items
	itemRows, err := r.db.Query(ctx, `
		SELECT id, category_id, store_id, name, description, base_price, image_url, is_available, sort_order
		FROM items WHERE store_id = $1 AND is_available = true
		ORDER BY sort_order
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	itemMap := make(map[string]*model.MenuItem)
	for itemRows.Next() {
		var i model.Item
		if err := itemRows.Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder); err != nil {
			return nil, err
		}
		if cat, ok := catMap[i.CategoryID]; ok {
			mi := model.MenuItem{Item: i}
			cat.Items = append(cat.Items, mi)
			itemMap[i.ID] = &cat.Items[len(cat.Items)-1]
		}
	}

	// 3. Modifier groups
	mgRows, err := r.db.Query(ctx, `
		SELECT id, item_id, store_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		FROM modifier_groups WHERE store_id = $1
		ORDER BY sort_order
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer mgRows.Close()

	mgMap := make(map[string]*model.ModifierGroup)
	for mgRows.Next() {
		var g model.ModifierGroup
		if err := mgRows.Scan(&g.ID, &g.ItemID, &g.StoreID, &g.Name, &g.SelectionType, &g.IsRequired, &g.MinSelections, &g.MaxSelections, &g.SortOrder); err != nil {
			return nil, err
		}
		if item, ok := itemMap[g.ItemID]; ok {
			item.ModifierGroups = append(item.ModifierGroups, g)
			mgMap[g.ID] = &item.ModifierGroups[len(item.ModifierGroups)-1]
		}
	}

	// 4. Modifiers
	modRows, err := r.db.Query(ctx, `
		SELECT id, modifier_group_id, store_id, name, price_adjustment, is_available, sort_order
		FROM modifiers WHERE store_id = $1 AND is_available = true
		ORDER BY sort_order
	`, storeID)
	if err != nil {
		return nil, err
	}
	defer modRows.Close()

	for modRows.Next() {
		var m model.Modifier
		if err := modRows.Scan(&m.ID, &m.ModifierGroupID, &m.StoreID, &m.Name, &m.PriceAdjustment, &m.IsAvailable, &m.SortOrder); err != nil {
			return nil, err
		}
		if mg, ok := mgMap[m.ModifierGroupID]; ok {
			mg.Modifiers = append(mg.Modifiers, m)
		}
	}

	// Ensure empty slices instead of nil
	for ci := range menu.Categories {
		if menu.Categories[ci].Items == nil {
			menu.Categories[ci].Items = []model.MenuItem{}
		}
		for ii := range menu.Categories[ci].Items {
			if menu.Categories[ci].Items[ii].ModifierGroups == nil {
				menu.Categories[ci].Items[ii].ModifierGroups = []model.ModifierGroup{}
			}
			for gi := range menu.Categories[ci].Items[ii].ModifierGroups {
				if menu.Categories[ci].Items[ii].ModifierGroups[gi].Modifiers == nil {
					menu.Categories[ci].Items[ii].ModifierGroups[gi].Modifiers = []model.Modifier{}
				}
			}
		}
	}

	return &menu, nil
}

// GetFullMenuByBranch loads the complete nested menu for a single branch.
func (r *MenuRepo) GetFullMenuByBranch(ctx context.Context, branchID string) (*model.Menu, error) {
	catRows, err := r.db.Query(ctx, `
		SELECT id, store_id, branch_id, name, sort_order, is_active
		FROM categories WHERE branch_id = $1 AND is_active = true
		ORDER BY sort_order
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	catMap := make(map[string]*model.MenuCategory)
	var menu model.Menu

	for catRows.Next() {
		var c model.Category
		if err := catRows.Scan(&c.ID, &c.StoreID, &c.BranchID, &c.Name, &c.SortOrder, &c.IsActive); err != nil {
			return nil, err
		}
		mc := model.MenuCategory{Category: c}
		menu.Categories = append(menu.Categories, mc)
		catMap[c.ID] = &menu.Categories[len(menu.Categories)-1]
	}

	if len(menu.Categories) == 0 {
		menu.Categories = []model.MenuCategory{}
		return &menu, nil
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order
		FROM items WHERE branch_id = $1 AND is_available = true
		ORDER BY sort_order
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	itemMap := make(map[string]*model.MenuItem)
	for itemRows.Next() {
		var i model.Item
		if err := itemRows.Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder); err != nil {
			return nil, err
		}
		if cat, ok := catMap[i.CategoryID]; ok {
			mi := model.MenuItem{Item: i}
			cat.Items = append(cat.Items, mi)
			itemMap[i.ID] = &cat.Items[len(cat.Items)-1]
		}
	}

	mgRows, err := r.db.Query(ctx, `
		SELECT id, item_id, store_id, branch_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		FROM modifier_groups WHERE branch_id = $1
		ORDER BY sort_order
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer mgRows.Close()

	mgMap := make(map[string]*model.ModifierGroup)
	for mgRows.Next() {
		var g model.ModifierGroup
		if err := mgRows.Scan(&g.ID, &g.ItemID, &g.StoreID, &g.BranchID, &g.Name, &g.SelectionType, &g.IsRequired, &g.MinSelections, &g.MaxSelections, &g.SortOrder); err != nil {
			return nil, err
		}
		if item, ok := itemMap[g.ItemID]; ok {
			item.ModifierGroups = append(item.ModifierGroups, g)
			mgMap[g.ID] = &item.ModifierGroups[len(item.ModifierGroups)-1]
		}
	}

	modRows, err := r.db.Query(ctx, `
		SELECT id, modifier_group_id, store_id, branch_id, name, price_adjustment, is_available, sort_order
		FROM modifiers WHERE branch_id = $1 AND is_available = true
		ORDER BY sort_order
	`, branchID)
	if err != nil {
		return nil, err
	}
	defer modRows.Close()

	for modRows.Next() {
		var m model.Modifier
		if err := modRows.Scan(&m.ID, &m.ModifierGroupID, &m.StoreID, &m.BranchID, &m.Name, &m.PriceAdjustment, &m.IsAvailable, &m.SortOrder); err != nil {
			return nil, err
		}
		if mg, ok := mgMap[m.ModifierGroupID]; ok {
			mg.Modifiers = append(mg.Modifiers, m)
		}
	}

	for ci := range menu.Categories {
		if menu.Categories[ci].Items == nil {
			menu.Categories[ci].Items = []model.MenuItem{}
		}
		for ii := range menu.Categories[ci].Items {
			if menu.Categories[ci].Items[ii].ModifierGroups == nil {
				menu.Categories[ci].Items[ii].ModifierGroups = []model.ModifierGroup{}
			}
			for gi := range menu.Categories[ci].Items[ii].ModifierGroups {
				if menu.Categories[ci].Items[ii].ModifierGroups[gi].Modifiers == nil {
					menu.Categories[ci].Items[ii].ModifierGroups[gi].Modifiers = []model.Modifier{}
				}
			}
		}
	}

	return &menu, nil
}
