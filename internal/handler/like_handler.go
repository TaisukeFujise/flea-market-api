package handler

import (
	"context"
	"net/http"
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

type likeItemResponse struct {
	Product   productSummaryResponse `json:"product"`
	CreatedAt time.Time              `json:"created_at"`
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

	limit, offset, err := parsePagination(c, 20)
	if err != nil {
		return err
	}
	f := domain.LikeFilter{Limit: limit, Offset: offset}

	likes, total, err := h.service.ListByUserID(c.Request().Context(), uid, f)
	if err != nil {
		return err
	}

	items := make([]likeItemResponse, len(likes))
	for i, l := range likes {
		items[i] = likeItemResponse{
			Product: productSummaryResponse{
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
