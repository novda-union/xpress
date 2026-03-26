package main

import (
	"context"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/xpressgo/server/internal/config"
	"github.com/xpressgo/server/internal/database"
	"github.com/xpressgo/server/internal/handler"
	"github.com/xpressgo/server/internal/repository"
	"github.com/xpressgo/server/internal/service"
	"github.com/xpressgo/server/internal/telegram"
	"github.com/xpressgo/server/internal/ws"
)

func main() {
	cfg := config.Load()

	// Database
	db := database.Connect(context.Background(), cfg.DatabaseURL)
	defer db.Close()

	// Repositories
	storeRepo := repository.NewStoreRepo(db)
	staffRepo := repository.NewStaffRepo(db)
	userRepo := repository.NewUserRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	itemRepo := repository.NewItemRepo(db)
	modGroupRepo := repository.NewModifierGroupRepo(db)
	modRepo := repository.NewModifierRepo(db)
	menuRepo := repository.NewMenuRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	txRepo := repository.NewTransactionRepo(db)

	// Services
	authService := service.NewAuthService(storeRepo, staffRepo, userRepo, cfg.JWTSecret, cfg.TelegramBotToken)
	orderService := service.NewOrderService(orderRepo, storeRepo, txRepo)

	// WebSocket hub
	hub := ws.NewHub()

	// Handlers
	handlers := &handler.Handlers{
		Auth:  handler.NewAuthHandler(authService),
		Store: handler.NewStoreHandler(storeRepo, menuRepo),
		Menu:  handler.NewMenuHandler(categoryRepo, itemRepo, modGroupRepo, modRepo, menuRepo),
		Order: handler.NewOrderHandler(orderService, hub),
	}

	// Echo server
	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	handler.SetupRoutes(e, handlers, hub, cfg.JWTSecret)

	// Telegram bot (runs in background)
	log.Printf("Telegram bot token length: %d", len(cfg.TelegramBotToken))
	bot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.AppURL)
	if err != nil {
		log.Printf("Warning: telegram bot failed to start: %v", err)
	} else {
		go bot.Start()
	}
	_ = bot // Available for notification integration

	log.Printf("Starting server on :%s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
