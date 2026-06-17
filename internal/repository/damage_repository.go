package repository

import (
	"context"
	"database/sql"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type DamageRepository struct {
	db *sql.DB
}

func NewDamageRepository(db *sql.DB) *DamageRepository {
	return &DamageRepository{db: db}
}

func (r *DamageRepository) CreateAll(ctx context.Context, damages []domain.DamageCreate) error {
	for _, d := range damages {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO damages (image_id, damage_type, bbox_x1, bbox_y1, bbox_x2, bbox_y2, description)
			VALUES ($1, $2::damage_type, $3, $4, $5, $6, $7)
		`, d.ImageID, string(d.DamageType), d.BboxX1, d.BboxY1, d.BboxX2, d.BboxY2, d.Description)
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to insert damage")
		}
	}
	return nil
}
