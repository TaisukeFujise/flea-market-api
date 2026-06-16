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

type OrderService interface {
	BuyProduct(ctx context.Context, productID, buyerID string) (domain.Order, error)
	ListOrders(ctx context.Context, userID string, f domain.OrderFilter) ([]domain.OrderListItem, int, error)
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

type orderProductResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	ThumbnailURL *string `json:"thumbnail_url"`
}

type orderListItemResponse struct {
	ID        string               `json:"id"`
	Product   orderProductResponse `json:"product"`
	Price     int                  `json:"price"`
	Status    string               `json:"status"`
	Role      string               `json:"role"`
	CreatedAt time.Time            `json:"created_at"`
}

type orderListResponse struct {
	Items  []orderListItemResponse `json:"items"`
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
}

func (h *OrderHandler) GetList(c *echo.Context) error {
	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	f := domain.OrderFilter{}

	if v := c.QueryParam("role"); v != "" {
		role := domain.OrderRole(v)
		if role != domain.OrderRoleBuyer && role != domain.OrderRoleSeller {
			return apperror.ErrValidation.New("role must be buyer or seller")
		}
		f.Role = &role
	}

	f.Limit, f.Offset, err = parsePagination(c, 20)
	if err != nil {
		return err
	}

	orders, total, err := h.service.ListOrders(c.Request().Context(), uid, f)
	if err != nil {
		return err
	}

	items := make([]orderListItemResponse, len(orders))
	for i, o := range orders {
		items[i] = orderListItemResponse{
			ID: o.ID,
			Product: orderProductResponse{
				ID:           o.Product.ID,
				Title:        o.Product.Title,
				ThumbnailURL: o.Product.ThumbnailURL,
			},
			Price:     o.Price,
			Status:    string(o.Status),
			Role:      string(o.Role),
			CreatedAt: o.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, orderListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}
