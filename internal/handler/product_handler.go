package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type ProductService interface {
	ListProducts(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error)
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(s ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

type productModelResponse struct {
	Status string  `json:"status"`
	GLBURL *string `json:"glb_url"`
}

type productListResponse struct {
	Items  []productListItemResponse `json:"items"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
}

type productListItemResponse struct {
	ID           string                `json:"id"`
	CategoryID   string                `json:"category_id"`
	Title        string                `json:"title"`
	Price        int                   `json:"price"`
	Condition    string                `json:"condition"`
	Status       string                `json:"status"`
	ThumbnailURL *string               `json:"thumbnail_url"`
	Model        *productModelResponse `json:"model"`
	CreatedAt    time.Time             `json:"created_at"`
}

func (h *ProductHandler) GetList(c *echo.Context) error {
	f := domain.ProductFilter{
		Sort:  domain.SortCreatedAtDesc,
		Limit: 20,
	}

	if q := c.QueryParam("q"); q != "" {
		f.Query = &q
	}
	if v := c.QueryParam("category_id"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return apperror.ErrValidation.New("invalid category_id")
		}
		f.CategoryID = &v
	}
	if v := c.QueryParam("min_price"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return apperror.ErrValidation.New("invalid min_price")
		}
		f.MinPrice = &n
	}
	if v := c.QueryParam("max_price"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return apperror.ErrValidation.New("invalid max_price")
		}
		f.MaxPrice = &n
	}
	if v := c.QueryParam("condition"); v != "" {
		c := domain.ProductCondition(v)
		if c != domain.ConditionGood && c != domain.ConditionFair && c != domain.ConditionPoor {
			return apperror.ErrValidation.New("condition must be one of: good, fair, poor")
		}
		f.Condition = &c
	}
	if v := c.QueryParam("sort"); v != "" {
		s := domain.ProductSort(v)
		if s != domain.SortCreatedAtDesc && s != domain.SortPriceAsc && s != domain.SortPriceDesc {
			return apperror.ErrValidation.New("sort must be one of: created_at_desc, price_asc, price_desc")
		}
		f.Sort = s
	}
	if v := c.QueryParam("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return apperror.ErrValidation.New("invalid limit")
		}
		f.Limit = min(n, 100) // 上限の設定 n < 100
	}
	if v := c.QueryParam("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return apperror.ErrValidation.New("invalid offset")
		}
		f.Offset = n
	}

	products, total, err := h.service.ListProducts(c.Request().Context(), f)
	if err != nil {
		return err
	}

	items := make([]productListItemResponse, len(products))
	for i, p := range products {
		var model *productModelResponse
		if p.ModelStatus != nil {
			model = &productModelResponse{
				Status: *p.ModelStatus,
				GLBURL: p.ModelGLBURL,
			}
		}
		items[i] = productListItemResponse{
			ID:           p.ID,
			CategoryID:   p.CategoryID,
			Title:        p.Title,
			Price:        p.Price,
			Condition:    p.Condition,
			Status:       p.Status,
			ThumbnailURL: p.ThumbnailURL,
			Model:        model,
			CreatedAt:    p.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, productListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}
