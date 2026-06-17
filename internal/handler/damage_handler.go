package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/labstack/echo/v5"
)

type DamageService interface {
	ListByProductID(ctx context.Context, productID string) ([]domain.Damage, error)
}

type DamageHandler struct {
	service DamageService
}

func NewDamageHandler(s DamageService) *DamageHandler {
	return &DamageHandler{service: s}
}

type damageResponse struct {
	ID          string   `json:"id"`
	ImageID     string   `json:"image_id"`
	DamageType  string   `json:"damage_type"`
	BboxX1      *int     `json:"bbox_x1"`
	BboxY1      *int     `json:"bbox_y1"`
	BboxX2      *int     `json:"bbox_x2"`
	BboxY2      *int     `json:"bbox_y2"`
	Description *string  `json:"description"`
	ModelX      *float64 `json:"model_x"`
	ModelY      *float64 `json:"model_y"`
	ModelZ      *float64 `json:"model_z"`
}

type listDamagesResponse struct {
	Damages []damageResponse `json:"damages"`
}

func (h *DamageHandler) ListByProductID(c *echo.Context) error {
	productID := c.Param("id")

	damages, err := h.service.ListByProductID(c.Request().Context(), productID)
	if err != nil {
		return err
	}

	resp := make([]damageResponse, len(damages))
	for i, d := range damages {
		resp[i] = damageResponse{
			ID:          d.ID,
			ImageID:     d.ImageID,
			DamageType:  string(d.DamageType),
			BboxX1:      d.BboxX1,
			BboxY1:      d.BboxY1,
			BboxX2:      d.BboxX2,
			BboxY2:      d.BboxY2,
			Description: d.Description,
			ModelX:      d.ModelX,
			ModelY:      d.ModelY,
			ModelZ:      d.ModelZ,
		}
	}
	return c.JSON(http.StatusOK, listDamagesResponse{Damages: resp})
}
