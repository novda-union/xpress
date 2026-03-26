package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

type StoreHandler struct {
	storeRepo *repository.StoreRepo
	menuRepo  *repository.MenuRepo
}

func NewStoreHandler(storeRepo *repository.StoreRepo, menuRepo *repository.MenuRepo) *StoreHandler {
	return &StoreHandler{storeRepo: storeRepo, menuRepo: menuRepo}
}

func (h *StoreHandler) GetBySlug(c echo.Context) error {
	slug := c.Param("slug")
	store, err := h.storeRepo.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "store not found"})
	}
	return c.JSON(http.StatusOK, store)
}

func (h *StoreHandler) GetMenu(c echo.Context) error {
	slug := c.Param("slug")
	store, err := h.storeRepo.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "store not found"})
	}

	menu, err := h.menuRepo.GetFullMenu(c.Request().Context(), store.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load menu"})
	}

	return c.JSON(http.StatusOK, menu)
}

// Admin endpoints

func (h *StoreHandler) AdminGetStore(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	store, err := h.storeRepo.GetByID(c.Request().Context(), storeID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "store not found"})
	}
	return c.JSON(http.StatusOK, store)
}

type updateStoreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	LogoURL     string `json:"logo_url"`
}

func (h *StoreHandler) AdminUpdateStore(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var req updateStoreRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	store := &model.Store{
		ID:          storeID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		LogoURL:     req.LogoURL,
	}

	if err := h.storeRepo.Update(c.Request().Context(), store); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update store"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}
