package domain

import "time"

type ProductImage struct {
	ID        string
	ProductID *string
	SummaryID *string
	URL       string
	Angle     string
	CreatedAt time.Time
}

type DamageDetectionSummary struct {
	ID            string
	UserID        string
	Condition     string
	ConditionNote string
	CreatedAt     time.Time
}
