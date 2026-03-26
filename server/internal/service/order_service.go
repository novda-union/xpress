package service

import (
	"context"
	"errors"
	"math"

	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

type OrderService struct {
	orderRepo       *repository.OrderRepo
	storeRepo       *repository.StoreRepo
	transactionRepo *repository.TransactionRepo
}

func NewOrderService(orderRepo *repository.OrderRepo, storeRepo *repository.StoreRepo, transactionRepo *repository.TransactionRepo) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		storeRepo:       storeRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) error {
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return err
	}

	// Create transaction for commission tracking
	store, err := s.storeRepo.GetByID(ctx, order.StoreID)
	if err != nil {
		return nil // Order created, commission tracking can fail gracefully
	}

	commissionAmount := int64(math.Round(float64(order.TotalPrice) * store.CommissionRate / 100))
	tx := &model.Transaction{
		OrderID:          order.ID,
		StoreID:          order.StoreID,
		OrderTotal:       order.TotalPrice,
		CommissionRate:   store.CommissionRate,
		CommissionAmount: commissionAmount,
	}
	s.transactionRepo.Create(ctx, tx)

	return nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*model.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListByStore(ctx context.Context, storeID, status string) ([]model.Order, error) {
	return s.orderRepo.ListByStore(ctx, storeID, status)
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
