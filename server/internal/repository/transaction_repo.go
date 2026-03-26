package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, t *model.Transaction) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO transactions (order_id, store_id, order_total, commission_rate, commission_amount)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at
	`, t.OrderID, t.StoreID, t.OrderTotal, t.CommissionRate, t.CommissionAmount).Scan(&t.ID, &t.CreatedAt)
}
