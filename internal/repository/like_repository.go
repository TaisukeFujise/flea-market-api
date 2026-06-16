package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
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

func (r *LikeRepository) Delete(ctx context.Context, productID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM likes WHERE product_id = $1::UUID AND user_id = $2
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
