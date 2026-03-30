package service

import (
	"fmt"

	"github.com/xpressgo/server/internal/middleware"
)

type PermissionService struct{}

func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

func (p *PermissionService) Can(role string, branchID *string, action string, targetRole ...string) bool {
	if middleware.IsDirector(role) {
		return true
	}

	switch action {
	case "branch:create", "branch:edit", "branch:delete", "staff:create:manager", "settings:store", "orders:view:all", "dashboard:all":
		return false
	case "staff:create:barista", "menu:manage", "settings:branch":
		return middleware.IsManager(role) && branchID != nil
	case "orders:view", "dashboard:branch":
		return middleware.IsBranchScoped(role) && branchID != nil
	case "staff:edit":
		if !middleware.IsManager(role) || branchID == nil {
			return false
		}
		if len(targetRole) == 0 {
			return false
		}
		return targetRole[0] == "barista"
	default:
		return false
	}
}

func (p *PermissionService) Require(role string, branchID *string, action string, targetRole ...string) error {
	if p.Can(role, branchID, action, targetRole...) {
		return nil
	}

	return errPermissionDenied(action)
}

func (p *PermissionService) Scope(role string, branchID *string) middleware.AdminScope {
	return middleware.AdminScope{
		Role:     role,
		BranchID: branchID,
	}
}

func (p *PermissionService) IsStoreWide(role string, branchID *string) bool {
	return middleware.IsDirector(role) || branchID == nil
}

func (p *PermissionService) IsBranchScoped(role string, branchID *string) bool {
	return middleware.IsBranchScoped(role) && branchID != nil
}

func errPermissionDenied(action string) error {
	return fmt.Errorf("permission denied: %s", action)
}
