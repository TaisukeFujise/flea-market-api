package domain

import "time"

type User struct {
	ID           string
	DisplayName  string
	AvatarURL    *string
	RatingAvg    *float64
	RatingCount  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type UserUpdate struct {
	DisplayName *string
}
