package repository

import (
	"context"
	"database/sql"
	"errors"

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
		JOIN users u ON c.user_id = u.id AND u.deleted_at IS NULL
		WHERE c.product_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC, c.id ASC
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

func (r *CommentRepository) Create(ctx context.Context, input domain.CommentCreate) (domain.Comment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Comment{}, apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM products WHERE id = $1::UUID AND deleted_at IS NULL)
	`, input.ProductID).Scan(&exists); err != nil {
		return domain.Comment{}, apperror.ErrInternal.Wrap(err, "failed to check product existence")
	}
	if !exists {
		return domain.Comment{}, apperror.ErrNotFound.New("product not found")
	}

	var c domain.Comment
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO comments (product_id, user_id, content)
		VALUES ($1::UUID, $2, $3)
		RETURNING id, content, created_at
	`, input.ProductID, input.UserID, input.Content).Scan(&c.ID, &c.Content, &c.CreatedAt); err != nil {
		return domain.Comment{}, apperror.ErrInternal.Wrap(err, "failed to insert comment")
	}

	if err := tx.Commit(); err != nil {
		return domain.Comment{}, apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}
	return c, nil
}

func (r *CommentRepository) GetOwnerID(ctx context.Context, id string) (string, error) {
	var userID string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM comments WHERE id = $1::UUID AND deleted_at IS NULL
	`, id).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", apperror.ErrNotFound.New("comment not found")
	}
	if err != nil {
		return "", apperror.ErrInternal.Wrap(err, "failed to get comment owner")
	}
	return userID, nil
}

func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE comments SET deleted_at = NOW() WHERE id = $1::UUID AND deleted_at IS NULL
	`, id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to delete comment")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return apperror.ErrNotFound.New("comment not found")
	}
	return nil
}
