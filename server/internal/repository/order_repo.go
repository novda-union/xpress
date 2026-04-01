package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xpressgo/server/internal/model"
)

type OrderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepo(db *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, store_id, branch_id, status, total_price, payment_method, payment_status, eta_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, order_number, created_at, updated_at
	`, o.UserID, o.StoreID, o.BranchID, "pending", o.TotalPrice, o.PaymentMethod, "pending", o.ETAMinutes).Scan(
		&o.ID, &o.OrderNumber, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return err
	}
	o.Status = "pending"
	o.PaymentStatus = "pending"

	for i := range o.Items {
		item := &o.Items[i]
		err = tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, item_id, item_name, item_price, quantity)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, o.ID, item.ItemID, item.ItemName, item.ItemPrice, item.Quantity).Scan(&item.ID)
		if err != nil {
			return err
		}
		item.OrderID = o.ID

		for j := range item.Modifiers {
			mod := &item.Modifiers[j]
			err = tx.QueryRow(ctx, `
				INSERT INTO order_item_modifiers (order_item_id, modifier_id, modifier_name, price_adjustment)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`, item.ID, mod.ModifierID, mod.ModifierName, mod.PriceAdjustment).Scan(&mod.ID)
			if err != nil {
				return err
			}
			mod.OrderItemID = item.ID
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRow(ctx, `
		SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
		       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.id = $1
	`, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.CustomerPhone, &o.StoreID, &o.BranchID, &o.Status, &o.TotalPrice,
		&o.PaymentMethod, &o.PaymentStatus, &o.ETAMinutes, &o.RejectionReason, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load items
	itemRows, err := r.db.Query(ctx, `
		SELECT id, order_id, item_id, item_name, item_price, quantity
		FROM order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var item model.OrderItem
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ItemID, &item.ItemName, &item.ItemPrice, &item.Quantity); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}

	// Load modifiers for each item
	for i := range o.Items {
		modRows, err := r.db.Query(ctx, `
			SELECT id, order_item_id, modifier_id, modifier_name, price_adjustment
			FROM order_item_modifiers WHERE order_item_id = $1
		`, o.Items[i].ID)
		if err != nil {
			return nil, err
		}
		for modRows.Next() {
			var mod model.OrderItemModifier
			if err := modRows.Scan(&mod.ID, &mod.OrderItemID, &mod.ModifierID, &mod.ModifierName, &mod.PriceAdjustment); err != nil {
				modRows.Close()
				return nil, err
			}
			o.Items[i].Modifiers = append(o.Items[i].Modifiers, mod)
		}
		modRows.Close()
		if o.Items[i].Modifiers == nil {
			o.Items[i].Modifiers = []model.OrderItemModifier{}
		}
	}

	if o.Items == nil {
		o.Items = []model.OrderItem{}
	}

	return o, nil
}

func (r *OrderRepo) ListByStore(ctx context.Context, storeID string, status string) ([]model.Order, error) {
	return r.ListByScope(ctx, storeID, nil, status)
}

