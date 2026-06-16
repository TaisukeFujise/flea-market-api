package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type RatingRepository struct {
	db *sql.DB
}

func NewRatingRepository(db *sql.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Create(ctx context.Context, rating domain.RatingCreate) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ratings (order_id, rater_id, ratee_id, score)
		VALUES ($1::UUID, $2, $3, $4)
	`, rating.OrderID, rating.RaterID, rating.RateeID, rating.Score)
	if err != nil {
		if pq.As(err, pqerror.UniqueViolation) != nil {
			return apperror.ErrConflict.New("feedback already submitted")
		}
		return apperror.ErrInternal.Wrap(err, "failed to insert rating")
	}
	return nil
}
