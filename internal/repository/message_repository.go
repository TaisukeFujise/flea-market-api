package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) FindRoomByID(ctx context.Context, id string) (domain.MessageRoom, error) {
	var room domain.MessageRoom
	err := r.db.QueryRowContext(ctx, `
		SELECT id, buyer_id, seller_id FROM message_rooms WHERE id = $1::UUID AND deleted_at IS NULL
	`, id).Scan(&room.ID, &room.BuyerID, &room.SellerID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.MessageRoom{}, apperror.ErrNotFound.New("message room not found")
	}
	if err != nil {
		return domain.MessageRoom{}, apperror.ErrInternal.Wrap(err, "failed to find message room")
	}
	return room, nil
}

func (r *MessageRepository) ListByRoomID(ctx context.Context, roomID string, f domain.MessageFilter) ([]domain.Message, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM messages WHERE room_id = $1::UUID AND deleted_at IS NULL
	`, roomID).Scan(&total); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to count messages")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.room_id, u.id, u.display_name, u.avatar_url, m.content, m.created_at
		FROM messages m
		JOIN users u ON u.id = m.sender_id AND u.deleted_at IS NULL
		WHERE m.room_id = $1::UUID AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC, m.id ASC
		LIMIT $2 OFFSET $3
	`, roomID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to list messages")
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.Sender.ID, &m.Sender.DisplayName, &m.Sender.AvatarURL, &m.Content, &m.CreatedAt); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan message")
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate messages")
	}

	return messages, total, nil
}

func (r *MessageRepository) Create(ctx context.Context, input domain.MessageCreate) (domain.Message, error) {
	var m domain.Message
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO messages (room_id, sender_id, content)
		VALUES ($1::UUID, $2, $3)
		RETURNING id, room_id, sender_id, content, created_at
	`, input.RoomID, input.SenderID, input.Content).Scan(&m.ID, &m.RoomID, &m.Sender.ID, &m.Content, &m.CreatedAt)
	if err != nil {
		return domain.Message{}, apperror.ErrInternal.Wrap(err, "failed to insert message")
	}
	return m, nil
}
