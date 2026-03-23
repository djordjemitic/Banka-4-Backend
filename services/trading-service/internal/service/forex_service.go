package service

import (
	"context"
	"log"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

const refreshInterval = 1 * time.Hour

type ForexService struct {
	repo   repository.ForexRepository
	client client.ExchangeRateClient
}

func NewForexService(repo repository.ForexRepository, client client.ExchangeRateClient) *ForexService {
	return &ForexService{
		repo:   repo,
		client: client,
	}
}

func (s *ForexService) Initialize(ctx context.Context) {
	var count int64

	count, err := s.repo.Count(ctx)
	if err != nil {
		log.Println("failed counting forex pairs:", err)
		return
	}

	if count > 0 {
		log.Println("forex pairs loaded from DB")
		return
	}

	if err := s.refreshFromAPI(ctx); err != nil {
		log.Println("initial forex fetch failed:", err)
	}
}

func (s *ForexService) StartBackgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := s.refreshFromAPI(ctx); err != nil {
					log.Println("forex refresh failed:", err)
				}
			}
		}
	}()
}

func (s *ForexService) refreshFromAPI(ctx context.Context) error {
	resp, err := s.client.FetchRates(ctx)
	if err != nil {
		return err
	}

	providerUpdatedAt := time.Unix(resp.TimeLastUpdateUnix, 0)
	providerNextUpdateAt := time.Unix(resp.TimeNextUpdateUnix, 0)

	// sve valute koje podržava banka
	supported := []string{"EUR", "USD", "CHF", "GBP", "JPY", "CAD", "AUD", "RSD"}

	rates := resp.ConversionRates
	// Ako base valuta API-ja nije u map, dodaj je sa 1.0
	rates[resp.BaseCode] = 1.0

	for _, base := range supported {
		for _, quote := range supported {
			if base == quote {
				continue
			}

			baseRate, ok1 := rates[base]
			quoteRate, ok2 := rates[quote]

			if !ok1 || !ok2 {
				// ako nemamo vrednost za neku od valuta, preskoči
				continue
			}

			// formula za konverziju: base→quote = quoteRate / baseRate
			pair := model.ForexPair{
				Base:                 base,
				Quote:                quote,
				Rate:                 quoteRate / baseRate,
				ProviderUpdatedAt:    providerUpdatedAt,
				ProviderNextUpdateAt: providerNextUpdateAt,
			}

			if err := s.repo.Upsert(ctx, pair); err != nil {
				return err
			}
		}
	}

	log.Println("forex pairs refreshed from API")
	return nil
}
