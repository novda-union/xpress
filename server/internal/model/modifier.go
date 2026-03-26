package model

type ModifierGroup struct {
	ID            string     `json:"id"`
	ItemID        string     `json:"item_id"`
	StoreID       string     `json:"store_id"`
	Name          string     `json:"name"`
	SelectionType string     `json:"selection_type"`
	IsRequired    bool       `json:"is_required"`
	MinSelections int        `json:"min_selections"`
	MaxSelections int        `json:"max_selections"`
	SortOrder     int        `json:"sort_order"`
	Modifiers     []Modifier `json:"modifiers,omitempty"`
}

type Modifier struct {
	ID              string `json:"id"`
	ModifierGroupID string `json:"modifier_group_id"`
	StoreID         string `json:"store_id"`
	Name            string `json:"name"`
	PriceAdjustment int64  `json:"price_adjustment"`
	IsAvailable     bool   `json:"is_available"`
	SortOrder       int    `json:"sort_order"`
}
