package handler

import (
	"context"
	"net/http"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type OrderService interface {
	BuyProduct(ctx context.Context, productID, buyerID string) (domain.Order, error)
}

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(s OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

type orderCreateResponse struct {
	ID            string `json:"id"`
	MessageRoomID string `json:"message_room_id"`
}

func (h *OrderHandler) Create(c *echo.Context) error {
	productID := c.Param("id")
	if _, err := uuid.Parse(productID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	order, err := h.service.BuyProduct(c.Request().Context(), productID, uid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, orderCreateResponse{
		ID:            order.ID,
		MessageRoomID: order.MessageRoomID,
	})
}
