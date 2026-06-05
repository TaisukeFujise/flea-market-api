package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/labstack/echo/v5"
)

type CategoryService interface {
	GetAll(ctx context.Context) ([]domain.Category, error)
}

type CategoryHandler struct {
	service CategoryService
}

func NewCategoryHandler(s CategoryService) *CategoryHandler {
	return &CategoryHandler{service: s}
}

type categoryResponse struct {
	ID       string             `json:"id"`
	ParentID *string            `json:"parent_id"`
	Name     string             `json:"name"`
	Children []categoryResponse `json:"children"`
}

func (h *CategoryHandler) GetAll(c *echo.Context) error {
	categories, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	categoryByID := make(map[string]*categoryResponse, len(categories))
	var roots []*categoryResponse
	for _, cat := range categories {
		node := &categoryResponse{
			ID:       cat.ID,
			ParentID: cat.ParentID,
			Name:     cat.Name,
			Children: []categoryResponse{},
		}
		categoryByID[cat.ID] = node
		if cat.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := categoryByID[*cat.ParentID]; ok {
			parent.Children = append(parent.Children, *node)
		}
	}
	return c.JSON(http.StatusOK, map[string]any{"categories": roots})
}
