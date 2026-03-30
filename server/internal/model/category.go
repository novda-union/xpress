package model

type Category struct {
	ID        string `json:"id"`
	StoreID   string `json:"store_id"`
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}
