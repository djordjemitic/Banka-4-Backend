package dto

type ExerciseOptionRequest struct {
	AccountNumber string `json:"account_number" binding:"required"`
}

type ExerciseOptionResponse struct {
	OptionAssetID           uint    `json:"option_asset_id"`
	StockAssetID            uint    `json:"stock_asset_id"`
	ExercisedContracts      uint    `json:"exercised_contracts"`
	PurchasedShares         float64 `json:"purchased_shares"`
	StrikePrice             float64 `json:"strike_price"`
	TotalCost               float64 `json:"total_cost"`
	RemainingOptionShares   float64 `json:"remaining_option_shares"`
	RemainingContracts      uint    `json:"remaining_contracts"`
	SourceAmount            float64 `json:"source_amount"`
	SourceCurrencyCode      string  `json:"source_currency_code"`
	DestinationAmount       float64 `json:"destination_amount"`
	DestinationCurrencyCode string  `json:"destination_currency_code"`
}
