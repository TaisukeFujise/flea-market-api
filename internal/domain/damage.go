package domain

type DamageType string

const (
	DamageTypeScratch DamageType = "scratch"
	DamageTypeDirt    DamageType = "dirt"
	DamageTypeWear    DamageType = "wear"
)

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
