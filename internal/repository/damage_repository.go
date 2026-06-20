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

func (r *DamageRepository) ListByProductID(ctx context.Context, productID string) ([]domain.Damage, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.image_id, d.damage_type,
		       d.bbox_x1, d.bbox_y1, d.bbox_x2, d.bbox_y2,
		       d.description, d.model_x, d.model_y, d.model_z
		FROM damages d
		LEFT JOIN product_images pi ON d.image_id = pi.id AND pi.deleted_at IS NULL
		WHERE pi.product_id = $1
		  AND d.deleted_at IS NULL
		ORDER BY d.id
	`, productID)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to query damages")
	}
	defer rows.Close()

	var damages []domain.Damage
	for rows.Next() {
		var d domain.Damage
		var dt string
		if err := rows.Scan(
			&d.ID, &d.ImageID, &dt,
			&d.BboxX1, &d.BboxY1, &d.BboxX2, &d.BboxY2,
			&d.Description, &d.ModelX, &d.ModelY, &d.ModelZ,
		); err != nil {
			return nil, apperror.ErrInternal.Wrap(err, "failed to scan damage")
		}
		d.DamageType = domain.DamageType(dt)
		damages = append(damages, d)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to iterate damages")
	}
	return damages, nil
}

func (r *DamageRepository) UpdateModelCoordinates(ctx context.Context, id string, userID string, input domain.DamageModelCoordinatesUpdate) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE damages d
		SET model_x = $1,
		    model_y = $2,
		    model_z = $3
		FROM product_images pi
		JOIN products p ON p.id = pi.product_id
		WHERE d.id = $4::UUID
		  AND d.image_id = pi.id
		  AND p.user_id = $5
		  AND d.deleted_at IS NULL
		  AND pi.deleted_at IS NULL
		  AND p.deleted_at IS NULL
	`, input.ModelX, input.ModelY, input.ModelZ, id, userID)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update damage model coordinates")
	}
	n, err := result.RowsAffected()
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to get affected damage rows")
	}
	if n == 0 {
		return apperror.ErrNotFound.New("damage not found")
	}
	return nil
}
