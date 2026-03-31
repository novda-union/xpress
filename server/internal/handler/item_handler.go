package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/repository"
)

type ItemHandler struct {
	itemRepo          *repository.ItemRepo
	modifierGroupRepo *repository.ModifierGroupRepo
	modifierRepo      *repository.ModifierRepo
	branchRepo        *repository.BranchRepo
}

func NewItemHandler(
	itemRepo *repository.ItemRepo,
	modifierGroupRepo *repository.ModifierGroupRepo,
	modifierRepo *repository.ModifierRepo,
	branchRepo *repository.BranchRepo,
) *ItemHandler {
	return &ItemHandler{
		itemRepo:          itemRepo,
		modifierGroupRepo: modifierGroupRepo,
		modifierRepo:      modifierRepo,
		branchRepo:        branchRepo,
	}
}

type itemModifierResponse struct {
	ID              string `json:"id"`
	ModifierGroupID string `json:"modifier_group_id"`
	Name            string `json:"name"`
	PriceAdjustment int64  `json:"price_adjustment"`
	IsAvailable     bool   `json:"is_available"`
	SortOrder       int    `json:"sort_order"`
}

type itemModifierGroupResponse struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	SelectionType string                 `json:"selection_type"`
	IsRequired    bool                   `json:"is_required"`
	MinSelections int                    `json:"min_selections"`
	MaxSelections int                    `json:"max_selections"`
	Modifiers     []itemModifierResponse `json:"modifiers"`
}

type itemDetailResponse struct {
	ID             string                      `json:"id"`
	CategoryID     string                      `json:"category_id"`
	StoreID        string                      `json:"store_id"`
	BranchID       string                      `json:"branch_id"`
	Name           string                      `json:"name"`
	Description    string                      `json:"description"`
	BasePrice      int64                       `json:"base_price"`
	ImageURL       string                      `json:"image_url"`
	IsAvailable    bool                        `json:"is_available"`
	ModifierGroups []itemModifierGroupResponse `json:"modifier_groups"`
}

func (h *ItemHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()
	itemID := c.Param("id")
	branchID := c.QueryParam("branch")
	if branchID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "branch query param required")
	}

	branch, err := h.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "branch not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	item, err := h.itemRepo.GetByIDForBranch(ctx, itemID, branchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "item not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	groups, err := h.modifierGroupRepo.ListByItem(ctx, item.ID, item.StoreID, &item.BranchID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	groupResponses := make([]itemModifierGroupResponse, len(groups))
	for gi, group := range groups {
		modifiers, err := h.modifierRepo.ListByGroup(ctx, group.ID, item.StoreID, &item.BranchID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		modResponses := make([]itemModifierResponse, len(modifiers))
		for mi, modifier := range modifiers {
			modResponses[mi] = itemModifierResponse{
				ID:              modifier.ID,
				ModifierGroupID: modifier.ModifierGroupID,
				Name:            modifier.Name,
				PriceAdjustment: modifier.PriceAdjustment,
				IsAvailable:     modifier.IsAvailable,
				SortOrder:       modifier.SortOrder,
			}
		}

		groupResponses[gi] = itemModifierGroupResponse{
			ID:            group.ID,
			Name:          group.Name,
			SelectionType: group.SelectionType,
			IsRequired:    group.IsRequired,
			MinSelections: group.MinSelections,
			MaxSelections: group.MaxSelections,
			Modifiers:     modResponses,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"item": itemDetailResponse{
			ID:             item.ID,
			CategoryID:     item.CategoryID,
			StoreID:        item.StoreID,
			BranchID:       item.BranchID,
			Name:           item.Name,
			Description:    item.Description,
			BasePrice:      item.BasePrice,
			ImageURL:       item.ImageURL,
			IsAvailable:    item.IsAvailable,
			ModifierGroups: groupResponses,
		},
		"branch": branch,
	})
}
