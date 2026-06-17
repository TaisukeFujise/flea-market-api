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
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO damage_detection_summaries (user_id, status)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, summary.UserID, string(summary.Status)).Scan(&summary.ID, &summary.CreatedAt)
	if err != nil {
		return domain.DamageDetectionSummary{}, apperror.ErrInternal.Wrap(err, "failed to insert damage detection summary")
	}
	return summary, nil
}

func (r *DamageDetectionSummaryRepository) Update(ctx context.Context, id string, condition domain.ProductCondition, conditionNote string, status domain.DetectionStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE damage_detection_summaries
		SET condition = $2, condition_note = $3, status = $4
		WHERE id = $1
	`, id, string(condition), conditionNote, string(status))
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update damage detection summary")
	}
	return nil
}
