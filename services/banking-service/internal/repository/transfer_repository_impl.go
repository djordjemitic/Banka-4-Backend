package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &transferRepository{db: db}
}

// CreateTransfer kreira zapis transfera
// TODO: Kada bude Transaction tabela, koristiti je za zapis
func (r *transferRepository) CreateTransfer(ctx context.Context, sourceAccount, destAccount string, amount float64, description string) error {
	// TODO: Implementirati čuvanje u Transaction tabelu
	// Za sada samo logovanje
	return nil
}

// GetTransferHistory vraća istoriju transfera za račun
// TODO: Kada bude Transaction tabela, čitati odatle sa filteriranjem po type='TRANSFER'
func (r *transferRepository) GetTransferHistory(ctx context.Context, accountNum string, status string, startDate, endDate string, page, pageSize int) ([]TransferHistory, int64, error) {
	var transfers []TransferHistory
	var total int64

	query := r.db.WithContext(ctx)

	// TODO: Filtriranje na osnovu Transaction tabele
	// WHERE type = 'TRANSFER' AND (from_account = ? OR to_account = ?)

	if status != "" {
		// TODO: status filtriranje iz Transaction tabele
	}

	if startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}

	if endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	if err := query.Model(&TransferHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&transfers).Error; err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}
