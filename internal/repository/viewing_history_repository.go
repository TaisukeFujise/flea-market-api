package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
)

type ViewingHistoryRepository struct {
	db *sql.DB
}

func NewViewingHistoryRepository(db *sql.DB) *ViewingHistoryRepository {
	return &ViewingHistoryRepository{db: db}
}

func (r *ViewingHistoryRepository) Upsert(ctx context.Context, userID, productID string) error {
	sqlStr := `
		INSERT INTO viewing_history (user_id, product_id)
		VALUES ($1, $2::UUID)
		ON CONFLICT (user_id, product_id) DO UPDATE SET viewed_at = NOW()
	`
	if _, err := r.db.ExecContext(ctx, sqlStr, userID, productID); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to upsert viewing history")
	}
	return nil
}
