package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type LikeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Create(ctx context.Context, productID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		WITH product_check AS (
			SELECT id FROM products WHERE id = $1::UUID AND deleted_at IS NULL
		)
		INSERT INTO likes (product_id, user_id)
		SELECT $1::UUID, $2 FROM product_check
	`, productID, userID)
	if err != nil {
		if pq.As(err, pqerror.UniqueViolation) != nil {
			return apperror.ErrConflict.New("already liked")
		}
		return apperror.ErrInternal.Wrap(err, "failed to insert like")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return apperror.ErrNotFound.New("product not found")
	}
	return nil
}

func (r *LikeRepository) ListByUserID(ctx context.Context, userID string, f domain.LikeFilter) ([]domain.Like, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM likes l
		JOIN products p ON l.product_id = p.id AND p.deleted_at IS NULL
		WHERE l.user_id = $1
	`, userID).Scan(&total); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to count likes")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.price,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			p.status::TEXT, l.created_at
		FROM likes l
		JOIN products p ON l.product_id = p.id AND p.deleted_at IS NULL
		WHERE l.user_id = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to list likes")
	}
	defer rows.Close()

	likes := make([]domain.Like, 0)
	for rows.Next() {
		var like domain.Like
		var thumbnailURL sql.NullString
		if err := rows.Scan(&like.ProductID, &like.Title, &like.Price, &thumbnailURL, &like.Status, &like.CreatedAt); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan like")
		}
		if thumbnailURL.Valid {
			like.ThumbnailURL = &thumbnailURL.String
		}
		likes = append(likes, like)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate likes")
	}

	return likes, total, nil
}

func (r *LikeRepository) Delete(ctx context.Context, productID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM likes
		USING products p
		WHERE likes.product_id = $1::UUID
		  AND likes.user_id = $2
		  AND likes.product_id = p.id
		  AND p.deleted_at IS NULL
	`, productID, userID)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to delete like")
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return apperror.ErrNotFound.New("like not found")
	}
	return nil
}
