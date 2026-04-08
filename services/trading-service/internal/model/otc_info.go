package model

import "time"

type OTCInfo struct {
	OTCInfoID        uint           `gorm:"primaryKey;autoIncrement"`
	AssetOwnershipID uint           `gorm:"not null;uniqueIndex"`
	AssetOwnership   AssetOwnership
	AssetID          uint   `gorm:"not null"`
	Stock            *Stock `gorm:"foreignKey:AssetID;references:AssetID;constraint:-"`
	PublicCount      float64 `gorm:"not null;default:0"`
	ReservedCount    float64 `gorm:"not null;default:0"`
	UpdatedAt        time.Time
}
