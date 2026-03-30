package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type StoreRepo struct {
	db *pgxpool.Pool
}

func NewStoreRepo(db *pgxpool.Pool) *StoreRepo {
	return &StoreRepo{db: db}
}

func (r *StoreRepo) GetBySlug(ctx context.Context, slug string) (*model.Store, error) {
	s := &model.Store{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, code, slug, category, description, address, phone, logo_url,
		       telegram_group_chat_id, subscription_tier, subscription_expires_at,
		       commission_rate, is_active, created_at, updated_at
		FROM stores WHERE slug = $1 AND is_active = true
	`, slug).Scan(
		&s.ID, &s.Name, &s.Code, &s.Slug, &s.Category, &s.Description, &s.Address, &s.Phone, &s.LogoURL,
		&s.TelegramGroupChatID, &s.SubscriptionTier, &s.SubscriptionExpires,
		&s.CommissionRate, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *StoreRepo) GetByCode(ctx context.Context, code string) (*model.Store, error) {
	s := &model.Store{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, code, slug, category, description, address, phone, logo_url,
		       telegram_group_chat_id, subscription_tier, subscription_expires_at,
		       commission_rate, is_active, created_at, updated_at
		FROM stores WHERE code = $1
	`, code).Scan(
		&s.ID, &s.Name, &s.Code, &s.Slug, &s.Category, &s.Description, &s.Address, &s.Phone, &s.LogoURL,
		&s.TelegramGroupChatID, &s.SubscriptionTier, &s.SubscriptionExpires,
		&s.CommissionRate, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *StoreRepo) GetByID(ctx context.Context, id string) (*model.Store, error) {
	s := &model.Store{}
	err := r.db.QueryRow(ctx, `
		SELECT id, name, code, slug, category, description, address, phone, logo_url,
		       telegram_group_chat_id, subscription_tier, subscription_expires_at,
		       commission_rate, is_active, created_at, updated_at
		FROM stores WHERE id = $1
	`, id).Scan(
		&s.ID, &s.Name, &s.Code, &s.Slug, &s.Category, &s.Description, &s.Address, &s.Phone, &s.LogoURL,
		&s.TelegramGroupChatID, &s.SubscriptionTier, &s.SubscriptionExpires,
		&s.CommissionRate, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *StoreRepo) Update(ctx context.Context, s *model.Store) error {
	_, err := r.db.Exec(ctx, `
		UPDATE stores SET name=$2, description=$3, address=$4, phone=$5, logo_url=$6,
		       telegram_group_chat_id=$7, updated_at=NOW()
		WHERE id = $1
	`, s.ID, s.Name, s.Description, s.Address, s.Phone, s.LogoURL, s.TelegramGroupChatID)
	return err
}
