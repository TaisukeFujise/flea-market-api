package repository

import (
	"context"
	"database/sql"

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
