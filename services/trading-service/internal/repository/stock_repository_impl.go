package repository

import (
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

// Upsert inserts or updates a Stock matched by ListingID.
func (r *stockRepository) Upsert(stock *model.Stock) error {
	return r.db.
		Where(model.Stock{ListingID: stock.ListingID}).
		Assign(*stock).
		FirstOrCreate(stock).Error
}
