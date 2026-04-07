package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type OptionRepository interface {
	Upsert(ctx context.Context, option *model.Option) error
	FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.Option, error)
	FindByStockID(ctx context.Context, stockID uint) ([]model.Option, error)
}
