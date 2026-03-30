package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type telegramAuthRequest struct {
	InitData    string `json:"init_data"`
	Phone       string `json:"phone"`
	PhoneNumber string `json:"phone_number"`
}

func (h *AuthHandler) TelegramAuth(c echo.Context) error {
	var req telegramAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	phone := req.Phone
	if phone == "" {
		phone = req.PhoneNumber
	}

	token, user, err := h.authService.ValidateTelegramAuth(c.Request().Context(), req.InitData, phone)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

type adminAuthRequest struct {
	StoreCode string `json:"store_code"`
	StaffCode string `json:"staff_code"`
	Password  string `json:"password"`
}

func (h *AuthHandler) AdminAuth(c echo.Context) error {
	var req adminAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	token, staff, err := h.authService.AdminLogin(c.Request().Context(), req.StoreCode, req.StaffCode, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"staff": staff,
	})
}

type devAuthRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

func (h *AuthHandler) DevAuth(c echo.Context) error {
	var req devAuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	token, user, err := h.authService.GenerateDevToken(c.Request().Context(), req.TelegramID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}
