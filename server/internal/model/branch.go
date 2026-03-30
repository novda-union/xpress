package model

import "time"

type Branch struct {
	ID                  string    `json:"id"`
	StoreID             string    `json:"store_id"`
	Name                string    `json:"name"`
	Address             string    `json:"address"`
	Lat                 *float64  `json:"lat,omitempty"`
	Lng                 *float64  `json:"lng,omitempty"`
	BannerImageURL      string    `json:"banner_image_url"`
	TelegramGroupChatID *int64    `json:"telegram_group_chat_id,omitempty"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
