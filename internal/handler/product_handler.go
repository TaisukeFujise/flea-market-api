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
	GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error)
	Create(ctx context.Context, sellerID string, input domain.ProductCreate) (domain.Product, error)
	Update(ctx context.Context, id string, sellerID string, input domain.ProductUpdate) error
	Delete(ctx context.Context, id string, sellerID string) error
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(s ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func toModelResponse(status *domain.ModelStatus, glbURL *string) *productModelResponse {
	if status == nil {
		return nil
	}
	return &productModelResponse{Status: string(*status), GLBURL: glbURL}
}

type productSellerResponse struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	AvatarURL   *string  `json:"avatar_url"`
	RatingAvg   *float64 `json:"rating_avg"`
	RatingCount int      `json:"rating_count"`
}

type productModelResponse struct {
	Status string  `json:"status"`
	GLBURL *string `json:"glb_url"`
}

type productImageResponse struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Angle string `json:"angle"`
}

type productDetailResponse struct {
	ID            string                 `json:"id"`
	Seller        productSellerResponse  `json:"seller"`
	CategoryID    string                 `json:"category_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Price         int                    `json:"price"`
	Condition     string                 `json:"condition"`
	ConditionNote *string                `json:"condition_note"`
	Status        string                 `json:"status"`
	Images        []productImageResponse `json:"images"`
	Model         *productModelResponse  `json:"model"`
	Liked         *bool                  `json:"liked"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
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
	if f.MinPrice != nil && f.MaxPrice != nil && *f.MinPrice > *f.MaxPrice {
		return apperror.ErrValidation.New("min_price must be less than or equal to max_price")
	}
	if v := c.QueryParam("condition"); v != "" {
		cond := domain.ProductCondition(v)
		if cond != domain.ConditionGood && cond != domain.ConditionFair && cond != domain.ConditionPoor {
			return apperror.ErrValidation.New("condition must be one of: good, fair, poor")
		}
		f.Condition = &cond
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
		items[i] = productListItemResponse{
			ID:           p.ID,
			CategoryID:   p.CategoryID,
			Title:        p.Title,
			Price:        p.Price,
			Condition:    string(p.Condition),
			Status:       string(p.Status),
			ThumbnailURL: p.ThumbnailURL,
			Model:        toModelResponse(p.ModelStatus, p.ModelGLBURL),
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

func (h *ProductHandler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var uid *string
	if s, ok := c.Get("firebase_uid").(string); ok && s != "" {
		uid = &s
	}

	product, err := h.service.GetByID(c.Request().Context(), id, uid)
	if err != nil {
		return err
	}

	images := make([]productImageResponse, len(product.Images))
	for i, img := range product.Images {
		images[i] = productImageResponse{
			ID:    img.ID,
			URL:   img.URL,
			Angle: string(img.Angle),
		}
	}

	return c.JSON(http.StatusOK, productDetailResponse{
		ID: product.ID,
		Seller: productSellerResponse{
			ID:          product.SellerID,
			DisplayName: product.SellerName,
			AvatarURL:   product.SellerAvatarURL,
			RatingAvg:   product.SellerRatingAvg,
			RatingCount: product.SellerRatingCount,
		},
		CategoryID:    product.CategoryID,
		Title:         product.Title,
		Description:   product.Description,
		Price:         product.Price,
		Condition:     string(product.Condition),
		ConditionNote: product.ConditionNote,
		Status:        string(product.Status),
		Images:        images,
		Model:         toModelResponse(product.ModelStatus, product.ModelGLBURL),
		Liked:         product.Liked,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	})
}

type productCreateRequest struct {
	ImageIDs    []string `json:"image_ids"    validate:"required,min=1,dive,uuid"`
	CategoryID  string   `json:"category_id"  validate:"required,uuid"`
	Title       string   `json:"title"        validate:"required"`
	Description string   `json:"description"  validate:"required"`
	Price       int      `json:"price"        validate:"required,min=1"`
}

type productCreateResponse struct {
	ID string `json:"id"`
}

func (h *ProductHandler) Create(c *echo.Context) error {
	var req productCreateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	product, err := h.service.Create(c.Request().Context(), uid, domain.ProductCreate{
		ImageIDs:    req.ImageIDs,
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, productCreateResponse{
		ID: product.ID,
	})
}

type productUpdateRequest struct {
	Title       *string `json:"title"        validate:"omitempty,min=1"`
	Description *string `json:"description"  validate:"omitempty,min=1"`
	Price       *int    `json:"price"        validate:"omitempty,min=1"`
}

func (h *ProductHandler) Update(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req productUpdateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	if err := h.service.Update(c.Request().Context(), id, uid, domain.ProductUpdate{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
	}); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *ProductHandler) Delete(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, ok := c.Get("firebase_uid").(string)
	if !ok || uid == "" {
		return apperror.ErrUnauthorized.New("unauthorized")
	}

	if err := h.service.Delete(c.Request().Context(), id, uid); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
