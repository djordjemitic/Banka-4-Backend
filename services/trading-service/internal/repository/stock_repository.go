package repository

import "github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"

type StockRepository interface {
	Upsert(stock *model.Stock) error
}
