package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageReportRepository struct {
	db *sql.DB
}

func NewDamageReportRepository(db *sql.DB) *DamageReportRepository {
	return &DamageReportRepository{db: db}
}

func (r *DamageReportRepository) ValidateImageForProduct(ctx context.Context, imageID, productID string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM product_images
			WHERE id = $1::UUID AND product_id = $2::UUID AND deleted_at IS NULL
		)
	`, imageID, productID).Scan(&exists)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to validate image")
	}
	if !exists {
		return apperror.ErrValidation.New("image_id does not belong to this product or has been deleted")
	}
	return nil
}

func (r *DamageReportRepository) Create(ctx context.Context, input domain.DamageReportCreate, uid string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO damage_reports (product_id, user_id, image_id, damage_type, bbox_x1, bbox_y1, bbox_x2, bbox_y2, description)
		VALUES ($1::UUID, $2, $3::UUID, $4, $5, $6, $7, $8, $9)
	`, input.ProductID, uid, input.ImageID, string(input.DamageType),
		input.BboxX1, input.BboxY1, input.BboxX2, input.BboxY2, input.Description)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to insert damage report")
	}
	return nil
}
