package service

import (
	"context"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type MessageRepository interface {
	FindRoomByID(ctx context.Context, id string) (domain.MessageRoom, error)
	ListByRoomID(ctx context.Context, roomID string, f domain.MessageFilter) ([]domain.Message, int, error)
	Create(ctx context.Context, input domain.MessageCreate) error
}

type MessageNotifier interface {
	NotifyNewMessage(userID string, roomID string)
}

type MessageService struct {
	repo     MessageRepository
	notifier MessageNotifier
}

func NewMessageService(r MessageRepository, n MessageNotifier) *MessageService {
	return &MessageService{repo: r, notifier: n}
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

func (s *MessageService) Create(ctx context.Context, roomID, uid, content string) error {
	room, err := s.repo.FindRoomByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room.BuyerID != uid && room.SellerID != uid {
		return apperror.ErrForbidden.New("not a participant")
	}
	if err = s.repo.Create(ctx, domain.MessageCreate{
		RoomID:   roomID,
		SenderID: uid,
		Content:  content,
	}); err != nil {
		return err
	}

	recipientID := room.SellerID
	if uid == room.SellerID {
		recipientID = room.BuyerID
	}
	s.notifier.NotifyNewMessage(recipientID, roomID)

	return nil
}
