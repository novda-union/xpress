package model

import "time"

type Order struct {
	ID              string      `json:"id"`
	OrderNumber     int         `json:"order_number"`
	UserID          string      `json:"user_id"`
	CustomerPhone   string      `json:"customer_phone,omitempty"`
	StoreID         string      `json:"store_id"`
	BranchID        string      `json:"branch_id"`
	Status          string      `json:"status"`
	TotalPrice      int64       `json:"total_price"`
	PaymentMethod   string      `json:"payment_method"`
	PaymentStatus   string      `json:"payment_status"`
	ETAMinutes      int         `json:"eta_minutes"`
	RejectionReason string      `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID        string              `json:"id"`
	OrderID   string              `json:"order_id"`
	ItemID    *string             `json:"item_id,omitempty"`
	ItemName  string              `json:"item_name"`
	ItemPrice int64               `json:"item_price"`
	Quantity  int                 `json:"quantity"`
	Modifiers []OrderItemModifier `json:"modifiers,omitempty"`
}

type OrderItemModifier struct {
	ID              string  `json:"id"`
	OrderItemID     string  `json:"order_item_id"`
	ModifierID      *string `json:"modifier_id,omitempty"`
	ModifierName    string  `json:"modifier_name"`
	PriceAdjustment int64   `json:"price_adjustment"`
}

type Transaction struct {
	ID               string    `json:"id"`
	OrderID          string    `json:"order_id"`
	StoreID          string    `json:"store_id"`
	OrderTotal       int64     `json:"order_total"`
	CommissionRate   float64   `json:"commission_rate"`
	CommissionAmount int64     `json:"commission_amount"`
	CreatedAt        time.Time `json:"created_at"`
}

// Menu is the full nested menu response
type Menu struct {
	Categories []MenuCategory `json:"categories"`
}

type MenuCategory struct {
	Category
	Items []MenuItem `json:"items"`
}

type MenuItem struct {
	Item
	ModifierGroups []ModifierGroup `json:"modifier_groups"`
}
