package domain

import "time"

type UserSummary struct {
	ID          string
	DisplayName string
	AvatarURL   *string
}

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
