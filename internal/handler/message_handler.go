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

type MessageService interface {
	ListByRoomID(ctx context.Context, roomID, uid string, f domain.MessageFilter) ([]domain.Message, int, error)
	Create(ctx context.Context, roomID, uid, content string) (domain.Message, error)
}

type MessageHandler struct {
	service MessageService
}

func NewMessageHandler(s MessageService) *MessageHandler {
	return &MessageHandler{service: s}
}

type messageSenderResponse struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type messageItemResponse struct {
	ID        string                 `json:"id"`
	Sender    messageSenderResponse  `json:"sender"`
	Content   string                 `json:"content"`
	CreatedAt time.Time              `json:"created_at"`
}

type messageListResponse struct {
	Items  []messageItemResponse `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func (h *MessageHandler) GetList(c *echo.Context) error {
	roomID := c.Param("id")
	if _, err := uuid.Parse(roomID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	uid, err := firebaseUID(c)
	if err != nil {
		return err
	}

	limit, offset, err := parsePagination(c, 20)
	if err != nil {
		return err
	}
	f := domain.MessageFilter{Limit: limit, Offset: offset}

	messages, total, err := h.service.ListByRoomID(c.Request().Context(), roomID, uid, f)
	if err != nil {
		return err
	}

	items := make([]messageItemResponse, len(messages))
	for i, m := range messages {
		items[i] = messageItemResponse{
			ID: m.ID,
			Sender: messageSenderResponse{
				ID:          m.Sender.ID,
				DisplayName: m.Sender.DisplayName,
				AvatarURL:   m.Sender.AvatarURL,
			},
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, messageListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	})
}

type messageCreateRequest struct {
	Content string `json:"content" validate:"required,max=10000"`
}

func (h *MessageHandler) Create(c *echo.Context) error {
	roomID := c.Param("id")
	if _, err := uuid.Parse(roomID); err != nil {
		return apperror.ErrValidation.New("invalid id")
	}

	var req messageCreateRequest
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

	msg, err := h.service.Create(c.Request().Context(), roomID, uid, req.Content)
	if err != nil {
		return err
	}

	// TODO(#12): hub.Send(receiverUID, msg) でWebSocket通知
	_ = msg

	return c.NoContent(http.StatusNoContent)
}
