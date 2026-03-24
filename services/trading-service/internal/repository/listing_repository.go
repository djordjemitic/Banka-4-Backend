package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type ListingRepository interface {
	FindAll() ([]model.Listing, error)
	Upsert(listing *model.Listing) error
	UpdatePriceAndAsk(listing *model.Listing, price, ask float64) error
	Count(ctx context.Context) (int64, error)
}
