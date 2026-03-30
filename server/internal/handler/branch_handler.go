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

type BranchHandler struct {
	branchRepo        *repository.BranchRepo
	menuRepo          *repository.MenuRepo
	permissionService *service.PermissionService
}

func NewBranchHandler(branchRepo *repository.BranchRepo, menuRepo *repository.MenuRepo, permissionService *service.PermissionService) *BranchHandler {
	return &BranchHandler{branchRepo: branchRepo, menuRepo: menuRepo, permissionService: permissionService}
}

func (h *BranchHandler) Discover(c echo.Context) error {
	category := c.QueryParam("category")
	branches, err := h.branchRepo.ListDiscover(c.Request().Context(), category)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load discovery list"})
	}
	return c.JSON(http.StatusOK, branches)
}

func (h *BranchHandler) GetByID(c echo.Context) error {
	detail, err := h.branchRepo.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}
	return c.JSON(http.StatusOK, detail)
}

func (h *BranchHandler) GetMenu(c echo.Context) error {
	detail, err := h.branchRepo.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}

	menu, err := h.menuRepo.GetFullMenuByBranch(c.Request().Context(), detail.Branch.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load menu"})
	}

	return c.JSON(http.StatusOK, menu)
}

type adminBranchRequest struct {
	Name                string   `json:"name"`
	Address             string   `json:"address"`
	Lat                 *float64 `json:"lat"`
	Lng                 *float64 `json:"lng"`
	BannerImageURL      string   `json:"banner_image_url"`
	TelegramGroupChatID *int64   `json:"telegram_group_chat_id"`
	IsActive            bool     `json:"is_active"`
}

func (h *BranchHandler) AdminList(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	branches, err := h.branchRepo.ListAdmin(c.Request().Context(), scope.StoreID, scope.BranchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list branches"})
	}
	return c.JSON(http.StatusOK, branches)
}

func (h *BranchHandler) AdminCreate(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "branch:create"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	var req adminBranchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	branch := &model.Branch{
		StoreID:             scope.StoreID,
		Name:                req.Name,
		Address:             req.Address,
		Lat:                 req.Lat,
		Lng:                 req.Lng,
		BannerImageURL:      req.BannerImageURL,
		TelegramGroupChatID: req.TelegramGroupChatID,
		IsActive:            req.IsActive,
	}
	if err := h.branchRepo.Create(c.Request().Context(), branch); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create branch"})
	}
	return c.JSON(http.StatusCreated, branch)
}

func (h *BranchHandler) AdminUpdate(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "branch:edit"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	current, err := h.branchRepo.GetByIDForStore(c.Request().Context(), c.Param("id"), scope.StoreID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
	}

	var req adminBranchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	current.Name = req.Name
	current.Address = req.Address
	current.Lat = req.Lat
	current.Lng = req.Lng
	current.BannerImageURL = req.BannerImageURL
	current.TelegramGroupChatID = req.TelegramGroupChatID
	current.IsActive = req.IsActive

	if err := h.branchRepo.Update(c.Request().Context(), current); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update branch"})
	}
	return c.JSON(http.StatusOK, current)
}

func (h *BranchHandler) AdminDelete(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	if err := h.permissionService.Require(scope.Role, scope.BranchID, "branch:delete"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	if _, err := h.branchRepo.GetByIDForStore(c.Request().Context(), c.Param("id"), scope.StoreID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "branch not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load branch"})
	}
	if err := h.branchRepo.Deactivate(c.Request().Context(), c.Param("id"), scope.StoreID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to deactivate branch"})
	}
	return c.NoContent(http.StatusNoContent)
}
