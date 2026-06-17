package domain

import "time"

type MessageSender struct {
	ID          string
	DisplayName string
	AvatarURL   *string
}

type Message struct {
	ID        string
	RoomID    string
	Sender    MessageSender
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
