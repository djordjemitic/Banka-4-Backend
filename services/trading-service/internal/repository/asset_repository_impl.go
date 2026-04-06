package repository

import (
	"context"
	"errors"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type assetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) AssetRepository {
	return &assetRepository{db: db}
}

func (r *assetRepository) Upsert(ctx context.Context, asset *model.Asset) error {
	return r.db.WithContext(ctx).
		Where(model.Asset{Ticker: asset.Ticker}).
		Assign(*asset).
		FirstOrCreate(asset).Error
}

func (r *assetRepository) FindByTicker(ctx context.Context, ticker string) (*model.Asset, error) {
	var asset model.Asset
	result := r.db.WithContext(ctx).Where("ticker = ?", ticker).First(&asset)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &asset, result.Error
}
