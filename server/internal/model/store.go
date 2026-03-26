package model

import "time"

type Store struct {
	ID                   string     `json:"id"`
	Name                 string     `json:"name"`
	Code                 string     `json:"code"`
	Slug                 string     `json:"slug"`
	Description          string     `json:"description"`
	Address              string     `json:"address"`
	Phone                string     `json:"phone"`
	LogoURL              string     `json:"logo_url"`
	TelegramGroupChatID  *int64     `json:"telegram_group_chat_id,omitempty"`
	SubscriptionTier     string     `json:"subscription_tier"`
	SubscriptionExpires  *time.Time `json:"subscription_expires_at,omitempty"`
	CommissionRate       float64    `json:"commission_rate"`
	IsActive             bool       `json:"is_active"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
