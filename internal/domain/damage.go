package domain

type DamageType string

const (
	DamageTypeScratch DamageType = "scratch"
	DamageTypeDirt    DamageType = "dirt"
	DamageTypeWear    DamageType = "wear"
)

type DetectionStatus string

const (
	DetectionStatusProcessing DetectionStatus = "processing"
	DetectionStatusDone       DetectionStatus = "done"
	DetectionStatusFailed     DetectionStatus = "failed"
)

type Damage struct {
	ID          string
	ImageID     string
	DamageType  DamageType
	BboxX1      *int
	BboxY1      *int
	BboxX2      *int
	BboxY2      *int
	Description *string
	ModelX      *float64
	ModelY      *float64
	ModelZ      *float64
}

// DamageCreate は AI 検出結果を damages テーブルに保存するための入力。
// 画像アップロード後の非同期処理で使用する。商品作成前に実行されるため ProductID を持たない。
type DamageCreate struct {
	ImageID     string
	DamageType  DamageType
	BboxX1      *int
	BboxY1      *int
	BboxX2      *int
	BboxY2      *int
	Description *string
}

type DamageModelCoordinatesUpdate struct {
	ModelX float64
	ModelY float64
	ModelZ float64
}

// DamageReportCreate は購入者が受け取り後に報告する傷を damage_reports テーブルに保存するための入力。
// damages テーブル（AI検出）とは別テーブル。
type DamageReportCreate struct {
	ProductID   string
	ImageID     string
	DamageType  DamageType
	BboxX1      *int
	BboxY1      *int
	BboxX2      *int
	BboxY2      *int
	Description *string
}
