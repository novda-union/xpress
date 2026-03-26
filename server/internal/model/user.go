package model

import "time"

type User struct {
	ID         string    `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Phone      string    `json:"phone"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
}
