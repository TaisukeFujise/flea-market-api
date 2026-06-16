package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) ListByProductID(ctx context.Context, productID string, f domain.CommentFilter) ([]domain.Comment, int, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM products WHERE id = $1::UUID AND deleted_at IS NULL)
	`, productID).Scan(&exists); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to check product existence")
	}
	if !exists {
		return nil, 0, apperror.ErrNotFound.New("product not found")
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT c.id, u.id, u.display_name, u.avatar_url, c.content, c.created_at, COUNT(*) OVER()
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.product_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
		LIMIT $2 OFFSET $3
	`, productID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to list comments")
	}
	defer rows.Close()

	var total int
	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.UserID, &c.UserDisplayName, &c.UserAvatarURL, &c.Content, &c.CreatedAt, &total); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan comment")
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate comments")
	}

	return comments, total, nil
}
