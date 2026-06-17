package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type MessageRepository interface {
	FindRoomByID(ctx context.Context, id string) (domain.MessageRoom, error)
	ListByRoomID(ctx context.Context, roomID string, f domain.MessageFilter) ([]domain.Message, int, error)
	Create(ctx context.Context, input domain.MessageCreate) (domain.Message, error)
}

type MessageService struct {
	repo MessageRepository
}

func NewMessageService(r MessageRepository) *MessageService {
	return &MessageService{repo: r}
}

func (s *MessageService) ListByRoomID(ctx context.Context, roomID, uid string, f domain.MessageFilter) ([]domain.Message, int, error) {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, 0, err
	}
	if room.BuyerID != uid && room.SellerID != uid {
		return nil, 0, apperror.ErrForbidden.New("not a participant")
	}
	return s.repo.ListByRoomID(ctx, roomID, f)
}

func (s *MessageService) Create(ctx context.Context, roomID, uid, content string) (domain.Message, error) {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return domain.Message{}, err
	}
	if room.BuyerID != uid && room.SellerID != uid {
		return domain.Message{}, apperror.ErrForbidden.New("not a participant")
	}
	return s.repo.Create(ctx, domain.MessageCreate{
		RoomID:   roomID,
		SenderID: uid,
		Content:  content,
	})
}
