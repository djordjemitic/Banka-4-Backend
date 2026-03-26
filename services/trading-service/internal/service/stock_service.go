package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

const (
	maxCallsPerMinute    = 55
	priceRefreshInterval = 15 * time.Minute
)

type StockService struct {
	listingRepo repository.ListingRepository
	stockRepo   repository.StockRepository
	optionRepo  repository.OptionRepository
	client      *client.StockClient

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewStockService(
	listingRepo repository.ListingRepository,
	stockRepo repository.StockRepository,
	optionRepo repository.OptionRepository,
	client *client.StockClient,
) *StockService {
	return &StockService{
		listingRepo: listingRepo,
		stockRepo:   stockRepo,
		optionRepo:  optionRepo,
		client:      client,
	}
}

func (s *StockService) Initialize(ctx context.Context) {
	count, err := s.listingRepo.Count(ctx)
	if err != nil {
		log.Printf("[seed] failed to count listings: %v", err)
		return
	}

	if count > 0 {
		log.Printf("[seed] listings already exist (%d), skipping seed", count)
		return
	}

	if err := s.SeedStocks(ctx, 10); err != nil {
		log.Printf("[seed] stock seed failed: %v", err)
		return
	}

	if err := s.SeedOptions(ctx, 5); err != nil {
		log.Printf("[seed] option seed failed: %v", err)
	}
}

func (s *StockService) Start() {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	ticker := time.NewTicker(priceRefreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RefreshPrices(ctx); err != nil {
					log.Println("[refresh] failed:", err)
				}
				if err := s.RefreshOptions(ctx); err != nil {
					log.Println("[refresh-options] failed:", err)
				}
			}
		}
	}()
}

func (s *StockService) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *StockService) SeedStocks(ctx context.Context, limit int) error {
	symbols, err := s.client.GetSymbols("US")
	if err != nil {
		return fmt.Errorf("failed to fetch symbols: %w", err)
	}

	log.Printf("[seed] fetched %d symbols, seeding up to %d", len(symbols), limit)

	callsThisMinute := 1
	minuteStart := time.Now()
	count := 0

	for _, sym := range symbols {
		if count >= limit {
			break
		}

		if strings.ContainsRune(sym.Symbol, '.') {
			continue
		}

		if callsThisMinute+3 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				log.Printf("[seed] rate limit reached, waiting %s...", wait.Round(time.Second))
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		profile, err := s.client.GetProfile(sym.Symbol)
		callsThisMinute++
		if err != nil {
			log.Printf("[seed] skipping %s: profile error: %v", sym.Symbol, err)
			continue
		}
		if profile.Name == "" {
			log.Printf("[seed] skipping %s: empty profile", sym.Symbol)
			continue
		}

		quote, err := s.client.GetQuote(sym.Symbol)
		callsThisMinute++
		if err != nil {
			log.Printf("[seed] skipping %s: quote error: %v", sym.Symbol, err)
			continue
		}
		if quote.CurrentPrice == 0 {
			log.Printf("[seed] skipping %s: no price data", sym.Symbol)
			continue
		}

		financials, err := s.client.GetBasicFinancials(sym.Symbol)
		callsThisMinute++
		if err != nil {
			log.Printf("[seed] skipping %s: financials error: %v", sym.Symbol, err)
			continue
		}

		listing := &model.Listing{
			Ticker:      sym.Symbol,
			Name:        profile.Name,
			ExchangeMIC: profile.Exchange,
			LastRefresh: time.Now(),
			Price:       quote.CurrentPrice,
			Ask:         quote.High,
		}
		if err := s.listingRepo.Upsert(listing); err != nil {
			log.Printf("[seed] skipping %s: listing upsert error: %v", sym.Symbol, err)
			continue
		}

		stock := &model.Stock{
			ListingID:         listing.ListingID,
			OutstandingShares: profile.ShareOutstanding,
			DividendYield:     financials.Metric.DividendYieldIndicatedAnnual,
		}
		if err := s.stockRepo.Upsert(stock); err != nil {
			log.Printf("[seed] skipping %s: stock upsert error: %v", sym.Symbol, err)
			continue
		}

		count++
		log.Printf("[seed] [%d/%d] seeded %s", count, limit, sym.Symbol)
	}

	log.Printf("[seed] done. seeded %d stocks.", count)
	return nil
}
func (s *StockService) SeedOptions(ctx context.Context, limit int) error {
	listings, err := s.listingRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load listings: %w", err)
	}

	log.Printf("[seed-options] seeding options for up to %d stocks from %d listings", limit, len(listings))

	callsThisMinute := 0
	minuteStart := time.Now()
	stockCount := 0
	optionCount := 0

	for _, listing := range listings {
		if stockCount >= limit {
			break
		}

		if callsThisMinute+1 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				log.Printf("[seed-options] rate limit reached, waiting %s...", wait.Round(time.Second))
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		chain, err := s.client.GetOptionChain(listing.Ticker)
		callsThisMinute++
		if err != nil {
			log.Printf("[seed-options] skipping %s: option chain error: %v", listing.Ticker, err)
			continue
		}
		if len(chain.Data) == 0 {
			log.Printf("[seed-options] skipping %s: no option chain data", listing.Ticker)
			continue
		}

		expiration := chain.Data[0]

		seeded := s.seedContractsFromExpiration(listing, expiration)
		optionCount += seeded
		stockCount++
		log.Printf("[seed-options] [%d/%d] seeded %d options for %s (exp %s)",
			stockCount, limit, seeded, listing.Ticker, expiration.ExpirationDate)
	}

	log.Printf("[seed-options] done. seeded %d options across %d stocks.", optionCount, stockCount)
	return nil
}

