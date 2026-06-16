package domain

import "time"

type Rating struct {
	ID        string
	OrderID   string
	RaterID   string
	RateeID   string
	Score     int
	CreatedAt time.Time
}
