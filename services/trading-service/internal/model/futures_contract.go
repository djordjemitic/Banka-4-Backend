package model

import "time"

type FuturesContract struct {
	FuturesContractID uint `gorm:"primaryKey;autoIncrement"`
	AssetID           uint `gorm:"not null;uniqueIndex"`
	Asset             Asset
	Listing           *Listing  `gorm:"foreignKey:AssetID;references:AssetID;constraint:-"`
	ContractSize      float64   `gorm:"not null"`
	ContractUnit      string    `gorm:"not null"`
	SettlementDate    time.Time `gorm:"not null"`
}