const maxOptionsPerType = 2

func (s *StockService) seedContractsFromExpiration(stockListing model.Listing, expiration client.OptionChainExpiration) int {
	settlementDate, err := time.Parse("2006-01-02", expiration.ExpirationDate)
	if err != nil {
		log.Printf("[seed-options] skipping expiration %s: invalid date: %v", expiration.ExpirationDate, err)
		return 0
	}

	count := 0

	for i, contract := range expiration.Options.Call {
		if i >= maxOptionsPerType {
			break
		}
		if contract.Strike == 0 || contract.LastPrice == 0 {
			continue
		}
		if s.seedSingleOption(stockListing, contract, model.OptionTypeCall, settlementDate) {
			count++
		}
	}

	for i, contract := range expiration.Options.Put {
		if i >= maxOptionsPerType {
			break
		}
		if contract.Strike == 0 || contract.LastPrice == 0 {
			continue
		}
		if s.seedSingleOption(stockListing, contract, model.OptionTypePut, settlementDate) {
			count++
		}
	}

	return count
}

func (s *StockService) seedSingleOption(
	stockListing model.Listing,
	contract client.OptionContract,
	optType model.OptionType,
	settlementDate time.Time,
) bool {
	ticker := contract.ContractName
	if ticker == "" {
		ticker = fmt.Sprintf("%s:%s:%.2f", stockListing.Ticker, optType, contract.Strike)
	}

	listing := &model.Listing{
		Ticker:      ticker,
		Name:        fmt.Sprintf("%s %s %.2f %s", stockListing.Ticker, optType, contract.Strike, settlementDate.Format("2006-01-02")),
		ExchangeMIC: stockListing.ExchangeMIC,
		LastRefresh: time.Now(),
		Price:       contract.LastPrice,
		Ask:         contract.Ask,
	}
	if err := s.listingRepo.Upsert(listing); err != nil {
		log.Printf("[seed-options] failed to upsert listing for %s: %v", ticker, err)
		return false
	}

	contractSize := 100
	if contract.ContractSize != "" {
		if parsed, parseErr := strconv.Atoi(contract.ContractSize); parseErr == nil {
			contractSize = parsed
		}
	}

	option := &model.Option{
		ListingID:         listing.ListingID,
		OptionType:        optType,
		StrikePrice:       contract.Strike,
		ContractSize:      contractSize,
		SettlementDate:    settlementDate,
		ImpliedVolatility: contract.ImpliedVolatility,
		OpenInterest:      contract.OpenInterest,
	}
	if err := s.optionRepo.Upsert(option); err != nil {
		log.Printf("[seed-options] failed to upsert option for %s: %v", ticker, err)
		return false
	}

	return true
}

func (s *StockService) RefreshPrices(ctx context.Context) error {
	listings, err := s.listingRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load listings: %w", err)
	}

	log.Printf("[refresh] refreshing prices for %d listings", len(listings))

	callsThisMinute := 0
	minuteStart := time.Now()

	for _, listing := range listings {
		if callsThisMinute+1 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				log.Printf("[refresh] rate limit reached, waiting %s...", wait.Round(time.Second))
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		quote, err := s.client.GetQuote(listing.Ticker)
		callsThisMinute++
		if err != nil {
			log.Printf("[refresh] skipping %s: %v", listing.Ticker, err)
			continue
		}
		if quote.CurrentPrice == 0 {
			log.Printf("[refresh] skipping %s: no price data", listing.Ticker)
			continue
		}

		if err := s.listingRepo.UpdatePriceAndAsk(&listing, quote.CurrentPrice, quote.High); err != nil {
			log.Printf("[refresh] failed to update %s: %v", listing.Ticker, err)
			continue
		}

		log.Printf("[refresh] updated %s → price=%.4f ask=%.4f", listing.Ticker, quote.CurrentPrice, quote.High)
	}

	log.Printf("[refresh] done")
	return nil
}
func (s *StockService) RefreshOptions(ctx context.Context) error {
	listings, err := s.listingRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load listings: %w", err)
	}

	log.Printf("[refresh-options] refreshing options data")

	callsThisMinute := 0
	minuteStart := time.Now()

	for _, listing := range listings {
		if strings.Contains(listing.Ticker, ":") {
			continue
		}

		if callsThisMinute+1 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				log.Printf("[refresh-options] rate limit reached, waiting %s...", wait.Round(time.Second))
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		chain, err := s.client.GetOptionChain(listing.Ticker)
		callsThisMinute++
		if err != nil {
			log.Printf("[refresh-options] skipping %s: %v", listing.Ticker, err)
			continue
		}
		if len(chain.Data) == 0 {
			continue
		}

		expiration := chain.Data[0]
		s.seedContractsFromExpiration(listing, expiration)
		log.Printf("[refresh-options] refreshed options for %s", listing.Ticker)
	}

	log.Printf("[refresh-options] done")
	return nil
}

func waitForNextCall(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
