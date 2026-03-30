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
	"golang.org/x/crypto/bcrypt"
)

type StaffHandler struct {
	staffRepo         *repository.StaffRepo
	branchRepo        *repository.BranchRepo
	permissionService *service.PermissionService
}

func NewStaffHandler(staffRepo *repository.StaffRepo, branchRepo *repository.BranchRepo, permissionService *service.PermissionService) *StaffHandler {
	return &StaffHandler{
		staffRepo:         staffRepo,
		branchRepo:        branchRepo,
		permissionService: permissionService,
	}
}

type createStaffRequest struct {
	Name      string  `json:"name"`
	StaffCode string  `json:"staff_code"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
	BranchID  *string `json:"branch_id"`
}

type updateStaffRequest struct {
	Name      string  `json:"name"`
	StaffCode string  `json:"staff_code"`
	Password  string  `json:"password"`
	Role      string  `json:"role"`
	BranchID  *string `json:"branch_id"`
}

func (h *StaffHandler) List(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	if scope.BranchID == nil {
		if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:edit", "manager"); err != nil && !middleware.IsDirector(scope.Role) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
	} else if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:create:barista"); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}

	groups, err := h.staffRepo.ListByScope(c.Request().Context(), scope.StoreID, scope.BranchID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list staff"})
	}
	return c.JSON(http.StatusOK, groups)
}

func (h *StaffHandler) Create(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req createStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	branchID, err := h.resolveStaffBranch(c, scope, req.Role, req.BranchID, true)
	if err != nil {
		return h.staffError(c, err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
	}

	staff := &model.Staff{
		StoreID:      scope.StoreID,
		BranchID:     branchID,
		StaffCode:    req.StaffCode,
		Name:         req.Name,
		PasswordHash: string(passwordHash),
		Role:         req.Role,
		IsActive:     true,
	}
	if err := h.staffRepo.Create(c.Request().Context(), staff); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create staff"})
	}
	return c.JSON(http.StatusCreated, staff)
}

func (h *StaffHandler) Update(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	current, err := h.staffRepo.GetByID(c.Request().Context(), c.Param("id"), scope.StoreID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "staff not found"})
	}

	var req updateStaffRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	branchID, err := h.resolveStaffBranch(c, scope, req.Role, req.BranchID, false)
	if err != nil {
		return h.staffError(c, err)
	}
	if branchID == nil && req.Role != "director" {
		branchID = current.BranchID
	}

	passwordHash := ""
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		}
		passwordHash = string(hash)
	}

	if scope.BranchID != nil {
		if current.Role != "barista" || (current.BranchID != nil && *current.BranchID != *scope.BranchID) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied: staff:edit"})
		}
	}

	staff := &model.Staff{
		ID:           current.ID,
		StoreID:      current.StoreID,
		BranchID:     branchID,
		StaffCode:    req.StaffCode,
		Name:         req.Name,
		PasswordHash: passwordHash,
		Role:         req.Role,
		IsActive:     current.IsActive,
	}
	if err := h.staffRepo.Update(c.Request().Context(), staff); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update staff"})
	}
	return c.JSON(http.StatusOK, staff)
}

func (h *StaffHandler) Delete(c echo.Context) error {
	scope, ok := middleware.AdminScopeFromContext(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	staff, err := h.staffRepo.GetByID(c.Request().Context(), c.Param("id"), scope.StoreID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "staff not found"})
	}

	if middleware.IsDirector(scope.Role) {
		if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:edit", staff.Role); err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
	} else {
		if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:edit", staff.Role); err != nil {
			return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
		}
		if staff.BranchID == nil || scope.BranchID == nil || *staff.BranchID != *scope.BranchID {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied: staff:edit"})
		}
	}

	if err := h.staffRepo.Deactivate(c.Request().Context(), staff.ID, scope.StoreID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to deactivate staff"})
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *StaffHandler) resolveStaffBranch(c echo.Context, scope middleware.AdminScope, role string, requestedBranchID *string, creating bool) (*string, error) {
	switch role {
	case "director":
		if creating {
			return nil, errors.New("director can only be created by the system")
		}
		if !middleware.IsDirector(scope.Role) {
			return nil, errors.New("permission denied: staff:create:manager")
		}
		return nil, nil
	case "manager":
		if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:create:manager"); err != nil {
			return nil, err
		}
	case "barista":
		if err := h.permissionService.Require(scope.Role, scope.BranchID, "staff:create:barista"); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid role")
	}

	if scope.BranchID != nil {
		return scope.BranchID, nil
	}
	if requestedBranchID == nil || *requestedBranchID == "" {
		if creating {
			return nil, errors.New("branch_id is required")
		}
		return nil, nil
	}

	branch, err := h.branchRepo.GetByIDForStore(c.Request().Context(), *requestedBranchID, scope.StoreID)
	if err != nil {
		return nil, pgx.ErrNoRows
	}
	return &branch.ID, nil
}

func (h *StaffHandler) staffError(c echo.Context, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "branch not found"})
	default:
		return c.JSON(http.StatusForbidden, map[string]string{"error": err.Error()})
	}
}
