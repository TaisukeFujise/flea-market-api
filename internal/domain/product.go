package domain

import "time"

type ProductCondition string

const (
	ConditionGood ProductCondition = "good"
	ConditionFair ProductCondition = "fair"
	ConditionPoor ProductCondition = "poor"
)

type ProductSort string

const (
	SortCreatedAtDesc ProductSort = "created_at_desc"
	SortPriceAsc      ProductSort = "price_asc"
	SortPriceDesc     ProductSort = "price_desc"
)

type Product struct {
	ID           string
	CategoryID   string
	Title        string
	Price        int
	Condition    string
	Status       string
	ThumbnailURL *string
	ModelStatus  *string
	ModelGLBURL  *string
	CreatedAt    time.Time
}

type ProductFilter struct {
	Query      *string
	CategoryID *string
	MinPrice   *int
	MaxPrice   *int
	Condition  *ProductCondition
	Sort       ProductSort
	Limit      int
	Offset     int
}
