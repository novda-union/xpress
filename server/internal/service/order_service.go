package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

type OrderService struct {
	orderRepo       *repository.OrderRepo
	branchRepo      *repository.BranchRepo
	menuRepo        *repository.MenuRepo
	transactionRepo *repository.TransactionRepo
}

func NewOrderService(orderRepo *repository.OrderRepo, branchRepo *repository.BranchRepo, menuRepo *repository.MenuRepo, transactionRepo *repository.TransactionRepo) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		branchRepo:      branchRepo,
		menuRepo:        menuRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) error {
	if strings.TrimSpace(order.BranchID) == "" {
		return errors.New("branch_id is required")
	}

	branchDetail, err := s.branchRepo.GetByID(ctx, order.BranchID)
	if err != nil {
		return errors.New("branch not found")
	}

	validatedItems, total, err := s.validateOrderItems(ctx, branchDetail.Branch.ID, order.Items)
	if err != nil {
		return err
	}
	order.Items = validatedItems
	order.StoreID = branchDetail.Store.ID
	order.TotalPrice = total

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return err
	}

	// Create transaction for commission tracking
	commissionAmount := int64(math.Round(float64(order.TotalPrice) * branchDetail.Store.CommissionRate / 100))
	tx := &model.Transaction{
		OrderID:          order.ID,
		StoreID:          order.StoreID,
		OrderTotal:       order.TotalPrice,
		CommissionRate:   branchDetail.Store.CommissionRate,
		CommissionAmount: commissionAmount,
	}
	if err := s.transactionRepo.Create(ctx, tx); err != nil {
		return err
	}

	return nil
}

func (s *OrderService) validateOrderItems(ctx context.Context, branchID string, requested []model.OrderItem) ([]model.OrderItem, int64, error) {
	if len(requested) == 0 {
		return nil, 0, errors.New("at least one item is required")
	}

	menu, err := s.menuRepo.GetFullMenuByBranch(ctx, branchID)
	if err != nil {
		return nil, 0, err
	}

	itemsByID := make(map[string]*model.MenuItem)
	modifiersByID := make(map[string]model.Modifier)
	modifierItemByID := make(map[string]string)

	for i := range menu.Categories {
		for j := range menu.Categories[i].Items {
			item := &menu.Categories[i].Items[j]
			itemsByID[item.ID] = item
			for k := range item.ModifierGroups {
				group := &item.ModifierGroups[k]
				for m := range group.Modifiers {
					mod := group.Modifiers[m]
					modifiersByID[mod.ID] = mod
					modifierItemByID[mod.ID] = item.ID
				}
			}
		}
	}

	validatedItems := make([]model.OrderItem, 0, len(requested))
	var total int64

	for _, requestedItem := range requested {
		if requestedItem.ItemID == nil || strings.TrimSpace(*requestedItem.ItemID) == "" {
			return nil, 0, errors.New("invalid item")
		}

		menuItem, ok := itemsByID[*requestedItem.ItemID]
		if !ok {
			return nil, 0, fmt.Errorf("item %s does not belong to this branch", *requestedItem.ItemID)
		}

		if requestedItem.Quantity <= 0 {
			return nil, 0, errors.New("item quantity must be greater than zero")
		}

		validatedMods := make([]model.OrderItemModifier, 0, len(requestedItem.Modifiers))
		modTotal := int64(0)
		for _, requestedMod := range requestedItem.Modifiers {
			if requestedMod.ModifierID == nil || strings.TrimSpace(*requestedMod.ModifierID) == "" {
				return nil, 0, errors.New("invalid modifier")
			}

			mod, ok := modifiersByID[*requestedMod.ModifierID]
			if !ok {
				return nil, 0, fmt.Errorf("modifier %s does not belong to this branch", *requestedMod.ModifierID)
			}
			if modifierItemByID[mod.ID] != menuItem.ID {
				return nil, 0, fmt.Errorf("modifier %s does not belong to item %s", mod.ID, menuItem.ID)
			}

			validatedMods = append(validatedMods, model.OrderItemModifier{
				ModifierID:      &mod.ID,
				ModifierName:    mod.Name,
				PriceAdjustment: mod.PriceAdjustment,
			})
			modTotal += mod.PriceAdjustment
		}

		validatedItem := model.OrderItem{
			ItemID:    &menuItem.ID,
			ItemName:  menuItem.Name,
			ItemPrice: menuItem.BasePrice,
			Quantity:  requestedItem.Quantity,
			Modifiers: validatedMods,
		}
		total += (menuItem.BasePrice + modTotal) * int64(requestedItem.Quantity)
		validatedItems = append(validatedItems, validatedItem)
	}

	return validatedItems, total, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListByStore(ctx context.Context, storeID, status string) ([]model.Order, error) {
	return s.orderRepo.ListByScope(ctx, storeID, nil, status)
}

func (s *OrderService) ListByScope(ctx context.Context, storeID string, branchID *string, status string) ([]model.Order, error) {
	return s.orderRepo.ListByScope(ctx, storeID, branchID, status)
}

func (s *OrderService) ListByUser(ctx context.Context, userID string) ([]model.Order, error) {
	return s.orderRepo.ListByUser(ctx, userID)
}

var validTransitions = map[string][]string{
	"pending":   {"accepted", "rejected"},
	"accepted":  {"preparing"},
	"preparing": {"ready"},
	"ready":     {"picked_up"},
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID, newStatus, reason string) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	allowed, ok := validTransitions[order.Status]
	if !ok {
		return nil, errors.New("order is in a terminal state")
	}

	valid := false
	for _, s := range allowed {
		if s == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("invalid status transition from " + order.Status + " to " + newStatus)
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, newStatus, reason); err != nil {
		return nil, err
	}

	order.Status = newStatus
	order.RejectionReason = reason
	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID, userID string) (*model.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("not your order")
	}

	if order.Status != "pending" {
		return nil, errors.New("can only cancel pending orders")
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, "cancelled", ""); err != nil {
		return nil, err
	}

	order.Status = "cancelled"
	return order, nil
}
