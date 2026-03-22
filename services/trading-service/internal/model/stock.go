package model

type Stock struct {
	StockID           uint `gorm:"primaryKey;autoIncrement"`
	ListingID         uint `gorm:"not null;uniqueIndex"`
	Listing           Listing
	OutstandingShares int64   `gorm:"not null;default:0"`
	DividendYield     float64 `gorm:"not null;default:0"`
}
