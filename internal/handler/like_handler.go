package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type LikeService interface {
	Create(ctx context.Context, productID, userID string) error
	Delete(ctx context.Context, productID, userID string) error
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
