package model

import "time"

type Item struct {
	ID          string    `json:"id"`
	CategoryID  string    `json:"category_id"`
	StoreID     string    `json:"store_id"`
	BranchID    string    `json:"branch_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BasePrice   int64     `json:"base_price"`
	ImageURL    string    `json:"image_url"`
	IsAvailable bool      `json:"is_available"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}
