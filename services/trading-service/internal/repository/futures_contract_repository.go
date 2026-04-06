package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type FuturesContractRepository interface {
	FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.FuturesContract, error)
}
