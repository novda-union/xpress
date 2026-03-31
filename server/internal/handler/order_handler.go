package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
	"github.com/xpressgo/server/internal/service"
	"github.com/xpressgo/server/internal/telegram"
	"github.com/xpressgo/server/internal/ws"
)

type OrderHandler struct {
	orderService         *service.OrderService
	orderNotificationSvc *service.OrderNotificationService
	branchRepo           *repository.BranchRepo
	telegramBot          *telegram.Bot
	hub                  *ws.Hub
}

func NewOrderHandler(orderService *service.OrderService, orderNotificationSvc *service.OrderNotificationService, branchRepo *repository.BranchRepo, telegramBot *telegram.Bot, hub *ws.Hub) *OrderHandler {
	return &OrderHandler{
		orderService:         orderService,
		orderNotificationSvc: orderNotificationSvc,
		branchRepo:           branchRepo,
		telegramBot:          telegramBot,
		hub:                  hub,
	}
}

type createOrderRequest struct {
	BranchID      string                   `json:"branch_id"`
	PaymentMethod string                   `json:"payment_method"`
	ETAMinutes    int                      `json:"eta_minutes"`
	Items         []createOrderItemRequest `json:"items"`
}

type createOrderItemRequest struct {
	ItemID    string                       `json:"item_id"`
	ItemName  string                       `json:"item_name"`
	ItemPrice int64                        `json:"item_price"`
	Quantity  int                          `json:"quantity"`
	Modifiers []createOrderModifierRequest `json:"modifiers"`
}

type createOrderModifierRequest struct {
	ModifierID      string `json:"modifier_id"`
	ModifierName    string `json:"modifier_name"`
	PriceAdjustment int64  `json:"price_adjustment"`
}

func (h *OrderHandler) CreateOrder(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req createOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if strings.TrimSpace(req.BranchID) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}

	var orderItems []model.OrderItem
	for _, ri := range req.Items {
		var mods []model.OrderItemModifier
		for _, rm := range ri.Modifiers {
			mods = append(mods, model.OrderItemModifier{
				ModifierID:      &rm.ModifierID,
				ModifierName:    rm.ModifierName,
				PriceAdjustment: rm.PriceAdjustment,
			})
		}
		if mods == nil {
			mods = []model.OrderItemModifier{}
		}
		itemID := ri.ItemID
		orderItems = append(orderItems, model.OrderItem{
			ItemID:    &itemID,
			ItemName:  ri.ItemName,
			ItemPrice: ri.ItemPrice,
			Quantity:  ri.Quantity,
			Modifiers: mods,
		})
	}

	paymentMethod := req.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "pay_at_pickup"
	}

	order := &model.Order{
		UserID:        userID,
		BranchID:      req.BranchID,
		PaymentMethod: paymentMethod,
		ETAMinutes:    req.ETAMinutes,
		Items:         orderItems,
	}

	if err := h.orderService.CreateOrder(c.Request().Context(), order); err != nil {
		status := http.StatusInternalServerError
		if isOrderValidationError(err) {
			status = http.StatusBadRequest
		}
		return c.JSON(status, map[string]string{"error": err.Error()})
	}

	h.notifyOrderCreated(c, order)

	return c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) notifyOrderCreated(c echo.Context, order *model.Order) {
	if order == nil {
		return
	}

	h.hub.NotifyStoreAndBranch(order.StoreID, order.BranchID, ws.Message{
		Type:  "order:new",
		Order: order,
	})

	if h.telegramBot == nil || h.branchRepo == nil {
		return
	}

	detail, err := h.branchRepo.GetByID(c.Request().Context(), order.BranchID)
	if err != nil {
		return
	}

	switch {
	case detail.Branch.TelegramGroupChatID != nil:
		h.telegramBot.SendNewOrderToChat(*detail.Branch.TelegramGroupChatID, detail.Branch.Name, order)
	case detail.Store.TelegramGroupChatID != nil:
		h.telegramBot.SendNewOrderToChat(*detail.Store.TelegramGroupChatID, detail.Store.Name, order)
	}
}

func (h *OrderHandler) GetOrder(c echo.Context) error {
	order, err := h.orderService.GetOrder(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "order not found"})
	}
	return c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) ListOrders(c echo.Context) error {
	userID := c.Get("user_id").(string)
	orders, err := h.orderService.ListByUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list orders"})
	}
	return c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) CancelOrder(c echo.Context) error {
	userID := c.Get("user_id").(string)
	order, err := h.orderService.CancelOrder(c.Request().Context(), c.Param("id"), userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	h.hub.NotifyStoreAndBranch(order.StoreID, order.BranchID, ws.Message{
		Type:    "order:cancelled",
		OrderID: order.ID,
	})

	if h.telegramBot != nil && h.branchRepo != nil {
		detail, err := h.branchRepo.GetByID(c.Request().Context(), order.BranchID)
		if err == nil {
			switch {
			case detail.Branch.TelegramGroupChatID != nil:
				h.telegramBot.SendOrderCancelledToChat(*detail.Branch.TelegramGroupChatID, detail.Branch.Name, order.OrderNumber)
			case detail.Store.TelegramGroupChatID != nil:
				h.telegramBot.SendOrderCancelledToChat(*detail.Store.TelegramGroupChatID, detail.Store.Name, order.OrderNumber)
			}
		}
	}

	return c.JSON(http.StatusOK, order)
}

// Admin endpoints

func (h *OrderHandler) AdminListOrders(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	branchID, _ := c.Get("branch_id").(*string)
	status := c.QueryParam("status")
	orders, err := h.orderService.ListByScope(c.Request().Context(), storeID, branchID, status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list orders"})
	}
	return c.JSON(http.StatusOK, orders)
}

type updateStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (h *OrderHandler) AdminUpdateStatus(c echo.Context) error {
	var req updateStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	order, err := h.orderService.UpdateStatus(c.Request().Context(), c.Param("id"), req.Status, req.Reason)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Notify user via WebSocket
	if req.Status == "rejected" {
		h.hub.NotifyUser(order.UserID, ws.Message{
			Type:    "order:rejected",
			OrderID: order.ID,
			Status:  req.Status,
			Reason:  req.Reason,
		})
	} else {
		h.hub.NotifyUser(order.UserID, ws.Message{
			Type:    "order:status",
			OrderID: order.ID,
			Status:  req.Status,
		})
	}

	if h.orderNotificationSvc != nil {
		if err := h.orderNotificationSvc.NotifyBranchOrderStatus(c.Request().Context(), order); err != nil {
			log.Printf("notifications: failed to send branch status update for order %s: %v", order.ID, err)
		}
	}

	return c.JSON(http.StatusOK, order)
}

func isOrderValidationError(err error) bool {
	if err == nil {
		return false
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "branch_id is required"):
		return true
	case strings.Contains(msg, "branch not found"):
		return true
	case strings.Contains(msg, "at least one item is required"):
		return true
	case strings.Contains(msg, "invalid item"):
		return true
	case strings.Contains(msg, "invalid modifier"):
		return true
	case strings.Contains(msg, "does not belong to this branch"):
		return true
	case strings.Contains(msg, "does not belong to item"):
		return true
	case strings.Contains(msg, "quantity must be greater than zero"):
		return true
	default:
		return false
	}
}