func (r *OrderRepo) ListByScope(ctx context.Context, storeID string, branchID *string, status string) ([]model.Order, error) {
	query := `
		SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
		       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.store_id = $1
	`
	args := []any{storeID}
	if branchID != nil && *branchID != "" {
		args = append(args, *branchID)
		query += fmt.Sprintf(" AND branch_id = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}
	query += ` ORDER BY o.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.CustomerPhone, &o.StoreID, &o.BranchID, &o.Status, &o.TotalPrice,
			&o.PaymentMethod, &o.PaymentStatus, &o.ETAMinutes, &o.RejectionReason, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		o.Items = []model.OrderItem{}
		orders = append(orders, o)
	}
	rows.Close()

	if len(orders) == 0 {
		return orders, nil
	}

	// Batch-load items for all orders in a single query
	orderIDs := make([]string, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.ID
	}

	itemRows, err := r.db.Query(ctx, `
		SELECT id, order_id, item_id, item_name, item_price, quantity
		FROM order_items WHERE order_id = ANY($1)
	`, orderIDs)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	var allItemIDs []string
	itemsByOrder := make(map[string][]model.OrderItem)
	for itemRows.Next() {
		var item model.OrderItem
		if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ItemID, &item.ItemName, &item.ItemPrice, &item.Quantity); err != nil {
			return nil, err
		}
		item.Modifiers = []model.OrderItemModifier{}
		itemsByOrder[item.OrderID] = append(itemsByOrder[item.OrderID], item)
		allItemIDs = append(allItemIDs, item.ID)
	}
	itemRows.Close()

	// Batch-load modifiers for all items
	if len(allItemIDs) > 0 {
		modRows, err := r.db.Query(ctx, `
			SELECT id, order_item_id, modifier_id, modifier_name, price_adjustment
			FROM order_item_modifiers WHERE order_item_id = ANY($1)
		`, allItemIDs)
		if err != nil {
			return nil, err
		}
		defer modRows.Close()

		modsByItem := make(map[string][]model.OrderItemModifier)
		for modRows.Next() {
			var mod model.OrderItemModifier
			if err := modRows.Scan(&mod.ID, &mod.OrderItemID, &mod.ModifierID, &mod.ModifierName, &mod.PriceAdjustment); err != nil {
				return nil, err
			}
			modsByItem[mod.OrderItemID] = append(modsByItem[mod.OrderItemID], mod)
		}

		for orderID, items := range itemsByOrder {
			for j := range items {
				if mods := modsByItem[items[j].ID]; len(mods) > 0 {
					items[j].Modifiers = mods
				}
			}
			itemsByOrder[orderID] = items
		}
	}

	// Assign items to their respective orders
	for i := range orders {
		if items := itemsByOrder[orders[i].ID]; len(items) > 0 {
			orders[i].Items = items
		}
	}

	return orders, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID string) ([]model.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT o.id, o.order_number, o.user_id, u.phone, o.store_id, o.branch_id, o.status, o.total_price,
		       o.payment_method, o.payment_status, o.eta_minutes, o.rejection_reason, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		WHERE o.user_id = $1 ORDER BY o.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.CustomerPhone, &o.StoreID, &o.BranchID, &o.Status, &o.TotalPrice,
			&o.PaymentMethod, &o.PaymentStatus, &o.ETAMinutes, &o.RejectionReason, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func (r *OrderRepo) GetBranchDailySummary(ctx context.Context, branchID string, summaryDate time.Time, startUTC, endUTC time.Time) (*model.BranchDailyOrderSummary, error) {
	summary := &model.BranchDailyOrderSummary{
		BranchID:  branchID,
		LocalDate: summaryDate,
	}

	err := r.db.QueryRow(ctx, `
		SELECT
			b.name,
			COUNT(o.id)::bigint AS total_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'pending')::bigint AS pending_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'accepted')::bigint AS accepted_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'preparing')::bigint AS preparing_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'ready')::bigint AS ready_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'picked_up')::bigint AS picked_up_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'rejected')::bigint AS rejected_orders,
			COUNT(o.id) FILTER (WHERE o.status = 'cancelled')::bigint AS cancelled_orders,
			COALESCE(SUM(o.total_price), 0)::bigint AS total_created_order_sum,
			COALESCE(SUM(o.total_price) FILTER (WHERE o.status = 'picked_up'), 0)::bigint AS total_picked_up_order_sum
		FROM branches b
		LEFT JOIN orders o
			ON o.branch_id = b.id
			AND o.created_at >= $2
			AND o.created_at < $3
		WHERE b.id = $1
		GROUP BY b.name
	`, branchID, startUTC, endUTC).Scan(
		&summary.BranchName,
		&summary.TotalOrders,
		&summary.PendingOrders,
		&summary.AcceptedOrders,
		&summary.PreparingOrders,
		&summary.ReadyOrders,
		&summary.PickedUpOrders,
		&summary.RejectedOrders,
		&summary.CancelledOrders,
		&summary.TotalCreatedOrderSum,
		&summary.TotalPickedUpOrderSum,
	)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id, status, rejectionReason string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE orders SET status=$2, rejection_reason=$3, updated_at=NOW() WHERE id=$1
	`, id, status, rejectionReason)
	return err
}
