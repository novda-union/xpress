package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByTelegramID(ctx context.Context, telegramID int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, telegram_id, phone, first_name, last_name, username, created_at
		FROM users WHERE telegram_id = $1
	`, telegramID).Scan(
		&u.ID, &u.TelegramID, &u.Phone, &u.FirstName, &u.LastName, &u.Username, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, telegram_id, phone, first_name, last_name, username, created_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.TelegramID, &u.Phone, &u.FirstName, &u.LastName, &u.Username, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, u.TelegramID, u.Phone, u.FirstName, u.LastName, u.Username).Scan(&u.ID, &u.CreatedAt)
}

func (r *UserRepo) Upsert(ctx context.Context, u *model.User) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO users (telegram_id, phone, first_name, last_name, username)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (telegram_id) DO UPDATE SET
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username
		RETURNING id, created_at
	`, u.TelegramID, u.Phone, u.FirstName, u.LastName, u.Username).Scan(&u.ID, &u.CreatedAt)
}
