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

type LikeService interface {
	Create(ctx context.Context, productID, userID string) error
	Delete(ctx context.Context, productID, userID string) error
	ListByUserID(ctx context.Context, userID string, f domain.LikeFilter) ([]domain.Like, int, error)
}

type LikeHandler struct {
	service LikeService
}

func NewLikeHandler(s LikeService) *LikeHandler {
	return &LikeHandler{service: s}
}

func (h *LikeHandler) Create(c *echo.Context) error {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	if err := h.service.Create(c.Request().Context(), productID, uid); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *LikeHandler) Delete(c *echo.Context) error {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Request().Context(), productID, uid); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

type likeProductResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Price        int     `json:"price"`
	ThumbnailURL *string `json:"thumbnail_url"`
	Status       string  `json:"status"`
}

type likeItemResponse struct {
	Product   likeProductResponse `json:"product"`
	CreatedAt time.Time           `json:"created_at"`
}

type likesListResponse struct {
	Items  []likeItemResponse `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}

func (h *LikeHandler) GetLikes(c *echo.Context) error {
	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	f := domain.LikeFilter{Limit: 20}

	if v := c.QueryParam("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return apperror.ErrValidation.New("invalid limit")
		}
		if n > 100 {
			n = 100
		}
		f.Limit = n
	}
	if v := c.QueryParam("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return apperror.ErrValidation.New("invalid offset")
		}
		f.Offset = n
	}

	likes, total, err := h.service.ListByUserID(c.Request().Context(), uid, f)
	if err != nil {
		return err
	}

	items := make([]likeItemResponse, len(likes))
	for i, l := range likes {
		items[i] = likeItemResponse{
			Product: likeProductResponse{
				ID:           l.ProductID,
				Title:        l.Title,
				Price:        l.Price,
				ThumbnailURL: l.ThumbnailURL,
				Status:       string(l.Status),
			},
			CreatedAt: l.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, likesListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}
