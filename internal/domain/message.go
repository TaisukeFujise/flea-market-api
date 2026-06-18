package domain

import "time"

type Message struct {
	ID        string
	RoomID    string
	Sender    UserSummary
	Content   string
	CreatedAt time.Time
}

type MessageRoom struct {
	ID       string
	BuyerID  string
	SellerID string
}

type MessageFilter struct {
	Limit  int
	Offset int
}

type MessageCreate struct {
	RoomID   string
	SenderID string
	Content  string
}
