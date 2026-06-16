package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
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

func (r *ViewingHistoryRepository) ListByUserID(ctx context.Context, userID string, f domain.ViewingHistoryFilter) ([]domain.ViewingHistory, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM viewing_history vh
		JOIN products p ON vh.product_id = p.id AND p.deleted_at IS NULL
		WHERE vh.user_id = $1
	`, userID).Scan(&total); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to count viewing history")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.price,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			p.status::TEXT, vh.viewed_at
		FROM viewing_history vh
		JOIN products p ON vh.product_id = p.id AND p.deleted_at IS NULL
		WHERE vh.user_id = $1
		ORDER BY vh.viewed_at DESC
		LIMIT $2 OFFSET $3
	`, userID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to list viewing history")
	}
	defer rows.Close()

	histories := make([]domain.ViewingHistory, 0)
	for rows.Next() {
		var vh domain.ViewingHistory
		var thumbnailURL sql.NullString
		if err := rows.Scan(&vh.ProductID, &vh.Title, &vh.Price, &thumbnailURL, &vh.Status, &vh.ViewedAt); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan viewing history")
		}
		if thumbnailURL.Valid {
			vh.ThumbnailURL = &thumbnailURL.String
		}
		histories = append(histories, vh)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate viewing history")
	}

	return histories, total, nil
}
