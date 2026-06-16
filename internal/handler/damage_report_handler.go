package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type DamageReportService interface {
	Create(ctx context.Context, orderID, uid string, input domain.DamageReportCreate) error
}

type DamageReportHandler struct {
	service DamageReportService
}

func NewDamageReportHandler(s DamageReportService) *DamageReportHandler {
	return &DamageReportHandler{service: s}
}

type damageReportCreateRequest struct {
	ImageID     *string `json:"image_id" validate:"required"`
	DamageType  string  `json:"damage_type" validate:"required,oneof=scratch dirt wear"`
	BboxX1      *int    `json:"bbox_x1" validate:"required"`
	BboxY1      *int    `json:"bbox_y1" validate:"required"`
	BboxX2      *int    `json:"bbox_x2" validate:"required"`
	BboxY2      *int    `json:"bbox_y2" validate:"required"`
	Description *string `json:"description"`
}

func (h *DamageReportHandler) Create(c *echo.Context) error {
	orderID := c.Param("id")
	if _, err := uuid.Parse(orderID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req damageReportCreateRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	if _, err := uuid.Parse(*req.ImageID); err != nil {
		return apperror.ErrValidation.New("invalid image_id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	input := domain.DamageReportCreate{
		ImageID:     req.ImageID,
		DamageType:  domain.DamageType(req.DamageType),
		BboxX1:      req.BboxX1,
		BboxY1:      req.BboxY1,
		BboxX2:      req.BboxX2,
		BboxY2:      req.BboxY2,
		Description: req.Description,
	}

	if err := h.service.Create(c.Request().Context(), orderID, uid, input); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
