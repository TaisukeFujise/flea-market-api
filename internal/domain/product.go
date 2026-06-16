package domain

import "time"

type ProductCondition string

const (
	ConditionGood ProductCondition = "good"
	ConditionFair ProductCondition = "fair"
	ConditionPoor ProductCondition = "poor"
)

type ProductStatus string

const (
	StatusOnSale  ProductStatus = "on_sale"
	StatusSoldOut ProductStatus = "sold_out"
)

type ModelStatus string

const (
	ModelStatusPending    ModelStatus = "pending"
	ModelStatusProcessing ModelStatus = "processing"
	ModelStatusDone       ModelStatus = "done"
	ModelStatusFailed     ModelStatus = "failed"
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
	Condition    ProductCondition
	Status       ProductStatus
	ThumbnailURL *string
	ModelStatus  *ModelStatus
	ModelGLBURL  *string
	CreatedAt    time.Time
}

type ProductDetail struct {
	ID                string
	SellerID          string
	SellerName        string
	SellerAvatarURL   *string
	SellerRatingAvg   *float64
	SellerRatingCount int
	CategoryID      string
	Title           string
	Description     string
	Price           int
	Condition       ProductCondition
	ConditionNote   *string
	Status          ProductStatus
	Images          []ProductImage
	ModelStatus     *ModelStatus
	ModelGLBURL     *string
	Liked           *bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProductCreate struct {
	ImageIDs    []string
	CategoryID  string
	Title       string
	Description string
	Price       int
}

type ProductUpdate struct {
	Title       *string
	Description *string
	Price       *int
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

type Like struct {
	ProductID    string
	Title        string
	Price        int
	ThumbnailURL *string
	Status       ProductStatus
	CreatedAt    time.Time
}

type LikeFilter struct {
	Limit  int
	Offset int
}
