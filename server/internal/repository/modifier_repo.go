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

func (r *ModifierGroupRepo) ListByItem(ctx context.Context, itemID, storeID string) ([]model.ModifierGroup, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, item_id, store_id, name, selection_type, is_required, min_selections, max_selections, sort_order
		FROM modifier_groups WHERE item_id = $1 AND store_id = $2
		ORDER BY sort_order
	`, itemID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.ModifierGroup
	for rows.Next() {
		var g model.ModifierGroup
		if err := rows.Scan(&g.ID, &g.ItemID, &g.StoreID, &g.Name, &g.SelectionType, &g.IsRequired, &g.MinSelections, &g.MaxSelections, &g.SortOrder); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *ModifierGroupRepo) Create(ctx context.Context, g *model.ModifierGroup) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO modifier_groups (item_id, store_id, name, selection_type, is_required, min_selections, max_selections, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`, g.ItemID, g.StoreID, g.Name, g.SelectionType, g.IsRequired, g.MinSelections, g.MaxSelections, g.SortOrder).Scan(&g.ID)
}

func (r *ModifierGroupRepo) Update(ctx context.Context, g *model.ModifierGroup) error {
	_, err := r.db.Exec(ctx, `
		UPDATE modifier_groups SET name=$2, selection_type=$3, is_required=$4, min_selections=$5, max_selections=$6, sort_order=$7
		WHERE id=$1 AND store_id=$8
	`, g.ID, g.Name, g.SelectionType, g.IsRequired, g.MinSelections, g.MaxSelections, g.SortOrder, g.StoreID)
	return err
}

func (r *ModifierGroupRepo) Delete(ctx context.Context, id, storeID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM modifier_groups WHERE id=$1 AND store_id=$2`, id, storeID)
	return err
}

// ModifierRepo

type ModifierRepo struct {
	db *pgxpool.Pool
}

func NewModifierRepo(db *pgxpool.Pool) *ModifierRepo {
	return &ModifierRepo{db: db}
}

func (r *ModifierRepo) ListByGroup(ctx context.Context, groupID, storeID string) ([]model.Modifier, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, modifier_group_id, store_id, name, price_adjustment, is_available, sort_order
		FROM modifiers WHERE modifier_group_id = $1 AND store_id = $2
		ORDER BY sort_order
	`, groupID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mods []model.Modifier
	for rows.Next() {
		var m model.Modifier
		if err := rows.Scan(&m.ID, &m.ModifierGroupID, &m.StoreID, &m.Name, &m.PriceAdjustment, &m.IsAvailable, &m.SortOrder); err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, nil
}

func (r *ModifierRepo) Create(ctx context.Context, m *model.Modifier) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO modifiers (modifier_group_id, store_id, name, price_adjustment, sort_order)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, m.ModifierGroupID, m.StoreID, m.Name, m.PriceAdjustment, m.SortOrder).Scan(&m.ID)
}

func (r *ModifierRepo) Update(ctx context.Context, m *model.Modifier) error {
	_, err := r.db.Exec(ctx, `
		UPDATE modifiers SET name=$2, price_adjustment=$3, is_available=$4, sort_order=$5
		WHERE id=$1 AND store_id=$6
	`, m.ID, m.Name, m.PriceAdjustment, m.IsAvailable, m.SortOrder, m.StoreID)
	return err
}

func (r *ModifierRepo) Delete(ctx context.Context, id, storeID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM modifiers WHERE id=$1 AND store_id=$2`, id, storeID)
	return err
}
