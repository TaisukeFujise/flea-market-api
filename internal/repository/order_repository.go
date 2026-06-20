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
		UPDATE products SET status = $2::product_status, updated_at = NOW()
		WHERE id = $1::UUID AND status = $3::product_status AND deleted_at IS NULL
		RETURNING user_id
	`, productID, string(domain.StatusSoldOut), string(domain.StatusOnSale)).Scan(&sellerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Order{}, apperror.ErrConflict.New("product is already sold out")
		}
		return domain.Order{}, apperror.ErrInternal.Wrap(err, "failed to update product status")
	}

	var o domain.Order
	err = tx.QueryRowContext(ctx, `
		INSERT INTO orders (product_id, buyer_id, seller_id, price, status)
		VALUES ($1::UUID, $2, $3, $4, $5::order_status)
		RETURNING id, product_id, buyer_id, price, status::TEXT, created_at, updated_at
	`, productID, buyerID, sellerID, price, string(domain.OrderStatusPending)).Scan(
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
		whereClause = "o.seller_id = $1"
	default:
		whereClause = "(o.buyer_id = $1 OR o.seller_id = $1)"
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM orders o
		JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL
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
			CASE WHEN o.buyer_id = $1 THEN '`+string(domain.OrderRoleBuyer)+`' ELSE '`+string(domain.OrderRoleSeller)+`' END,
			o.created_at,
			CASE WHEN o.buyer_id = $1 THEN su.id      ELSE bu.id           END,
			CASE WHEN o.buyer_id = $1 THEN su.display_name ELSE bu.display_name END,
			CASE WHEN o.buyer_id = $1 THEN su.avatar_url   ELSE bu.avatar_url   END,
			(o.status = 'completed'::order_status AND r.id IS NOT NULL)
		FROM orders o
		JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL
		LEFT JOIN users bu ON bu.id = o.buyer_id AND bu.deleted_at IS NULL
		LEFT JOIN users su ON su.id = o.seller_id AND su.deleted_at IS NULL
		LEFT JOIN ratings r ON r.order_id = o.id AND r.rater_id = $1
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
		var counterpartID sql.NullString
		var counterpartDisplayName sql.NullString
		var counterpartAvatarURL sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.Product.ID,
			&item.Product.Title,
			&thumbnailURL,
			&item.Price,
			&item.Status,
			&item.Role,
			&item.CreatedAt,
			&counterpartID,
			&counterpartDisplayName,
			&counterpartAvatarURL,
			&item.HasFeedback,
		); err != nil {
			return nil, 0, apperror.ErrInternal.Wrap(err, "failed to scan order")
		}
		if thumbnailURL.Valid {
			item.Product.ThumbnailURL = &thumbnailURL.String
		}
		if counterpartID.Valid {
			item.Counterpart.ID = counterpartID.String
		}
		if counterpartDisplayName.Valid {
			item.Counterpart.DisplayName = counterpartDisplayName.String
		}
		if counterpartAvatarURL.Valid {
			item.Counterpart.AvatarURL = &counterpartAvatarURL.String
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
	var messageRoomID sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT
			o.id,
			p.id, p.title,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			o.buyer_id,
			o.seller_id,
			o.price, o.status::TEXT,
			mr.id,
			o.created_at, o.updated_at
		FROM orders o
		JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL
		LEFT JOIN message_rooms mr ON mr.order_id = o.id AND mr.deleted_at IS NULL
		WHERE o.id = $1::UUID
	`, id).Scan(
		&o.ID,
		&o.Product.ID, &o.Product.Title,
		&thumbnailURL,
		&o.BuyerID,
		&o.SellerID,
		&o.Price, &o.Status,
		&messageRoomID,
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
	if messageRoomID.Valid {
		o.MessageRoomID = messageRoomID.String
	}
	return o, nil
}

func (r *OrderRepository) FindByIDForUser(ctx context.Context, id, uid string) (domain.OrderDetail, error) {
	var o domain.OrderDetail
	var thumbnailURL sql.NullString
	var messageRoomID sql.NullString
	var counterpartID sql.NullString
	var counterpartDisplayName sql.NullString
	var counterpartAvatarURL sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT
			o.id,
			p.id, p.title,
			(SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.deleted_at IS NULL AND pi.angle = 'front' LIMIT 1),
			o.buyer_id,
			o.seller_id,
			CASE WHEN o.buyer_id = $2 THEN '`+string(domain.OrderRoleBuyer)+`' ELSE '`+string(domain.OrderRoleSeller)+`' END,
			CASE WHEN o.buyer_id = $2 THEN su.id           ELSE bu.id           END,
			CASE WHEN o.buyer_id = $2 THEN su.display_name ELSE bu.display_name END,
			CASE WHEN o.buyer_id = $2 THEN su.avatar_url   ELSE bu.avatar_url   END,
			o.price, o.status::TEXT,
			(o.status = 'completed'::order_status AND r.id IS NOT NULL),
			mr.id,
			o.created_at, o.updated_at
		FROM orders o
		JOIN products p ON p.id = o.product_id AND p.deleted_at IS NULL
		LEFT JOIN message_rooms mr ON mr.order_id = o.id AND mr.deleted_at IS NULL
		LEFT JOIN users bu ON bu.id = o.buyer_id AND bu.deleted_at IS NULL
		LEFT JOIN users su ON su.id = o.seller_id AND su.deleted_at IS NULL
		LEFT JOIN ratings r ON r.order_id = o.id AND r.rater_id = $2
		WHERE o.id = $1::UUID
	`, id, uid).Scan(
		&o.ID,
		&o.Product.ID, &o.Product.Title,
		&thumbnailURL,
		&o.BuyerID,
		&o.SellerID,
		&o.Role,
		&counterpartID,
		&counterpartDisplayName,
		&counterpartAvatarURL,
		&o.Price, &o.Status,
		&o.HasFeedback,
		&messageRoomID,
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
	if counterpartID.Valid {
		o.Counterpart.ID = counterpartID.String
	}
	if counterpartDisplayName.Valid {
		o.Counterpart.DisplayName = counterpartDisplayName.String
	}
	if counterpartAvatarURL.Valid {
		o.Counterpart.AvatarURL = &counterpartAvatarURL.String
	}
	if messageRoomID.Valid {
		o.MessageRoomID = messageRoomID.String
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
		UPDATE orders SET status = $2::order_status, updated_at = NOW()
		WHERE id = $1::UUID AND status = $3::order_status
		RETURNING product_id
	`, id, string(status), string(domain.OrderStatusPending)).Scan(&productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrConflict.New("order status has already been updated")
		}
		return apperror.ErrInternal.Wrap(err, "failed to update order status")
	}

	if status == domain.OrderStatusCancelled {
		if _, err := tx.ExecContext(ctx, `
			UPDATE products SET status = $2::product_status, updated_at = NOW()
			WHERE id = $1::UUID AND deleted_at IS NULL
		`, productID, string(domain.StatusOnSale)); err != nil {
			return apperror.ErrInternal.Wrap(err, "failed to revert product status")
		}
	}

	if err := tx.Commit(); err != nil {
		return apperror.ErrInternal.Wrap(err, "failed to commit transaction")
	}
	return nil
}
