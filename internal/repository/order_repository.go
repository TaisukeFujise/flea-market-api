package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/TaisukeFujise/flea-market-api/internal/apperror"
	"github.com/TaisukeFujise/flea-market-api/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, buyerID, productID string, price int) (domain.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Order{}, apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var sellerID string
	err = tx.QueryRowContext(ctx, `
		UPDATE products SET status = 'sold_out', updated_at = NOW()
		WHERE id = $1::UUID AND status = 'on_sale' AND deleted_at IS NULL
		RETURNING user_id
	`, productID).Scan(&sellerID)
	if err != nil {
		return domain.Order{}, apperror.ErrConflict.New("product is already sold out")
	}

	var o domain.Order
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (product_id, buyer_id, price, status)
		VALUES ($1::UUID, $2, $3, 'pending')
		RETURNING id, product_id, buyer_id, price, status::TEXT, created_at, updated_at
	`, productID, buyerID, price).Scan(
		&o.ID, &o.ProductID, &o.BuyerID, &o.Price, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, apperror.ErrInternal.Wrap(err, "failed to insert order")
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO message_rooms (order_id, buyer_id, seller_id)
		VALUES ($1::UUID, $2, $3)
		RETURNING id
	`, o.ID, buyerID, sellerID).Scan(&o.MessageRoomID)
	if err != nil {
		return domain.Order{}, apperror.ErrInternal.Wrap(err, "failed to insert message room")
	}

	if err := tx.Commit(); err != nil {
		return domain.Order{}, apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}

	return o, nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID string, f domain.OrderFilter) ([]domain.OrderListItem, int, error) {
	var whereClause string
	switch {
	case f.Role != nil && *f.Role == domain.OrderRoleBuyer:
		whereClause = "o.buyer_id = $1"
	case f.Role != nil && *f.Role == domain.OrderRoleSeller:
		whereClause = "p.user_id = $1"
	default:
		whereClause = "(o.buyer_id = $1 OR p.user_id = $1)"
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM orders o
		JOIN products p ON p.id = o.product_id
		WHERE `+whereClause, userID).Scan(&total); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to count orders")
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			o.id,
			p.id,
			p.title,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			o.price,
			o.status::TEXT,
			CASE WHEN o.buyer_id = $1 THEN 'buyer' ELSE 'seller' END,
			o.created_at
		FROM orders o
		JOIN products p ON p.id = o.product_id
		WHERE `+whereClause+`
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to list orders")
	}
	defer rows.Close()

	items := make([]domain.OrderListItem, 0)
	for rows.Next() {
		var item domain.OrderListItem
		var thumbnailURL sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Product.ID,
			&item.Product.Title,
			&thumbnailURL,
			&item.Price,
			&item.Status,
			&item.Role,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan order")
		}
		if thumbnailURL.Valid {
			item.Product.ThumbnailURL = &thumbnailURL.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperror.ErrInternal.Wrap(err, "failed to iterate orders")
	}

	return items, total, nil
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (domain.OrderDetail, error) {
	var o domain.OrderDetail
	var thumbnailURL sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT
			o.id,
			p.id, p.title,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			o.buyer_id,
			mr.seller_id,
			o.price, o.status::TEXT,
			mr.id,
			o.created_at, o.updated_at
		FROM orders o
		JOIN products p ON p.id = o.product_id
		JOIN message_rooms mr ON mr.order_id = o.id AND mr.deleted_at IS NULL
		WHERE o.id = $1::UUID
	`, id).Scan(
		&o.ID,
		&o.Product.ID, &o.Product.Title,
		&thumbnailURL,
		&o.BuyerID,
		&o.SellerID,
		&o.Price, &o.Status,
		&o.MessageRoomID,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.OrderDetail{}, apperror.ErrNotFound.New("order not found")
		}
		return domain.OrderDetail{}, apperror.ErrInternal.Wrap(err, "failed to get order")
	}
	if thumbnailURL.Valid {
		o.Product.ThumbnailURL = &thumbnailURL.String
	}
	return o, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	var productID string
	err = tx.QueryRowContext(ctx, `
		UPDATE orders SET status = $2, updated_at = NOW()
		WHERE id = $1::UUID
		RETURNING product_id
	`, id, string(status)).Scan(&productID)
	if err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to update order status")
	}

	if status == domain.OrderStatusCancelled {
		if _, err := tx.ExecContext(ctx, `
			UPDATE products SET status = 'on_sale', updated_at = NOW()
			WHERE id = $1::UUID
		`, productID); err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to revert product status")
		}
	}

	if err := tx.Commit(); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}
	return nil
}
