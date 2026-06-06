package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int, error) {
	args := make([]any, 0, 8)
	nextArg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	wheres := []string{"p.deleted_at IS NULL", "p.status = 'on_sale'"}

	if f.Query != nil && *f.Query != "" {
		p1 := nextArg("%" + *f.Query + "%")
		p2 := nextArg("%" + *f.Query + "%")
		wheres = append(wheres, fmt.Sprintf("(p.title ILIKE %s OR p.description ILIKE %s)", p1, p2))
	}
	if f.CategoryID != nil {
		wheres = append(wheres, fmt.Sprintf("p.category_id = %s::UUID", nextArg(*f.CategoryID)))
	}
	if f.MinPrice != nil {
		wheres = append(wheres, fmt.Sprintf("p.price >= %s", nextArg(*f.MinPrice)))
	}
	if f.MaxPrice != nil {
		wheres = append(wheres, fmt.Sprintf("p.price <= %s", nextArg(*f.MaxPrice)))
	}
	if f.Condition != nil {
		wheres = append(wheres, fmt.Sprintf("p.condition::TEXT = %s", nextArg(string(*f.Condition))))
	}

	whereClause := strings.Join(wheres, " AND ")

	var orderBy string
	switch f.Sort {
	case domain.SortCreatedAtDesc:
		orderBy = "p.created_at DESC"
	case domain.SortPriceAsc:
		orderBy = "p.price ASC"
	case domain.SortPriceDesc:
		orderBy = "p.price DESC"
	default:
		orderBy = "p.created_at DESC"
	}

	limitArg := nextArg(f.Limit)
	offsetArg := nextArg(f.Offset)

	sqlStr := fmt.Sprintf(`
		SELECT
			p.id,
			p.category_id,
			p.title,
			p.price,
			p.condition::TEXT,
			p.status::TEXT,
			(
				SELECT pi_t.url
				FROM product_images pi_t
				WHERE pi_t.product_id = p.id AND pi_t.deleted_at IS NULL
				ORDER BY pi_t.created_at
				LIMIT 1
			),
			pm.status,
			pm.glb_url,
			p.created_at,
			COUNT(*) OVER() AS total
		FROM products p
		LEFT JOIN LATERAL (
			SELECT status, glb_url
			FROM product_models
			WHERE product_id = p.id AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		) pm ON TRUE
		WHERE %s
		ORDER BY %s
		LIMIT %s OFFSET %s
	`, whereClause, orderBy, limitArg, offsetArg)

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to query products")
	}
	defer rows.Close()

	var total int
	products := make([]domain.Product, 0)
	for rows.Next() {
		var p domain.Product
		var thumbnailURL, modelStatus, modelGLBURL sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.CategoryID,
			&p.Title,
			&p.Price,
			&p.Condition,
			&p.Status,
			&thumbnailURL,
			&modelStatus,
			&modelGLBURL,
			&p.CreatedAt,
			&total,
		); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan product")
		}
		if thumbnailURL.Valid {
			p.ThumbnailURL = &thumbnailURL.String
		}
		if modelStatus.Valid {
			p.ModelStatus = &modelStatus.String
		}
		if modelGLBURL.Valid {
			p.ModelGLBURL = &modelGLBURL.String
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate products")
	}
	return products, total, nil
}
