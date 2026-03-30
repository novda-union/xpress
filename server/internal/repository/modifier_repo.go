package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type ModifierGroupRepo struct {
	db *pgxpool.Pool
}

func NewModifierGroupRepo(db *pgxpool.Pool) *ModifierGroupRepo {
	return &ModifierGroupRepo{db: db}
}

func (r *ModifierGroupRepo) ListByItem(ctx context.Context, itemID, storeID string, branchID *string) ([]model.ModifierGroup, error) {
	query := `
		SELECT id, item_id, store_id, branch_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		FROM modifier_groups WHERE item_id = $1 AND store_id = $2
	`
	args := []any{itemID, storeID}
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

	var groups []model.ModifierGroup
	for rows.Next() {
		var g model.ModifierGroup
		if err := rows.Scan(&g.ID, &g.ItemID, &g.StoreID, &g.BranchID, &g.Name, &g.SelectionType, &g.IsRequired, &g.MinSelections, &g.MaxSelections, &g.SortOrder); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *ModifierGroupRepo) Create(ctx context.Context, g *model.ModifierGroup) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO modifier_groups (item_id, store_id, branch_id, name, selection_type, is_required, min_selections, max_selections, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id
	`, g.ItemID, g.StoreID, g.BranchID, g.Name, g.SelectionType, g.IsRequired, g.MinSelections, g.MaxSelections, g.SortOrder).Scan(&g.ID)
}

func (r *ModifierGroupRepo) Update(ctx context.Context, g *model.ModifierGroup) error {
	_, err := r.db.Exec(ctx, `
		UPDATE modifier_groups SET name=$2, selection_type=$3, is_required=$4, min_selections=$5, max_selections=$6, sort_order=$7, branch_id=$8
		WHERE id=$1 AND store_id=$9
	`, g.ID, g.Name, g.SelectionType, g.IsRequired, g.MinSelections, g.MaxSelections, g.SortOrder, g.BranchID, g.StoreID)
	return err
}

func (r *ModifierGroupRepo) Delete(ctx context.Context, id, storeID string, branchID *string) error {
	query := `DELETE FROM modifier_groups WHERE id=$1 AND store_id=$2`
	args := []any{id, storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $3`
		args = append(args, *branchID)
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *ModifierGroupRepo) RequireOwnedByBranch(ctx context.Context, id, storeID, branchID string) (*model.ModifierGroup, error) {
	group := &model.ModifierGroup{}
	err := r.db.QueryRow(ctx, `
		SELECT id, item_id, store_id, branch_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		FROM modifier_groups
		WHERE id = $1 AND store_id = $2 AND branch_id = $3
	`, id, storeID, branchID).Scan(
		&group.ID, &group.ItemID, &group.StoreID, &group.BranchID, &group.Name, &group.SelectionType, &group.IsRequired, &group.MinSelections, &group.MaxSelections, &group.SortOrder,
	)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// ModifierRepo

type ModifierRepo struct {
	db *pgxpool.Pool
}

func NewModifierRepo(db *pgxpool.Pool) *ModifierRepo {
	return &ModifierRepo{db: db}
}

func (r *ModifierRepo) ListByGroup(ctx context.Context, groupID, storeID string, branchID *string) ([]model.Modifier, error) {
	query := `
		SELECT id, modifier_group_id, store_id, branch_id, name, price_adjustment, is_available, sort_order
		FROM modifiers WHERE modifier_group_id = $1 AND store_id = $2
	`
	args := []any{groupID, storeID}
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

	var mods []model.Modifier
	for rows.Next() {
		var m model.Modifier
		if err := rows.Scan(&m.ID, &m.ModifierGroupID, &m.StoreID, &m.BranchID, &m.Name, &m.PriceAdjustment, &m.IsAvailable, &m.SortOrder); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, nil
}

func (r *ModifierRepo) Create(ctx context.Context, m *model.Modifier) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO modifiers (modifier_group_id, store_id, branch_id, name, price_adjustment, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, m.ModifierGroupID, m.StoreID, m.BranchID, m.Name, m.PriceAdjustment, m.SortOrder).Scan(&m.ID)
}

func (r *ModifierRepo) Update(ctx context.Context, m *model.Modifier) error {
	_, err := r.db.Exec(ctx, `
		UPDATE modifiers SET name=$2, price_adjustment=$3, is_available=$4, sort_order=$5, branch_id=$6
		WHERE id=$1 AND store_id=$7
	`, m.ID, m.Name, m.PriceAdjustment, m.IsAvailable, m.SortOrder, m.BranchID, m.StoreID)
	return err
}

func (r *ModifierRepo) Delete(ctx context.Context, id, storeID string, branchID *string) error {
	query := `DELETE FROM modifiers WHERE id=$1 AND store_id=$2`
	args := []any{id, storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $3`
		args = append(args, *branchID)
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *ModifierRepo) RequireOwnedByBranch(ctx context.Context, id, storeID, branchID string) (*model.Modifier, error) {
	mod := &model.Modifier{}
	err := r.db.QueryRow(ctx, `
		SELECT id, modifier_group_id, store_id, branch_id, name, price_adjustment, is_available, sort_order
		FROM modifiers
		WHERE id = $1 AND store_id = $2 AND branch_id = $3
	`, id, storeID, branchID).Scan(
		&mod.ID, &mod.ModifierGroupID, &mod.StoreID, &mod.BranchID, &mod.Name, &mod.PriceAdjustment, &mod.IsAvailable, &mod.SortOrder,
	)
	if err != nil {
		return nil, err
	}
	return mod, nil
}
