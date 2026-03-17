package repository

import (
	"context"
)

type TransferData struct {
	SourceAccountNum string
	DestAccountNum   string
	Amount           float64
	Description      string
}

type TransferHistory struct {
	TransactionID    *uint
	SourceAccountNum string
	DestAccountNum   string
	Amount           float64
	Description      string
	Status           string
	CreatedAt        string
}

type TransferRepository interface {
	// CreateTransfer zapisuje transfer u transaction tabelu (TransactionID će biti popunjen kasnije)
	CreateTransfer(ctx context.Context, sourceAccount, destAccount string, amount float64, description string) error

	// GetTransferHistory vraća transfere za račun sa filteriranjem
	GetTransferHistory(ctx context.Context, accountNum string, status string, startDate, endDate string, page, pageSize int) ([]TransferHistory, int64, error)
}
