package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PhoneVerificationRepo struct {
	db *pgxpool.Pool
}

func NewPhoneVerificationRepo(db *pgxpool.Pool) *PhoneVerificationRepo {
	return &PhoneVerificationRepo{db: db}
}

// Save replaces any existing pending code for this telegram user and stores the new one.
func (r *PhoneVerificationRepo) Save(ctx context.Context, telegramID int64, phone, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `DELETE FROM phone_verifications WHERE telegram_id = $1`, telegramID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO phone_verifications (telegram_id, phone, code, expires_at)
		VALUES ($1, $2, $3, $4)
	`, telegramID, phone, code, expiresAt)
	return err
}

// Consume validates the code and returns the associated phone if valid.
// The record is deleted on success so codes are single-use.
func (r *PhoneVerificationRepo) Consume(ctx context.Context, telegramID int64, code string) (string, error) {
	var phone string
	err := r.db.QueryRow(ctx, `
		DELETE FROM phone_verifications
		WHERE telegram_id = $1 AND code = $2 AND expires_at > NOW()
		RETURNING phone
	`, telegramID, code).Scan(&phone)
	if err != nil {
		return "", errors.New("invalid or expired code")
	}
	return phone, nil
}
