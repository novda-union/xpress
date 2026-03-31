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
	verifyRepo := repository.NewPhoneVerificationRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	itemRepo := repository.NewItemRepo(db)
	modGroupRepo := repository.NewModifierGroupRepo(db)
	modRepo := repository.NewModifierRepo(db)
	branchRepo := repository.NewBranchRepo(db)
	menuRepo := repository.NewMenuRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	notificationDeliveryRepo := repository.NewNotificationDeliveryRepo(db)
	txRepo := repository.NewTransactionRepo(db)

	// Telegram bot (runs in background)
	log.Printf("Telegram bot token length: %d", len(cfg.TelegramBotToken))
	bot, err := telegram.NewBot(cfg.TelegramBotToken, cfg.AppURL, userRepo, verifyRepo)
	if err != nil {
		log.Printf("Warning: telegram bot failed to start: %v", err)
	} else {
		go bot.Start()
	}

	// Services
	authService := service.NewAuthService(storeRepo, staffRepo, userRepo, cfg.JWTSecret, cfg.TelegramBotToken)
	orderService := service.NewOrderService(orderRepo, branchRepo, menuRepo, txRepo)
	orderNotificationService := service.NewOrderNotificationService(branchRepo, orderRepo, notificationDeliveryRepo, bot)
	permissionService := service.NewPermissionService()
	orderNotificationService.StartDailySummaryScheduler(context.Background())

	// WebSocket hub
	hub := ws.NewHub()

	// Echo server
	e := echo.New()
	e.HideBanner = true

	e.Use(echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogURI:      true,
		LogMethod:   true,
		LogStatus:   true,
		LogRemoteIP: true,
		LogLatency:  true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, values echomw.RequestLoggerValues) error {
			if values.Error != nil {
				e.Logger.Errorf(
					"%s %s status=%d ip=%s latency=%s error=%v",
					values.Method,
					values.URI,
					values.Status,
					values.RemoteIP,
					values.Latency,
					values.Error,
				)
				return nil
			}

			e.Logger.Infof(
				"%s %s status=%d ip=%s latency=%s",
				values.Method,
				values.URI,
				values.Status,
				values.RemoteIP,
				values.Latency,
			)
			return nil
		},
	}))
	e.Use(echomw.Recover())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins: []string{
			"https://customer.novdaunion.uz",
			"https://admin.novdaunion.uz",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Handlers
	handlers := &handler.Handlers{
		Auth:   handler.NewAuthHandler(authService),
		Branch: handler.NewBranchHandler(branchRepo, menuRepo, permissionService),
		Staff:  handler.NewStaffHandler(staffRepo, branchRepo, permissionService),
		Store:  handler.NewStoreHandler(storeRepo, branchRepo, menuRepo),
		Menu:   handler.NewMenuHandler(categoryRepo, itemRepo, modGroupRepo, modRepo, menuRepo, permissionService),
		Order:  handler.NewOrderHandler(orderService, orderNotificationService, branchRepo, bot, hub),
	}

	handler.SetupRoutes(e, handlers, hub, cfg.JWTSecret)

	log.Printf("Starting server on :%s", cfg.Port)
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
