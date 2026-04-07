package model

type Stock struct {
	StockID           uint `gorm:"primaryKey;autoIncrement"`
	AssetID           uint `gorm:"not null;uniqueIndex"`
	Asset             Asset
	Listing           *Listing `gorm:"foreignKey:AssetID;references:AssetID;constraint:-"`
	OutstandingShares float64  `gorm:"not null;default:0"`
	DividendYield     float64  `gorm:"not null;default:0"`
}
