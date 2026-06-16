package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
	"github.com/lib/pq"
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

	wheres := []string{"p.deleted_at IS NULL", fmt.Sprintf("p.status::TEXT = %s", nextArg(string(domain.StatusOnSale)))}

	if f.Query != nil && *f.Query != "" {
		p1 := nextArg("%" + *f.Query + "%")
		p2 := nextArg("%" + *f.Query + "%")
		wheres = append(wheres, fmt.Sprintf("(p.title ILIKE %s OR p.description ILIKE %s)", p1, p2))
	}
	var cte string
	if f.CategoryID != nil {
		cte = fmt.Sprintf(`
			WITH RECURSIVE category_tree AS (
				SELECT id FROM categories WHERE id = %s::UUID
				UNION ALL
				SELECT c.id FROM categories c
				INNER JOIN category_tree ct ON c.parent_id = ct.id
			)
		`, nextArg(*f.CategoryID))
		wheres = append(wheres, "p.category_id IN (SELECT id FROM category_tree)")
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

	sqlStr := cte + fmt.Sprintf(`
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
			pm.status::TEXT,
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
			ms := domain.ModelStatus(modelStatus.String)
			p.ModelStatus = &ms
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

func (r *ProductRepository) GetByID(ctx context.Context, id string, uid *string) (domain.ProductDetail, error) {
	args := []any{id}

	likedExpr := "NULL::boolean"
	if uid != nil {
		args = append(args, *uid)
		likedExpr = fmt.Sprintf("EXISTS(SELECT 1 FROM likes WHERE product_id = p.id AND user_id = $%d)", len(args))
	}

	sqlStr := fmt.Sprintf(`
		SELECT
			p.id,
			p.category_id,
			p.title,
			COALESCE(p.description, ''),
			p.price,
			p.condition::TEXT,
			p.condition_note,
			p.status::TEXT,
			u.id,
			u.display_name,
			u.avatar_url,
			`+ratingsSelectSQL+`,
			pm.status::TEXT,
			pm.glb_url,
			p.created_at,
			p.updated_at,
			%s
		FROM products p
		JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
		`+ratingsJoinSQL+`
		LEFT JOIN LATERAL (
			SELECT status, glb_url
			FROM product_models
			WHERE product_id = p.id AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT 1
		) pm ON TRUE
		WHERE p.id = $1::UUID AND p.deleted_at IS NULL
		GROUP BY p.id, p.category_id, p.title, p.description, p.price, p.condition, p.condition_note, p.status, u.id, u.display_name, u.avatar_url, pm.status, pm.glb_url, p.created_at, p.updated_at
	`, likedExpr)

	var p domain.ProductDetail
	var conditionNote, avatarURL, modelStatus, modelGLBURL sql.NullString
	var sellerRatingAvg sql.NullFloat64
	var liked sql.NullBool
	err := r.db.QueryRowContext(ctx, sqlStr, args...).Scan(
		&p.ID,
		&p.CategoryID,
		&p.Title,
		&p.Description,
		&p.Price,
		&p.Condition,
		&conditionNote,
		&p.Status,
		&p.SellerID,
		&p.SellerName,
		&avatarURL,
		&sellerRatingAvg,
		&p.SellerRatingCount,
		&modelStatus,
		&modelGLBURL,
		&p.CreatedAt,
		&p.UpdatedAt,
		&liked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProductDetail{}, apperror.ErrNotFound.New("product not found")
		}
		return domain.ProductDetail{}, apperror.ErrInternal.Wrap(err, "failed to get product")
	}

	if conditionNote.Valid {
		p.ConditionNote = &conditionNote.String
	}
	if avatarURL.Valid {
		p.SellerAvatarURL = &avatarURL.String
	}
	if sellerRatingAvg.Valid {
		p.SellerRatingAvg = &sellerRatingAvg.Float64
	}
	if modelStatus.Valid {
		ms := domain.ModelStatus(modelStatus.String)
		p.ModelStatus = &ms
	}
	if modelGLBURL.Valid {
		p.ModelGLBURL = &modelGLBURL.String
	}
	if uid != nil {
		p.Liked = &liked.Bool
	}

	images, err := r.getImagesByProductID(ctx, id)
	if err != nil {
		return domain.ProductDetail{}, err
	}
	p.Images = images

	return p, nil
}

func (r *ProductRepository) getImagesByProductID(ctx context.Context, productID string) ([]domain.ProductImage, error) {
	sqlStr := `
		SELECT id, url, angle::TEXT
		FROM product_images
		WHERE product_id = $1::UUID AND deleted_at IS NULL
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, sqlStr, productID)
	if err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to query product images")
	}
	defer rows.Close()

	images := make([]domain.ProductImage, 0)
	for rows.Next() {
		var img domain.ProductImage
		if err := rows.Scan(&img.ID, &img.URL, &img.Angle); err != nil {
			return nil, apperror.ErrInternal.Wrap(err, "failed to scan product image")
		}
		images = append(images, img)
	}
	if err := rows.Err(); err != nil {
		return nil, apperror.ErrInternal.Wrap(err, "failed to iterate product images")
	}
	return images, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string, sellerID string) error {
	var ownerID string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM products
		WHERE id = $1::UUID AND deleted_at IS NULL
	`, id).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrNotFound.New("product not found")
		}
		return apperror.ErrInternal.Wrap(err, "failed to get product")
	}
	if ownerID != sellerID {
		return apperror.ErrForbidden.New("forbidden")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `
		UPDATE products SET deleted_at = NOW()
		WHERE id = $1::UUID
	`, id); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to delete product")
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE product_images SET deleted_at = NOW()
		WHERE product_id = $1::UUID AND deleted_at IS NULL
	`, id); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to delete product images")
	}

	if err := tx.Commit(); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func (r *ProductRepository) Create(ctx context.Context, sellerID string, input domain.ProductCreate) (domain.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var summaryID sql.NullString
	err = tx.QueryRowContext(ctx, `
		SELECT summary_id FROM product_images
		WHERE id = $1 AND deleted_at IS NULL
	`, input.ImageIDs[0]).Scan(&summaryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Product{}, apperror.ErrNotFound.New("image not found")
		}
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to get product image")
	}
	if !summaryID.Valid {
		return domain.Product{}, apperror.ErrBadRequest.New("damage detection not completed")
	}

	var condition domain.ProductCondition
	var conditionNote string
	err = tx.QueryRowContext(ctx, `
		SELECT condition::TEXT, condition_note FROM damage_detection_summaries
		WHERE id = $1
	`, summaryID.String).Scan(&condition, &conditionNote)
	if err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to get damage detection summary")
	}

	var p domain.Product
	err = tx.QueryRowContext(ctx, `
		INSERT INTO products (user_id, category_id, title, description, price, condition, condition_note, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'on_sale')
		RETURNING id
	`, sellerID, input.CategoryID, input.Title, input.Description, input.Price, string(condition), conditionNote).
		Scan(&p.ID)
	if err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to insert product")
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE product_images SET product_id = $1
		WHERE id = ANY($2) AND deleted_at IS NULL
	`, p.ID, pq.Array(input.ImageIDs))
	if err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to update product images")
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to get rows affected")
	}
	if int(updated) != len(input.ImageIDs) {
		return domain.Product{}, apperror.ErrNotFound.New("one or more image_ids not found")
	}

	if err := tx.Commit(); err != nil {
		return domain.Product{}, apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}

	return p, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, sellerID string, input domain.ProductUpdate) error {
	var ownerID string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM products
		WHERE id = $1::UUID AND deleted_at IS NULL
	`, id).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrNotFound.New("product not found")
		}
		return apperror.ErrInternal.Wrap(err, "failed to get product")
	}
	if ownerID != sellerID {
		return apperror.ErrForbidden.New("forbidden")
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE products
		SET
			title       = COALESCE($1, title),
			description = COALESCE($2, description),
			price       = COALESCE($3, price),
			updated_at  = NOW()
		WHERE id = $4::UUID
	`, input.Title, input.Description, input.Price, id)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update product")
	}
	return nil
}
