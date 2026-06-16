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

type OrderRole string

const (
	OrderRoleBuyer  OrderRole = "buyer"
	OrderRoleSeller OrderRole = "seller"
)

type OrderProduct struct {
	ID           string
	Title        string
	ThumbnailURL *string
}

type OrderListItem struct {
	ID        string
	Product   OrderProduct
	Price     int
	Status    OrderStatus
	Role      OrderRole
	CreatedAt time.Time
}

type OrderFilter struct {
	Role   *OrderRole
	Limit  int
	Offset int
}

type OrderDetail struct {
	ID            string
	Product       OrderProduct
	BuyerID       string
	SellerID      string
	Price         int
	Status        OrderStatus
	MessageRoomID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
