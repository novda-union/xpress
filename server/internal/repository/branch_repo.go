package repository

import (
	"context"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type BranchRepo struct {
	db *pgxpool.Pool
}

type BranchPreviewItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	BasePrice int64  `json:"base_price"`
}

type DiscoverBranch struct {
	StoreID        string              `json:"store_id"`
	StoreName      string              `json:"store_name"`
	StoreSlug      string              `json:"store_slug"`
	StoreLogoURL   string              `json:"store_logo_url"`
	StoreCategory  string              `json:"store_category"`
	BranchID       string              `json:"branch_id"`
	BranchName     string              `json:"branch_name"`
	BranchAddress  string              `json:"branch_address"`
	Lat            *float64            `json:"lat,omitempty"`
	Lng            *float64            `json:"lng,omitempty"`
	BannerImageURL string              `json:"banner_image_url"`
	PreviewItems   []BranchPreviewItem `json:"preview_items"`
}

type BranchDetail struct {
	Store  model.Store  `json:"store"`
	Branch model.Branch `json:"branch"`
}

type AdminBranchSummary struct {
	model.Branch
	StaffCount int `json:"staff_count"`
}

type StoreBranchResolution struct {
	Store         model.Store    `json:"store"`
	Branches      []model.Branch `json:"branches"`
	PrimaryBranch *model.Branch  `json:"primary_branch,omitempty"`
	Mode          string         `json:"mode"`
}

func NewBranchRepo(db *pgxpool.Pool) *BranchRepo {
	return &BranchRepo{db: db}
}

func (r *BranchRepo) ListDiscover(ctx context.Context, category string) ([]DiscoverBranch, error) {
	query := `
		SELECT
			b.id, b.store_id, b.name, b.address, b.lat, b.lng, b.banner_image_url,
			s.name, s.slug, s.logo_url, s.category
		FROM branches b
		JOIN stores s ON s.id = b.store_id
		WHERE b.is_active = true AND s.is_active = true
	`
	args := []any{}
	if category != "" {
		query += ` AND s.category = $1`
		args = append(args, category)
	}
	query += ` ORDER BY s.name, b.name`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	branches := make([]DiscoverBranch, 0)
	branchIDs := make([]string, 0)
	branchIndex := make(map[string]int)

	for rows.Next() {
		var entry DiscoverBranch
		if err := rows.Scan(
			&entry.BranchID, &entry.StoreID, &entry.BranchName, &entry.BranchAddress, &entry.Lat, &entry.Lng, &entry.BannerImageURL,
			&entry.StoreName, &entry.StoreSlug, &entry.StoreLogoURL, &entry.StoreCategory,
		); err != nil {
			return nil, err
		}
		branches = append(branches, entry)
		branchIDs = append(branchIDs, entry.BranchID)
		branchIndex[entry.BranchID] = len(branches) - 1
	}

	if len(branches) == 0 {
		return []DiscoverBranch{}, nil
	}

	itemQuery := `
		WITH ranked_items AS (
			SELECT
				i.branch_id, i.id, i.name, i.image_url, i.base_price,
				ROW_NUMBER() OVER (PARTITION BY i.branch_id ORDER BY i.sort_order, i.id) AS rn
			FROM items i
			JOIN branches b ON b.id = i.branch_id
			WHERE i.is_available = true AND b.is_active = true
	`
	itemArgs := []any{}
	if len(branchIDs) > 0 {
		itemQuery += ` AND i.branch_id = ANY($1)`
		itemArgs = append(itemArgs, branchIDs)
	}
	itemQuery += `
		)
		SELECT branch_id, id, name, image_url, base_price
		FROM ranked_items
		WHERE rn <= 5
		ORDER BY branch_id, rn
	`

	itemRows, err := r.db.Query(ctx, itemQuery, itemArgs...)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var branchID string
		var item BranchPreviewItem
		if err := itemRows.Scan(&branchID, &item.ID, &item.Name, &item.ImageURL, &item.BasePrice); err != nil {
			return nil, err
		}
		idx, ok := branchIndex[branchID]
		if !ok {
			continue
		}
		branches[idx].PreviewItems = append(branches[idx].PreviewItems, item)
	}

	for i := range branches {
		if branches[i].PreviewItems == nil {
			branches[i].PreviewItems = []BranchPreviewItem{}
		}
	}

	return branches, nil
}

