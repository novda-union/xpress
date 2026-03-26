package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Claims struct {
	UserID     string `json:"user_id,omitempty"`
	TelegramID int64  `json:"telegram_id,omitempty"`
	StoreID    string `json:"store_id,omitempty"`
	StaffID    string `json:"staff_id,omitempty"`
	Role       string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func UserAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := extractClaims(c, jwtSecret)
			if err != nil || claims.UserID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			c.Set("user_id", claims.UserID)
			c.Set("telegram_id", claims.TelegramID)
			return next(c)
		}
	}
}

func AdminAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, err := extractClaims(c, jwtSecret)
			if err != nil || claims.StoreID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			c.Set("store_id", claims.StoreID)
			c.Set("staff_id", claims.StaffID)
			c.Set("role", claims.Role)
			return next(c)
		}
	}
}

func extractClaims(c echo.Context, jwtSecret string) (*Claims, error) {
	auth := c.Request().Header.Get("Authorization")
	if auth == "" {
		// Also check query param for WebSocket connections
		auth = "Bearer " + c.QueryParam("token")
	}

	tokenStr := strings.TrimPrefix(auth, "Bearer ")
	if tokenStr == "" || tokenStr == "Bearer " {
		return nil, echo.ErrUnauthorized
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, echo.ErrUnauthorized
	}

	return claims, nil
}
