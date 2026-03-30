package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) ListByStore(ctx context.Context, storeID string, branchID *string) ([]model.Category, error) {
	query := `
		SELECT id, store_id, branch_id, name, sort_order, is_active
		FROM categories
		WHERE store_id = $1 AND is_active = true
	`
	args := []any{storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $2`
		args = append(args, *branchID)
	}
	query += ` ORDER BY sort_order`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.StoreID, &c.BranchID, &c.Name, &c.SortOrder, &c.IsActive); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id, storeID string) (*model.Category, error) {
	c := &model.Category{}
	err := r.db.QueryRow(ctx, `
		SELECT id, store_id, branch_id, name, sort_order, is_active
		FROM categories
		WHERE id = $1 AND store_id = $2
	`, id, storeID).Scan(&c.ID, &c.StoreID, &c.BranchID, &c.Name, &c.SortOrder, &c.IsActive)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, c *model.Category) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO categories (store_id, branch_id, name, sort_order)
		VALUES ($1, $2, $3, $4) RETURNING id
	`, c.StoreID, c.BranchID, c.Name, c.SortOrder).Scan(&c.ID)
}

func (r *CategoryRepo) Update(ctx context.Context, c *model.Category) error {
	_, err := r.db.Exec(ctx, `
		UPDATE categories SET name=$2, sort_order=$3, is_active=$4, branch_id=$5 WHERE id=$1 AND store_id=$6
	`, c.ID, c.Name, c.SortOrder, c.IsActive, c.BranchID, c.StoreID)
	return err
}

func (r *CategoryRepo) Delete(ctx context.Context, id, storeID string, branchID *string) error {
	query := `DELETE FROM categories WHERE id=$1 AND store_id=$2`
	args := []any{id, storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND branch_id = $3`
		args = append(args, *branchID)
	}
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
