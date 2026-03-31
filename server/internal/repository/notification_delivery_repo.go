package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type NotificationDeliveryRepo struct {
	db *pgxpool.Pool
}

func NewNotificationDeliveryRepo(db *pgxpool.Pool) *NotificationDeliveryRepo {
	return &NotificationDeliveryRepo{db: db}
}

func (r *NotificationDeliveryRepo) Exists(ctx context.Context, notificationType, branchID string, localDate time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM notification_deliveries
			WHERE notification_type = $1 AND branch_id = $2 AND local_date = $3::date
		)
	`, notificationType, branchID, localDate).Scan(&exists)
	return exists, err
}

func (r *NotificationDeliveryRepo) Create(ctx context.Context, delivery *model.NotificationDelivery) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO notification_deliveries (notification_type, branch_id, local_date)
		VALUES ($1, $2, $3::date)
		RETURNING id, sent_at, created_at
	`, delivery.NotificationType, delivery.BranchID, delivery.LocalDate).Scan(
		&delivery.ID,
		&delivery.SentAt,
		&delivery.CreatedAt,
	)
}
