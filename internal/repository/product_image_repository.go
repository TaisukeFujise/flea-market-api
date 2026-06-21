package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductImageRepository struct {
	db *sql.DB
}

func NewProductImageRepository(db *sql.DB) *ProductImageRepository {
	return &ProductImageRepository{db: db}
}

func (r *ProductImageRepository) CreateAll(ctx context.Context, images []domain.ProductImage) ([]string, error) {
	if len(images) == 0 {
		return []string{}, nil
	}
	placeholders := make([]string, len(images))
	args := make([]any, 0, len(images)*3)
	for i, img := range images {
		n := i * 3
		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d)", n+1, n+2, n+3)
		args = append(args, img.SummaryID, img.URL, img.Angle)
	}

	sqlStr := "INSERT INTO product_images (summary_id, url, angle) VALUES " +
		strings.Join(placeholders, ", ") +
		" RETURNING id"

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to insert product images")
	}
	defer rows.Close()

	ids := make([]string, 0, len(images))
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, apperror.ErrInternal.Wrap(err, "failed to scan product image id")
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to iterate product images")
	}
	return ids, nil
}

func (r *ProductImageRepository) GetURLsByProductID(ctx context.Context, productID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT url FROM product_images
		WHERE product_id = $1::UUID
		  AND deleted_at IS NULL
		  AND angle IN ($2, $3, $4, $5)
		ORDER BY CASE angle
			WHEN $2 THEN 1
			WHEN $3 THEN 2
			WHEN $4 THEN 3
			WHEN $5 THEN 4
		END
		LIMIT 4
	`, productID, string(domain.AngleFront), string(domain.AngleBack), string(domain.AngleRight), string(domain.AngleLeft))
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to query product image urls")
	}
	defer rows.Close()

	urls := make([]string, 0)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, apperror.ErrInternal.Wrap(err, "failed to scan product image url")
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to iterate product image urls")
	}
	return urls, nil
}
