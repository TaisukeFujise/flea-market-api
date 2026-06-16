package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID            string
	ProductID     string
	BuyerID       string
	Price         int
	Status        OrderStatus
	MessageRoomID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
