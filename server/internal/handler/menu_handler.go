package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
)

type MenuHandler struct {
	categoryRepo      *repository.CategoryRepo
	itemRepo          *repository.ItemRepo
	modifierGroupRepo *repository.ModifierGroupRepo
	modifierRepo      *repository.ModifierRepo
	menuRepo          *repository.MenuRepo
}

func NewMenuHandler(
	categoryRepo *repository.CategoryRepo,
	itemRepo *repository.ItemRepo,
	modifierGroupRepo *repository.ModifierGroupRepo,
	modifierRepo *repository.ModifierRepo,
	menuRepo *repository.MenuRepo,
) *MenuHandler {
	return &MenuHandler{
		categoryRepo:      categoryRepo,
		itemRepo:          itemRepo,
		modifierGroupRepo: modifierGroupRepo,
		modifierRepo:      modifierRepo,
		menuRepo:          menuRepo,
	}
}

// Categories

func (h *MenuHandler) ListCategories(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	cats, err := h.categoryRepo.ListByStore(c.Request().Context(), storeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
	}
	return c.JSON(http.StatusOK, cats)
}

func (h *MenuHandler) CreateCategory(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var cat model.Category
	if err := c.Bind(&cat); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	cat.StoreID = storeID
	cat.IsActive = true
	if err := h.categoryRepo.Create(c.Request().Context(), &cat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
	}
	return c.JSON(http.StatusCreated, cat)
}

func (h *MenuHandler) UpdateCategory(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var cat model.Category
	if err := c.Bind(&cat); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	cat.ID = c.Param("id")
	cat.StoreID = storeID
	if err := h.categoryRepo.Update(c.Request().Context(), &cat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
	}
	return c.JSON(http.StatusOK, cat)
}

func (h *MenuHandler) DeleteCategory(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	if err := h.categoryRepo.Delete(c.Request().Context(), c.Param("id"), storeID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Items

func (h *MenuHandler) ListItems(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	items, err := h.itemRepo.ListByCategory(c.Request().Context(), c.Param("id"), storeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list items"})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *MenuHandler) CreateItem(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var item model.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	item.StoreID = storeID
	item.IsAvailable = true
	if err := h.itemRepo.Create(c.Request().Context(), &item); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create item"})
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *MenuHandler) UpdateItem(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var item model.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	item.ID = c.Param("id")
	item.StoreID = storeID
	if err := h.itemRepo.Update(c.Request().Context(), &item); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update item"})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) DeleteItem(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	if err := h.itemRepo.Delete(c.Request().Context(), c.Param("id"), storeID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete item"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Modifier Groups

func (h *MenuHandler) CreateModifierGroup(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var mg model.ModifierGroup
	if err := c.Bind(&mg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mg.ItemID = c.Param("id")
	mg.StoreID = storeID
	if err := h.modifierGroupRepo.Create(c.Request().Context(), &mg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create modifier group"})
	}
	return c.JSON(http.StatusCreated, mg)
}

func (h *MenuHandler) UpdateModifierGroup(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var mg model.ModifierGroup
	if err := c.Bind(&mg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mg.ID = c.Param("id")
	mg.StoreID = storeID
	if err := h.modifierGroupRepo.Update(c.Request().Context(), &mg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update modifier group"})
	}
	return c.JSON(http.StatusOK, mg)
}

func (h *MenuHandler) DeleteModifierGroup(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	if err := h.modifierGroupRepo.Delete(c.Request().Context(), c.Param("id"), storeID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete modifier group"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Modifiers

func (h *MenuHandler) CreateModifier(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var mod model.Modifier
	if err := c.Bind(&mod); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mod.ModifierGroupID = c.Param("id")
	mod.StoreID = storeID
	mod.IsAvailable = true
	if err := h.modifierRepo.Create(c.Request().Context(), &mod); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create modifier"})
	}
	return c.JSON(http.StatusCreated, mod)
}

func (h *MenuHandler) UpdateModifier(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	var mod model.Modifier
	if err := c.Bind(&mod); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mod.ID = c.Param("id")
	mod.StoreID = storeID
	if err := h.modifierRepo.Update(c.Request().Context(), &mod); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update modifier"})
	}
	return c.JSON(http.StatusOK, mod)
}

func (h *MenuHandler) DeleteModifier(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	if err := h.modifierRepo.Delete(c.Request().Context(), c.Param("id"), storeID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete modifier"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Admin full menu (with all items, including unavailable)
func (h *MenuHandler) AdminGetFullMenu(c echo.Context) error {
	storeID := c.Get("store_id").(string)
	menu, err := h.menuRepo.GetFullMenu(c.Request().Context(), storeID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load menu"})
	}
	return c.JSON(http.StatusOK, menu)
}