func (r *BranchRepo) GetByID(ctx context.Context, id string) (*BranchDetail, error) {
	detail := &BranchDetail{}
	err := r.db.QueryRow(ctx, `
		SELECT
			b.id, b.store_id, b.name, b.address, b.lat, b.lng, b.banner_image_url, b.telegram_group_chat_id, b.is_active, b.created_at, b.updated_at,
			s.id, s.name, s.code, s.slug, s.category, s.description, s.address, s.phone, s.logo_url, s.telegram_group_chat_id,
			s.subscription_tier, s.subscription_expires_at, s.commission_rate, s.is_active, s.created_at, s.updated_at
		FROM branches b
		JOIN stores s ON s.id = b.store_id
		WHERE b.id = $1 AND b.is_active = true AND s.is_active = true
	`, id).Scan(
		&detail.Branch.ID, &detail.Branch.StoreID, &detail.Branch.Name, &detail.Branch.Address, &detail.Branch.Lat, &detail.Branch.Lng,
		&detail.Branch.BannerImageURL, &detail.Branch.TelegramGroupChatID, &detail.Branch.IsActive, &detail.Branch.CreatedAt, &detail.Branch.UpdatedAt,
		&detail.Store.ID, &detail.Store.Name, &detail.Store.Code, &detail.Store.Slug, &detail.Store.Category, &detail.Store.Description,
		&detail.Store.Address, &detail.Store.Phone, &detail.Store.LogoURL, &detail.Store.TelegramGroupChatID, &detail.Store.SubscriptionTier,
		&detail.Store.SubscriptionExpires, &detail.Store.CommissionRate, &detail.Store.IsActive, &detail.Store.CreatedAt, &detail.Store.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return detail, nil
}

func (r *BranchRepo) ListByStoreSlug(ctx context.Context, slug string) (*StoreBranchResolution, error) {
	resolution := &StoreBranchResolution{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, code, slug, category, description, address, phone, logo_url,
		       telegram_group_chat_id, subscription_tier, subscription_expires_at,
		       commission_rate, is_active, created_at, updated_at
		FROM stores WHERE slug = $1 AND is_active = true
	`, slug).Scan(
		&resolution.Store.ID, &resolution.Store.Name, &resolution.Store.Code, &resolution.Store.Slug, &resolution.Store.Category,
		&resolution.Store.Description, &resolution.Store.Address, &resolution.Store.Phone, &resolution.Store.LogoURL,
		&resolution.Store.TelegramGroupChatID, &resolution.Store.SubscriptionTier, &resolution.Store.SubscriptionExpires,
		&resolution.Store.CommissionRate, &resolution.Store.IsActive, &resolution.Store.CreatedAt, &resolution.Store.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active, created_at, updated_at
		FROM branches
		WHERE store_id = $1 AND is_active = true
		ORDER BY created_at ASC, name ASC
	`, resolution.Store.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var branch model.Branch
		if err := rows.Scan(
			&branch.ID, &branch.StoreID, &branch.Name, &branch.Address, &branch.Lat, &branch.Lng,
			&branch.BannerImageURL, &branch.TelegramGroupChatID, &branch.IsActive, &branch.CreatedAt, &branch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		resolution.Branches = append(resolution.Branches, branch)
	}

	if len(resolution.Branches) == 0 {
		return nil, pgx.ErrNoRows
	}

	if len(resolution.Branches) == 1 {
		resolution.Mode = "single"
		resolution.PrimaryBranch = &resolution.Branches[0]
		return resolution, nil
	}

	resolution.Mode = "picker"
	return resolution, nil
}

func (r *BranchRepo) GetPrimaryByStoreSlug(ctx context.Context, slug string) (*model.Branch, error) {
	resolution, err := r.ListByStoreSlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if resolution.PrimaryBranch != nil {
		return resolution.PrimaryBranch, nil
	}
	if len(resolution.Branches) > 0 {
		sort.SliceStable(resolution.Branches, func(i, j int) bool {
			return resolution.Branches[i].CreatedAt.Before(resolution.Branches[j].CreatedAt)
		})
		return &resolution.Branches[0], nil
	}
	return nil, errors.New("branch not found")
}

func (r *BranchRepo) ListAdmin(ctx context.Context, storeID string, branchID *string) ([]AdminBranchSummary, error) {
	query := `
		SELECT
			b.id, b.store_id, b.name, b.address, b.lat, b.lng, b.banner_image_url,
			b.telegram_group_chat_id, b.is_active, b.created_at, b.updated_at,
			COUNT(ss.id) FILTER (WHERE ss.is_active = true) AS staff_count
		FROM branches b
		LEFT JOIN store_staff ss ON ss.branch_id = b.id
		WHERE b.store_id = $1
	`
	args := []any{storeID}
	if branchID != nil && *branchID != "" {
		query += ` AND b.id = $2`
		args = append(args, *branchID)
	}
	query += `
		GROUP BY b.id
		ORDER BY b.created_at ASC, b.name ASC
	`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []AdminBranchSummary
	for rows.Next() {
		var branch AdminBranchSummary
		if err := rows.Scan(
			&branch.ID, &branch.StoreID, &branch.Name, &branch.Address, &branch.Lat, &branch.Lng,
			&branch.BannerImageURL, &branch.TelegramGroupChatID, &branch.IsActive, &branch.CreatedAt, &branch.UpdatedAt,
			&branch.StaffCount,
		); err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	if branches == nil {
		branches = []AdminBranchSummary{}
	}
	return branches, nil
}

func (r *BranchRepo) ListActiveWithTelegramChatIDs(ctx context.Context) ([]model.Branch, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active, created_at, updated_at
		FROM branches
		WHERE is_active = true AND telegram_group_chat_id IS NOT NULL
		ORDER BY created_at ASC, name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []model.Branch
	for rows.Next() {
		var branch model.Branch
		if err := rows.Scan(
			&branch.ID, &branch.StoreID, &branch.Name, &branch.Address, &branch.Lat, &branch.Lng,
			&branch.BannerImageURL, &branch.TelegramGroupChatID, &branch.IsActive, &branch.CreatedAt, &branch.UpdatedAt,
		); err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}

	if branches == nil {
		branches = []model.Branch{}
	}

	return branches, nil
}

func (r *BranchRepo) GetByIDForStore(ctx context.Context, id, storeID string) (*model.Branch, error) {
	branch := &model.Branch{}
	err := r.db.QueryRow(ctx, `
		SELECT id, store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active, created_at, updated_at
		FROM branches
		WHERE id = $1 AND store_id = $2
	`, id, storeID).Scan(
		&branch.ID, &branch.StoreID, &branch.Name, &branch.Address, &branch.Lat, &branch.Lng,
		&branch.BannerImageURL, &branch.TelegramGroupChatID, &branch.IsActive, &branch.CreatedAt, &branch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return branch, nil
}

func (r *BranchRepo) Create(ctx context.Context, branch *model.Branch) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO branches (
			store_id, name, address, lat, lng, banner_image_url, telegram_group_chat_id, is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, branch.StoreID, branch.Name, branch.Address, branch.Lat, branch.Lng, branch.BannerImageURL, branch.TelegramGroupChatID, branch.IsActive).Scan(
		&branch.ID, &branch.CreatedAt, &branch.UpdatedAt,
	)
}

func (r *BranchRepo) Update(ctx context.Context, branch *model.Branch) error {
	return r.db.QueryRow(ctx, `
		UPDATE branches
		SET
			name = $3,
			address = $4,
			lat = $5,
			lng = $6,
			banner_image_url = $7,
			telegram_group_chat_id = $8,
			is_active = $9,
			updated_at = NOW()
		WHERE id = $1 AND store_id = $2
		RETURNING updated_at
	`, branch.ID, branch.StoreID, branch.Name, branch.Address, branch.Lat, branch.Lng, branch.BannerImageURL, branch.TelegramGroupChatID, branch.IsActive).Scan(
		&branch.UpdatedAt,
	)
}

func (r *BranchRepo) Deactivate(ctx context.Context, id, storeID string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE branches
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND store_id = $2
	`, id, storeID)
	return err
}
