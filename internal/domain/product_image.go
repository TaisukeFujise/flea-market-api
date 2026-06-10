package domain

import "time"

type ImageAngle string

const (
	AngleFront ImageAngle = "front"
	AngleBack  ImageAngle = "back"
	AngleRight ImageAngle = "right"
	AngleLeft  ImageAngle = "left"
	AngleTop   ImageAngle = "top"
)

type ProductImage struct {
	ID        string
	ProductID *string
	SummaryID *string
	URL       string
	Angle     ImageAngle
	CreatedAt time.Time
}

type DamageDetectionSummary struct {
	ID            string
	UserID        string
	Condition     ProductCondition
	ConditionNote string
	CreatedAt     time.Time
}
