package repository

import (
	"context"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupListingRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:testdb_listing_repo_" + time.Now().Format("150405.000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&model.Exchange{}, &model.Asset{}, &model.Listing{}); err != nil {
		t.Fatal(err)
	}

	exchange := model.Exchange{
		Name:           "Nasdaq",
		Acronym:        "NASDAQ",
		MicCode:        "XNAS",
		Polity:         "USA",
		Currency:       "USD",
		TimeZone:       -4,
		OpenTime:       "09:30",
		CloseTime:      "16:00",
		TradingEnabled: true,
	}

	if err := db.Create(&exchange).Error; err != nil {
		t.Fatal(err)
	}

	return db
}

func TestListingRepositoryUpsert_RejectsUnknownExchangeMIC(t *testing.T) {
	db := setupListingRepositoryTestDB(t)
	repo := NewListingRepository(db)

	// Create a valid asset first
	asset := &model.Asset{
		Ticker:    "AAPL",
		Name:      "Apple Inc",
		AssetType: model.AssetTypeStock,
	}
	if err := db.Create(asset).Error; err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	valid := &model.Listing{
		AssetID:     asset.AssetID,
		ExchangeMIC: "XNAS",
		LastRefresh: time.Now(),
		Price:       150,
		Ask:         151,
	}

	if err := repo.Upsert(context.Background(), valid); err != nil {
		t.Fatalf("expected valid listing insert to succeed: %v", err)
	}

	// Create another asset for the invalid listing
	badAsset := &model.Asset{
		Ticker:    "BADX",
		Name:      "Broken Exchange Listing",
		AssetType: model.AssetTypeStock,
	}
	if err := db.Create(badAsset).Error; err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}

	invalid := &model.Listing{
		AssetID:     badAsset.AssetID,
		ExchangeMIC: "BADX",
		LastRefresh: time.Now(),
		Price:       1,
		Ask:         1,
	}

	if err := repo.Upsert(context.Background(), invalid); err == nil {
		t.Fatal("expected foreign key violation for invalid exchange_mic")
	}
}
