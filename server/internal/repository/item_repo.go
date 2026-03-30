package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type ItemRepo struct {
	db *pgxpool.Pool
}

func NewItemRepo(db *pgxpool.Pool) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) ListByCategory(ctx context.Context, categoryID, storeID string, branchID *string) ([]model.Item, error) {
	query := `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order
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
		if err := rows.Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (r *ItemRepo) GetByID(ctx context.Context, id, storeID string) (*model.Item, error) {
	i := &model.Item{}
	err := r.db.QueryRow(ctx, `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order
		FROM items WHERE id = $1 AND store_id = $2
	`, id, storeID).Scan(&i.ID, &i.CategoryID, &i.StoreID, &i.BranchID, &i.Name, &i.Description, &i.BasePrice, &i.ImageURL, &i.IsAvailable, &i.SortOrder)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (r *ItemRepo) Create(ctx context.Context, i *model.Item) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO items (category_id, store_id, branch_id, name, description, base_price, image_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`, i.CategoryID, i.StoreID, i.BranchID, i.Name, i.Description, i.BasePrice, i.ImageURL, i.SortOrder).Scan(&i.ID)
}

func (r *ItemRepo) Update(ctx context.Context, i *model.Item) error {
	_, err := r.db.Exec(ctx, `
		UPDATE items SET name=$2, description=$3, base_price=$4, image_url=$5, is_available=$6, sort_order=$7, category_id=$8, branch_id=$9
		WHERE id=$1 AND store_id=$10
	`, i.ID, i.Name, i.Description, i.BasePrice, i.ImageURL, i.IsAvailable, i.SortOrder, i.CategoryID, i.BranchID, i.StoreID)
	return err
}

func (r *ItemRepo) Delete(ctx context.Context, id, storeID string, branchID *string) error {
	query := `DELETE FROM items WHERE id=$1 AND store_id=$2`
	args := []any{id, storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $3`
		args = append(args, *branchID)
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *ItemRepo) RequireOwnedByBranch(ctx context.Context, id, storeID, branchID string) (*model.Item, error) {
	item, err := r.GetByID(ctx, id, storeID)
	if err != nil {
		return nil, err
	}
	if item.BranchID != branchID {
		return nil, pgx.ErrNoRows
	}
	return item, nil
}

func (r *ItemRepo) GetByBranchAndName(ctx context.Context, branchID, name string) (*model.Item, error) {
	item := &model.Item{}
	err := r.db.QueryRow(ctx, `
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order
		FROM items
		WHERE branch_id = $1 AND name = $2
	`, branchID, name).Scan(
		&item.ID, &item.CategoryID, &item.StoreID, &item.BranchID, &item.Name, &item.Description, &item.BasePrice, &item.ImageURL, &item.IsAvailable, &item.SortOrder,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}
