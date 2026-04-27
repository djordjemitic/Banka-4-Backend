package repository

import (
	"context"
	"errors"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type investmentFundRepository struct {
	db *gorm.DB
}

func NewInvestmentFundRepository(db *gorm.DB) InvestmentFundRepository {
	return &investmentFundRepository{db: db}
}

func (r *investmentFundRepository) Create(ctx context.Context, fund *model.InvestmentFund) error {
	return r.db.WithContext(ctx).Create(fund).Error
}

func (r *investmentFundRepository) FindByID(ctx context.Context, id uint) (*model.InvestmentFund, error) {
	var fund model.InvestmentFund
	result := r.db.WithContext(ctx).First(&fund, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &fund, result.Error
}

func (r *investmentFundRepository) FindByAccountNumber(ctx context.Context, accountNumber string) (*model.InvestmentFund, error) {
	var fund model.InvestmentFund
	result := r.db.WithContext(ctx).Where("account_number = ?", accountNumber).First(&fund)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &fund, result.Error
}

func (r *investmentFundRepository) FindByName(ctx context.Context, name string) (*model.InvestmentFund, error) {
	var fund model.InvestmentFund
	result := r.db.WithContext(ctx).Where("name = ?", name).First(&fund)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &fund, result.Error
}
