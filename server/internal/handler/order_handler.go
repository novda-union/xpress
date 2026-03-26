package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/service"
	"github.com/xpressgo/server/internal/ws"
)

type OrderHandler struct {
	orderService *service.OrderService
	hub          *ws.Hub
}

func NewOrderHandler(orderService *service.OrderService, hub *ws.Hub) *OrderHandler {
	return &OrderHandler{orderService: orderService, hub: hub}
}

type createOrderRequest struct {
	StoreID       string                   `json:"store_id"`
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

	// Calculate total
	var total int64
	var orderItems []model.OrderItem
	for _, ri := range req.Items {
		itemTotal := ri.ItemPrice
		var mods []model.OrderItemModifier
		for _, rm := range ri.Modifiers {
			itemTotal += rm.PriceAdjustment
			mods = append(mods, model.OrderItemModifier{
				ModifierID:      &rm.ModifierID,
				ModifierName:    rm.ModifierName,
				PriceAdjustment: rm.PriceAdjustment,
			})
		}
		if mods == nil {
			mods = []model.OrderItemModifier{}
		}
		total += itemTotal * int64(ri.Quantity)
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
		StoreID:       req.StoreID,
		TotalPrice:    total,
		PaymentMethod: paymentMethod,
		ETAMinutes:    req.ETAMinutes,
		Items:         orderItems,
	}

	if err := h.orderService.CreateOrder(c.Request().Context(), order); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create order"})
	}

	// Notify store via WebSocket
	h.hub.NotifyStore(order.StoreID, ws.Message{
		Type:  "order:new",
		Order: order,
	})

	return c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrder(c echo.Context) error {
	order, err := h.orderService.GetOrder(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "order not found"})
	}
	return c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c echo.Context) error {
	userID := c.Get("user_id").(string)
	order, err := h.orderService.CancelOrder(c.Request().Context(), c.Param("id"), userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Notify store
	h.hub.NotifyStore(order.StoreID, ws.Message{
		Type:    "order:cancelled",
		OrderID: order.ID,
	})

	return c.JSON(http.StatusOK, order)
}

// Admin endpoints

func (h *OrderHandler) AdminListOrders(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	status := c.QueryParam("status")
	orders, err := h.orderService.ListByStore(c.Request().Context(), storeID, status)
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

	return c.JSON(http.StatusOK, order)
}
