package dto

import "time"

type AssetType string

const (
	AssetTypeStock   AssetType = "STOCK"
	AssetTypeFutures AssetType = "FUTURES"
	AssetTypeOption  AssetType = "OPTION"
	AssetTypeForex   AssetType = "FOREX"
)

type PortfolioAssetResponse struct {
	Type              AssetType `json:"type"`
	Ticker            string    `json:"ticker"`
	Amount            float64   `json:"amount"`
	PricePerUnitRSD   float64   `json:"pricePerUnitRSD"`
	AvgBuyPriceRSD    float64   `json:"avgBuyPrice"`
	LastModified      time.Time `json:"lastModified"`
	Profit            float64   `json:"profit"`
	PublicAmount      float64   `json:"publicAmount"`
}

type PortfolioProfitResponse struct {
	TotalProfitRSD float64 `json:"totalProfitRSD"`
}
