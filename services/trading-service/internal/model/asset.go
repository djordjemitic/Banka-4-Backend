package model

type AssetType string

const (
	AssetTypeStock     AssetType = "stock"
	AssetTypeOption    AssetType = "option"
	AssetTypeFuture    AssetType = "future"
	AssetTypeForexPair AssetType = "forexPair"
)

type Asset struct {
	AssetID   uint      `gorm:"primaryKey;autoIncrement"`
	Ticker    string    `gorm:"not null;uniqueIndex;size:20"`
	Name      string    `gorm:"not null"`
	AssetType AssetType `gorm:"not null;size:10"`
}
