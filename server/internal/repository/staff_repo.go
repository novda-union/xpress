package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type StaffRepo struct {
	db *pgxpool.Pool
}

func NewStaffRepo(db *pgxpool.Pool) *StaffRepo {
	return &StaffRepo{db: db}
}

func (r *StaffRepo) GetByStoreAndCode(ctx context.Context, storeID, staffCode string) (*model.Staff, error) {
	s := &model.Staff{}
	err := r.db.QueryRow(ctx, `
		SELECT id, store_id, staff_code, name, password_hash, role, is_active, created_at
		FROM store_staff WHERE store_id = $1 AND staff_code = $2 AND is_active = true
	`, storeID, staffCode).Scan(
		&s.ID, &s.StoreID, &s.StaffCode, &s.Name, &s.PasswordHash, &s.Role, &s.IsActive, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}
