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

type CommentService interface {
	ListByProductID(ctx context.Context, productID string, f domain.CommentFilter) ([]domain.Comment, int, error)
	Create(ctx context.Context, input domain.CommentCreate) (domain.Comment, error)
	Delete(ctx context.Context, id string, uid string) error
}

type CommentHandler struct {
	service CommentService
}

func NewCommentHandler(s CommentService) *CommentHandler {
	return &CommentHandler{service: s}
}

type commentUserResponse struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type commentItemResponse struct {
	ID        string              `json:"id"`
	User      commentUserResponse `json:"user"`
	Content   string              `json:"content"`
	CreatedAt time.Time           `json:"created_at"`
}

type commentListResponse struct {
	Items  []commentItemResponse `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func (h *CommentHandler) GetList(c *echo.Context) error {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	f := domain.CommentFilter{Limit: 20}

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

	comments, total, err := h.service.ListByProductID(c.Request().Context(), productID, f)
	if err != nil {
		return err
	}

	items := make([]commentItemResponse, len(comments))
	for i, cm := range comments {
		items[i] = commentItemResponse{
			ID: cm.ID,
			User: commentUserResponse{
				ID:          cm.UserID,
				DisplayName: cm.UserDisplayName,
				AvatarURL:   cm.UserAvatarURL,
			},
			Content:   cm.Content,
			CreatedAt: cm.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, commentListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}

type commentCreateRequest struct {
	Content string `json:"content" validate:"required,max=1000"`
}

type commentCreateResponse struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *CommentHandler) Create(c *echo.Context) error {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req commentCreateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	comment, err := h.service.Create(c.Request().Context(), domain.CommentCreate{
		ProductID: productID,
		UserID:    uid,
		Content:   req.Content,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, commentCreateResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	})
}

func (h *CommentHandler) Delete(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Request().Context(), id, uid); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
