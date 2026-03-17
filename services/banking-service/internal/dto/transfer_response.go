package dto

import "time"

type TransferResponse struct {
	TransactionID    *uint     `json:"transaction_id,omitempty"`
	SourceAccountNum string    `json:"source_account_num"`
	DestAccountNum   string    `json:"dest_account_num"`
	Amount           float64   `json:"amount"`
	Description      string    `json:"description,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

type ListTransfersResponse struct {
	Data       []TransferResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}
