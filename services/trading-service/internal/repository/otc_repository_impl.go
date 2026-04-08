package repository

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type otcRepository struct {
	db *gorm.DB
}

func NewOTCRepository(db *gorm.DB) OTCRepository {
	return &otcRepository{db: db}
}

func (r *otcRepository) FindByOwnershipID(ctx context.Context, ownershipID uint) (*model.OTCInfo, error) {
	var info model.OTCInfo
	result := r.db.WithContext(ctx).Where("asset_ownership_id = ?", ownershipID).First(&info)
	if stderrors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &info, result.Error
}

func (r *otcRepository) FindAllPublic(ctx context.Context, page, pageSize int) ([]model.OTCInfo, int64, error) {
	var infos []model.OTCInfo
	var count int64

	db := r.db.WithContext(ctx).Model(&model.OTCInfo{}).Where("public_count > 0")

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.
		Preload("AssetOwnership").
		Preload("AssetOwnership.Asset").
		Preload("Stock").
		Limit(pageSize).Offset(offset).
		Find(&infos).Error
	return infos, count, err
}

func (r *otcRepository) Upsert(ctx context.Context, info *model.OTCInfo) (*model.OTCInfo, error) {
	info.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "asset_ownership_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"public_count", "updated_at"}),
		}).
		Create(info).Error
	if err != nil {
		return nil, err
	}

	// reload with Stock preloaded
	var saved model.OTCInfo
	if err := r.db.WithContext(ctx).Preload("Stock").First(&saved, info.OTCInfoID).Error; err != nil {
		return nil, err
	}
	return &saved, nil
}
