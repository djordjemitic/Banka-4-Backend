package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type OTCRepository interface {
	FindByOwnershipID(ctx context.Context, ownershipID uint) (*model.OTCInfo, error)
	FindAllPublic(ctx context.Context, page, pageSize int) ([]model.OTCInfo, int64, error)
	Upsert(ctx context.Context, info *model.OTCInfo) (*model.OTCInfo, error)
}
