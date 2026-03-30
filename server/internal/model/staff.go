package model

import "time"

type Staff struct {
	ID           string    `json:"id"`
	StoreID      string    `json:"store_id"`
	BranchID     *string   `json:"branch_id,omitempty"`
	StaffCode    string    `json:"staff_code"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
