package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageDetectionSummaryRepository struct {
	db *sql.DB
}

func NewDamageDetectionSummaryRepository(db *sql.DB) *DamageDetectionSummaryRepository {
	return &DamageDetectionSummaryRepository{db: db}
}

func (r *DamageDetectionSummaryRepository) Create(ctx context.Context, summary domain.DamageDetectionSummary) (domain.DamageDetectionSummary, error) {
	const sqlStr = `
		INSERT INTO damage_detection_summaries (user_id, condition, condition_note)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, sqlStr, summary.UserID, summary.Condition, summary.ConditionNote).Scan(
		&summary.ID,
		&summary.CreatedAt,
	)
	if err != nil {
		return domain.DamageDetectionSummary{}, apperror.ErrInternal.Wrap(err, "failed to insert damage detection summary")
	}
	return summary, nil
}
