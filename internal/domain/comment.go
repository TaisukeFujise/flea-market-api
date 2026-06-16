package domain

import "time"

type Comment struct {
	ID              string
	UserID          string
	UserDisplayName string
	UserAvatarURL   *string
	Content         string
	CreatedAt       time.Time
}

type CommentFilter struct {
	Limit  int
	Offset int
}

type CommentCreate struct {
	ProductID string
	UserID    string
	Content   string
}
