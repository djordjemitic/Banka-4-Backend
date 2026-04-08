package seed

import (
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func SeedPublicOTCStocks(db *gorm.DB) error {
	type entry struct {
		ticker      string
		identityID  uint
		ownerType   model.OwnerType
		totalAmount float64 // total held
		publicCount float64 // how many to list publicly
	}

	entries := []entry{
		{ticker: "AAPL", identityID: 4, ownerType: model.OwnerTypeActuary, totalAmount: 200, publicCount: 50},
		{ticker: "AAPL", identityID: 9, ownerType: model.OwnerTypeClient, totalAmount: 80, publicCount: 20},
		{ticker: "MSFT", identityID: 5, ownerType: model.OwnerTypeActuary, totalAmount: 150, publicCount: 40},
		{ticker: "MSFT", identityID: 10, ownerType: model.OwnerTypeClient, totalAmount: 60, publicCount: 15},
		{ticker: "TSLA", identityID: 6, ownerType: model.OwnerTypeActuary, totalAmount: 100, publicCount: 30},
		{ticker: "TSLA", identityID: 11, ownerType: model.OwnerTypeClient, totalAmount: 40, publicCount: 10},
		{ticker: "AMZN", identityID: 4, ownerType: model.OwnerTypeActuary, totalAmount: 50, publicCount: 12},
		{ticker: "GOOGL", identityID: 5, ownerType: model.OwnerTypeActuary, totalAmount: 75, publicCount: 20},
		{ticker: "NVDA", identityID: 9, ownerType: model.OwnerTypeClient, totalAmount: 120, publicCount: 35},
	}

	now := time.Now()

	for _, e := range entries {
		var asset model.Asset
		if err := db.Where("ticker = ? AND asset_type = ?", e.ticker, model.AssetTypeStock).First(&asset).Error; err != nil {
			log.Printf("otc seed: asset %q not found, skipping", e.ticker)
			continue
		}

		ownership := model.AssetOwnership{
			IdentityID:     e.identityID,
			OwnerType:      e.ownerType,
			AssetID:        asset.AssetID,
			Amount:         e.totalAmount,
			AvgBuyPriceRSD: 0,
			UpdatedAt:      now,
		}

		if err := db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "identity_id"}, {Name: "owner_type"}, {Name: "asset_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"amount", "updated_at"}),
			}).
			Create(&ownership).Error; err != nil {
			return err
		}

		if err := db.Where("identity_id = ? AND owner_type = ? AND asset_id = ?",
			e.identityID, e.ownerType, asset.AssetID).
			First(&ownership).Error; err != nil {
			return err
		}

		otcInfo := model.OTCInfo{
			AssetOwnershipID: ownership.AssetOwnershipID,
			AssetID:          asset.AssetID,
			PublicCount:      e.publicCount,
			ReservedCount:    0,
			UpdatedAt:        now,
		}

		if err := db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "asset_ownership_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"public_count", "updated_at"}),
			}).
			Create(&otcInfo).Error; err != nil {
			return err
		}

		log.Printf("otc seed: %s identity=%d (%s) public=%g/%g",
			e.ticker, e.identityID, e.ownerType, e.publicCount, e.totalAmount)
	}

	return nil
}
