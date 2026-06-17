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
	if len(damages) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	for _, d := range damages {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO damages (image_id, damage_type, bbox_x1, bbox_y1, bbox_x2, bbox_y2, description)
			VALUES ($1, $2::damage_type, $3, $4, $5, $6, $7)
		`, d.ImageID, string(d.DamageType), d.BboxX1, d.BboxY1, d.BboxX2, d.BboxY2, d.Description)
		if err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to insert damage")
		}
	}

	if err := tx.Commit(); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to commit damages")
	}
	return nil
}
