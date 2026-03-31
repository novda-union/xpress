package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xpressgo/server/internal/repository"
)

type DiscoverHandler struct {
	itemRepo *repository.ItemRepo
}

func NewDiscoverHandler(itemRepo *repository.ItemRepo) *DiscoverHandler {
	return &DiscoverHandler{itemRepo: itemRepo}
}

type discoverItemResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	ImageURL             string   `json:"image_url"`
	BasePrice            int64    `json:"base_price"`
	IsAvailable          bool     `json:"is_available"`
	CreatedAt            string   `json:"created_at"`
	OrderCount           int      `json:"order_count"`
	HasRequiredModifiers bool     `json:"has_required_modifiers"`
	BranchID             string   `json:"branch_id"`
	BranchName           string   `json:"branch_name"`
	BranchAddress        string   `json:"branch_address"`
	Lat                  *float64 `json:"lat"`
	Lng                  *float64 `json:"lng"`
	StoreID              string   `json:"store_id"`
	StoreName            string   `json:"store_name"`
	StoreCategory        string   `json:"store_category"`
}

type discoverSectionResponse struct {
	Title string                 `json:"title"`
	Type  string                 `json:"type"`
	Items []discoverItemResponse `json:"items"`
}

type discoverFeedResponse struct {
	Sections []discoverSectionResponse `json:"sections"`
}

type discoverItemsResponse struct {
	Items []discoverItemResponse `json:"items"`
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

func toDiscoverItemResponse(d repository.DiscoverItem) discoverItemResponse {
	return discoverItemResponse{
		ID:                   d.ID,
		Name:                 d.Name,
		Description:          d.Description,
		ImageURL:             d.ImageURL,
		BasePrice:            d.BasePrice,
		IsAvailable:          d.IsAvailable,
		CreatedAt:            d.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		OrderCount:           d.OrderCount,
		HasRequiredModifiers: d.HasRequiredModifiers,
		BranchID:             d.BranchID,
		BranchName:           d.BranchName,
		BranchAddress:        d.BranchAddress,
		Lat:                  d.Lat,
		Lng:                  d.Lng,
		StoreID:              d.StoreID,
		StoreName:            d.StoreName,
		StoreCategory:        d.StoreCategory,
	}
}

func toDiscoverItemResponses(items []repository.DiscoverItem) []discoverItemResponse {
	out := make([]discoverItemResponse, len(items))
	for i, item := range items {
		out[i] = toDiscoverItemResponse(item)
	}
	return out
}

func (h *DiscoverHandler) Feed(c echo.Context) error {
	ctx := c.Request().Context()

	newItems, err := h.itemRepo.GetFeedSection(ctx, "new", 10)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	popularItems, err := h.itemRepo.GetFeedSection(ctx, "popular", 10)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, discoverFeedResponse{
		Sections: []discoverSectionResponse{
			{Title: "New Arrivals", Type: "new", Items: toDiscoverItemResponses(newItems)},
			{Title: "Popular Right Now", Type: "popular", Items: toDiscoverItemResponses(popularItems)},
		},
	})
}

func (h *DiscoverHandler) Items(c echo.Context) error {
	ctx := c.Request().Context()

	category := c.QueryParam("category")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "new"
	}

	page := 1
	if rawPage := c.QueryParam("page"); rawPage != "" {
		if nextPage, err := strconv.Atoi(rawPage); err == nil && nextPage > 0 {
			page = nextPage
		}
	}

	limit := 20
	if rawLimit := c.QueryParam("limit"); rawLimit != "" {
		if nextLimit, err := strconv.Atoi(rawLimit); err == nil && nextLimit > 0 && nextLimit <= 50 {
			limit = nextLimit
		}
	}

	items, total, err := h.itemRepo.ListForFeed(ctx, category, sort, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, discoverItemsResponse{
		Items: toDiscoverItemResponses(items),
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
