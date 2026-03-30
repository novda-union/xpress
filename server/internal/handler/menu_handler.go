package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/middleware"
	"github.com/xpressgo/server/internal/model"
	"github.com/xpressgo/server/internal/repository"
	"github.com/xpressgo/server/internal/service"
)

type MenuHandler struct {
	categoryRepo      *repository.CategoryRepo
	itemRepo          *repository.ItemRepo
	modifierGroupRepo *repository.ModifierGroupRepo
	modifierRepo      *repository.ModifierRepo
	menuRepo          *repository.MenuRepo
	permissionService *service.PermissionService
}

func NewMenuHandler(
	categoryRepo *repository.CategoryRepo,
	itemRepo *repository.ItemRepo,
	modifierGroupRepo *repository.ModifierGroupRepo,
	modifierRepo *repository.ModifierRepo,
	menuRepo *repository.MenuRepo,
	permissionService *service.PermissionService,
) *MenuHandler {
	return &MenuHandler{
		categoryRepo:      categoryRepo,
		itemRepo:          itemRepo,
		modifierGroupRepo: modifierGroupRepo,
		modifierRepo:      modifierRepo,
		menuRepo:          menuRepo,
		permissionService: permissionService,
	}
}

// Categories

func (h *MenuHandler) ListCategories(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, false)
	if err != nil {
		return h.menuError(c, err)
	}

	cats, err := h.categoryRepo.ListByStore(c.Request().Context(), scope.StoreID, branchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list categories"})
	}
	return c.JSON(http.StatusOK, cats)
}

func (h *MenuHandler) CreateCategory(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var cat model.Category
	if err := c.Bind(&cat); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}

	cat.StoreID = scope.StoreID
	cat.BranchID = *branchID
	cat.IsActive = true
	if err := h.categoryRepo.Create(c.Request().Context(), &cat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create category"})
	}
	return c.JSON(http.StatusCreated, cat)
}

func (h *MenuHandler) UpdateCategory(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var cat model.Category
	if err := c.Bind(&cat); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	current, err := h.categoryRepo.GetByID(c.Request().Context(), c.Param("id"), scope.StoreID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "category not found"})
	}
	if scope.BranchID != nil && current.BranchID != *scope.BranchID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied: menu:manage"})
	}

	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		branchID = &current.BranchID
	}

	cat.ID = c.Param("id")
	cat.StoreID = scope.StoreID
	cat.BranchID = *branchID
	if err := h.categoryRepo.Update(c.Request().Context(), &cat); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update category"})
	}
	return c.JSON(http.StatusOK, cat)
}

func (h *MenuHandler) DeleteCategory(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "menu:manage"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	if err := h.categoryRepo.Delete(c.Request().Context(), c.Param("id"), scope.StoreID, scope.BranchID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete category"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Items

func (h *MenuHandler) ListItems(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, false)
	if err != nil {
		return h.menuError(c, err)
	}
	items, err := h.itemRepo.ListByCategory(c.Request().Context(), c.Param("id"), scope.StoreID, branchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list items"})
	}
	return c.JSON(http.StatusOK, items)
}

func (h *MenuHandler) CreateItem(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	var item model.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}

	item.StoreID = scope.StoreID
	item.BranchID = *branchID
	item.IsAvailable = true
	if err := h.itemRepo.Create(c.Request().Context(), &item); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create item"})
	}
	return c.JSON(http.StatusCreated, item)
}

func (h *MenuHandler) UpdateItem(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	var item model.Item
	if err := c.Bind(&item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	current, err := h.itemRepo.GetByID(c.Request().Context(), c.Param("id"), scope.StoreID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}
	if scope.BranchID != nil && current.BranchID != *scope.BranchID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied: menu:manage"})
	}

	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		branchID = &current.BranchID
	}

	item.ID = c.Param("id")
	item.StoreID = scope.StoreID
	item.BranchID = *branchID
	if err := h.itemRepo.Update(c.Request().Context(), &item); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update item"})
	}
	return c.JSON(http.StatusOK, item)
}

