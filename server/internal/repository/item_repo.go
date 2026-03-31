package repository

import (
	"context"
	"fmt"
	"time"

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
		SELECT id, category_id, store_id, branch_id, name, description, base_price, image_url, is_available, sort_order, created_at
		FROM items
		WHERE branch_id = $1 AND name = $2
	`, branchID, name).Scan(
		&item.ID, &item.CategoryID, &item.StoreID, &item.BranchID, &item.Name, &item.Description, &item.BasePrice, &item.ImageURL, &item.IsAvailable, &item.SortOrder, &item.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}

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
