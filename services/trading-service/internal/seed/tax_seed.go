package seed

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

func SeedAccumulatedTax(db *gorm.DB) error {
	now := time.Now()

	records := []model.AccumulatedTax{
		{
			AccumulatedTaxID: 1,
			AccountNumber:    "444000112345678911",
			TaxOwed:          5000,
			CurrencyCode:     "RSD",
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 2,
			AccountNumber:    "444000112345678913",
			TaxOwed:          45000,
			CurrencyCode:     "RSD",
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 3,
			AccountNumber:    "444000112345678921",
			TaxOwed:          50,
			CurrencyCode:     "EUR",
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 4,
			AccountNumber:    "444000112345678922",
			TaxOwed:          2000,
			CurrencyCode:     "USD",
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 5,
			AccountNumber:    "444000112345678922",
			TaxOwed:          2000,
			CurrencyCode:     "USD",
			LastUpdatedAt:    now,
		},
	}

	for _, r := range records {
		if err := db.FirstOrCreate(&r, model.AccumulatedTax{AccountNumber: r.AccountNumber}).Error; err != nil {
			return err
		}
	}

	return nil
}
