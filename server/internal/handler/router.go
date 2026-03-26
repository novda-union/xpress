package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/middleware"
	"github.com/xpressgo/server/internal/ws"
)

type Handlers struct {
	Auth  *AuthHandler
	Store *StoreHandler
	Menu  *MenuHandler
	Order *OrderHandler
}

func SetupRoutes(e *echo.Echo, h *Handlers, hub *ws.Hub, jwtSecret string) {
	// Public routes
	e.POST("/auth/telegram", h.Auth.TelegramAuth)
	e.POST("/auth/dev", h.Auth.DevAuth) // Dev-only endpoint

	e.GET("/stores/:slug", h.Store.GetBySlug)
	e.GET("/stores/:slug/menu", h.Store.GetMenu)

	// Authenticated user routes
	user := e.Group("", middleware.UserAuth(jwtSecret))
	user.POST("/orders", h.Order.CreateOrder)
	user.GET("/orders/:id", h.Order.GetOrder)
	user.PUT("/orders/:id/cancel", h.Order.CancelOrder)
	user.GET("/ws", ws.UserWebSocket(hub))

	// Admin routes
	e.POST("/admin/auth", h.Auth.AdminAuth)

	admin := e.Group("/admin", middleware.AdminAuth(jwtSecret))
	admin.GET("/store", h.Store.AdminGetStore)
	admin.PUT("/store", h.Store.AdminUpdateStore)

	admin.GET("/menu/categories", h.Menu.ListCategories)
	admin.POST("/menu/categories", h.Menu.CreateCategory)
	admin.PUT("/menu/categories/:id", h.Menu.UpdateCategory)
	admin.DELETE("/menu/categories/:id", h.Menu.DeleteCategory)
	admin.GET("/menu/categories/:id/items", h.Menu.ListItems)

	admin.POST("/menu/items", h.Menu.CreateItem)
	admin.PUT("/menu/items/:id", h.Menu.UpdateItem)
	admin.DELETE("/menu/items/:id", h.Menu.DeleteItem)
	admin.POST("/menu/items/:id/modifier-groups", h.Menu.CreateModifierGroup)

	admin.PUT("/menu/modifier-groups/:id", h.Menu.UpdateModifierGroup)
	admin.DELETE("/menu/modifier-groups/:id", h.Menu.DeleteModifierGroup)
	admin.POST("/menu/modifier-groups/:id/modifiers", h.Menu.CreateModifier)

	admin.PUT("/menu/modifiers/:id", h.Menu.UpdateModifier)
	admin.DELETE("/menu/modifiers/:id", h.Menu.DeleteModifier)

	admin.GET("/orders", h.Order.AdminListOrders)
	admin.PUT("/orders/:id/status", h.Order.AdminUpdateStatus)
	admin.GET("/menu", h.Menu.AdminGetFullMenu)

	admin.GET("/ws", ws.AdminWebSocket(hub))
}