func (h *MenuHandler) DeleteItem(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "menu:manage"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	if err := h.itemRepo.Delete(c.Request().Context(), c.Param("id"), scope.StoreID, scope.BranchID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete item"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Modifier Groups

func (h *MenuHandler) CreateModifierGroup(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}
	if _, err := h.itemRepo.RequireOwnedByBranch(c.Request().Context(), c.Param("id"), scope.StoreID, *branchID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "item does not belong to this branch"})
	}

	var mg model.ModifierGroup
	if err := c.Bind(&mg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mg.ItemID = c.Param("id")
	mg.StoreID = scope.StoreID
	mg.BranchID = *branchID
	if err := h.modifierGroupRepo.Create(c.Request().Context(), &mg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create modifier group"})
	}
	return c.JSON(http.StatusCreated, mg)
}

func (h *MenuHandler) UpdateModifierGroup(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}
	if _, err := h.modifierGroupRepo.RequireOwnedByBranch(c.Request().Context(), c.Param("id"), scope.StoreID, *branchID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "modifier group does not belong to this branch"})
	}

	var mg model.ModifierGroup
	if err := c.Bind(&mg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mg.ID = c.Param("id")
	mg.StoreID = scope.StoreID
	mg.BranchID = *branchID
	if err := h.modifierGroupRepo.Update(c.Request().Context(), &mg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update modifier group"})
	}
	return c.JSON(http.StatusOK, mg)
}

func (h *MenuHandler) DeleteModifierGroup(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "menu:manage"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	if err := h.modifierGroupRepo.Delete(c.Request().Context(), c.Param("id"), scope.StoreID, scope.BranchID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete modifier group"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Modifiers

func (h *MenuHandler) CreateModifier(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}
	if _, err := h.modifierGroupRepo.RequireOwnedByBranch(c.Request().Context(), c.Param("id"), scope.StoreID, *branchID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "modifier group does not belong to this branch"})
	}

	var mod model.Modifier
	if err := c.Bind(&mod); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mod.ModifierGroupID = c.Param("id")
	mod.StoreID = scope.StoreID
	mod.BranchID = *branchID
	mod.IsAvailable = true
	if err := h.modifierRepo.Create(c.Request().Context(), &mod); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create modifier"})
	}
	return c.JSON(http.StatusCreated, mod)
}

func (h *MenuHandler) UpdateModifier(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, true)
	if err != nil {
		return h.menuError(c, err)
	}
	if branchID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch_id is required"})
	}
	if _, err := h.modifierRepo.RequireOwnedByBranch(c.Request().Context(), c.Param("id"), scope.StoreID, *branchID); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "modifier does not belong to this branch"})
	}

	var mod model.Modifier
	if err := c.Bind(&mod); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	mod.ID = c.Param("id")
	mod.StoreID = scope.StoreID
	mod.BranchID = *branchID
	if err := h.modifierRepo.Update(c.Request().Context(), &mod); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update modifier"})
	}
	return c.JSON(http.StatusOK, mod)
}

func (h *MenuHandler) DeleteModifier(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "menu:manage"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
	if err := h.modifierRepo.Delete(c.Request().Context(), c.Param("id"), scope.StoreID, scope.BranchID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete modifier"})
	}
	return c.NoContent(http.StatusNoContent)
}

// Admin full menu (with all items, including unavailable)
func (h *MenuHandler) AdminGetFullMenu(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	branchID, err := h.effectiveBranchID(c, scope, false)
	if err != nil {
		return h.menuError(c, err)
	}

	var menu *model.Menu
	if branchID != nil {
		menu, err = h.menuRepo.GetFullMenuByBranch(c.Request().Context(), *branchID)
	} else {
		menu, err = h.menuRepo.GetFullMenu(c.Request().Context(), scope.StoreID)
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load menu"})
	}
	return c.JSON(http.StatusOK, menu)
}

func (h *MenuHandler) effectiveBranchID(c echo.Context, scope middleware.AdminScope, requireManage bool) (*string, error) {
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "menu:manage"); err != nil {
		return nil, err
	}

	if scope.BranchID != nil {
		return scope.BranchID, nil
	}

	branchID := c.QueryParam("branch_id")
	if branchID == "" {
		return nil, nil
	}
	return &branchID, nil
}

func (h *MenuHandler) menuError(c echo.Context, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return c.JSON(http.StatusNotFound, map[string]string{"error": "resource not found"})
	default:
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
}
