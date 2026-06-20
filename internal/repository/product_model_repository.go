package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductModelRepository struct {
	db *sql.DB
}

func NewProductModelRepository(db *sql.DB) *ProductModelRepository {
	return &ProductModelRepository{db: db}
}

func (r *ProductModelRepository) Create(ctx context.Context, productID string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO product_models (product_id, status)
		VALUES ($1::UUID, $2)
		RETURNING id
	`, productID, string(domain.ModelStatusPending)).Scan(&id)
	if err != nil {
		return "", apperror.ErrInternal.Wrap(err, "failed to insert product model")
	}
	return id, nil
}

func (r *ProductModelRepository) UpdateJobID(ctx context.Context, id, jobID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE product_models SET job_id = $1, status = $2, updated_at = NOW()
		WHERE id = $3::UUID
	`, jobID, string(domain.ModelStatusProcessing), id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update product model job_id")
	}
	return nil
}

func (r *ProductModelRepository) UpdateStatus(ctx context.Context, id string, status domain.ModelStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE product_models SET status = $1, updated_at = NOW()
		WHERE id = $2::UUID
	`, string(status), id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update product model status")
	}
	return nil
}

func (r *ProductModelRepository) UpdateDone(ctx context.Context, id, glbURL string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE product_models SET status = $1, glb_url = $2, updated_at = NOW()
		WHERE id = $3::UUID
	`, string(domain.ModelStatusDone), glbURL, id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update product model glb_url")
	}
	return nil
}
