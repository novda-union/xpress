package model

import "time"

type NotificationDelivery struct {
	ID               string    `json:"id"`
	NotificationType string    `json:"notification_type"`
	BranchID         string    `json:"branch_id"`
	LocalDate        time.Time `json:"local_date"`
	SentAt           time.Time `json:"sent_at"`
	CreatedAt        time.Time `json:"created_at"`
}

type BranchDailyOrderSummary struct {
	BranchID              string    `json:"branch_id"`
	BranchName            string    `json:"branch_name"`
	LocalDate             time.Time `json:"local_date"`
	TotalOrders           int64     `json:"total_orders"`
	PendingOrders         int64     `json:"pending_orders"`
	AcceptedOrders        int64     `json:"accepted_orders"`
	PreparingOrders       int64     `json:"preparing_orders"`
	ReadyOrders           int64     `json:"ready_orders"`
	PickedUpOrders        int64     `json:"picked_up_orders"`
	RejectedOrders        int64     `json:"rejected_orders"`
	CancelledOrders       int64     `json:"cancelled_orders"`
	TotalCreatedOrderSum  int64     `json:"total_created_order_sum"`
	TotalPickedUpOrderSum int64     `json:"total_picked_up_order_sum"`
}
