package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type RatingService interface {
	Create(ctx context.Context, orderID, uid string, score int) error
}

type RatingHandler struct {
	service RatingService
}

func NewRatingHandler(s RatingService) *RatingHandler {
	return &RatingHandler{service: s}
}

type ratingCreateRequest struct {
	Score int `json:"score" validate:"required,min=1,max=5"`
}

func (h *RatingHandler) Create(c *echo.Context) error {
	orderID := c.Param("id")
	if _, err := uuid.Parse(orderID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req ratingCreateRequest
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

	if err := h.service.Create(c.Request().Context(), orderID, uid, req.Score); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
