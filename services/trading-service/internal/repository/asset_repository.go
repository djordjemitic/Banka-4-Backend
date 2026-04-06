package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type AssetRepository interface {
	Upsert(ctx context.Context, asset *model.Asset) error
	FindByTicker(ctx context.Context, ticker string) (*model.Asset, error)
}
