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

type categoryListResponse struct {
	Categories []*categoryResponse `json:"categories"`
}

type categoryResponse struct {
	ID       string              `json:"id"`
	ParentID *string             `json:"parent_id"`
	Name     string              `json:"name"`
	Children []*categoryResponse `json:"children"`
}

func (h *CategoryHandler) GetAll(c *echo.Context) error {
	categories, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return err
	}

	categoryByID := make(map[string]*categoryResponse, len(categories))
	roots := make([]*categoryResponse, 0)
	for _, category := range categories {
		node := &categoryResponse{
			ID:       category.ID,
			ParentID: category.ParentID,
			Name:     category.Name,
			Children: []*categoryResponse{},
		}
		categoryByID[category.ID] = node
	}
	for _, node := range categoryByID {
		if node.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := categoryByID[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}
	return c.JSON(http.StatusOK, categoryListResponse{Categories: roots})
}
