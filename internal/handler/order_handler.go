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
	GetOrder(ctx context.Context, id, uid string) (domain.OrderDetail, error)
	UpdateOrderStatus(ctx context.Context, id, uid string, status domain.OrderStatus) error
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

type orderDetailResponse struct {
	ID            string               `json:"id"`
	Product       orderProductResponse `json:"product"`
	BuyerID       string               `json:"buyer_id"`
	Price         int                  `json:"price"`
	Status        string               `json:"status"`
	MessageRoomID string               `json:"message_room_id"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

func (h *OrderHandler) GetByID(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	order, err := h.service.GetOrder(c.Request().Context(), id, uid)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, orderDetailResponse{
		ID: order.ID,
		Product: orderProductResponse{
			ID:           order.Product.ID,
			Title:        order.Product.Title,
			ThumbnailURL: order.Product.ThumbnailURL,
		},
		BuyerID:       order.BuyerID,
		Price:         order.Price,
		Status:        string(order.Status),
		MessageRoomID: order.MessageRoomID,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	})
}

type orderUpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

func (h *OrderHandler) UpdateStatus(c *echo.Context) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req orderUpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	status := domain.OrderStatus(req.Status)
	if status != domain.OrderStatusCompleted && status != domain.OrderStatusCancelled {
		return apperror.ErrValidation.New("status must be completed or cancelled")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	if err := h.service.UpdateOrderStatus(c.Request().Context(), id, uid, status); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
