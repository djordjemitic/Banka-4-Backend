package service

import (
	"banking-service/internal/dto"
	"banking-service/internal/repository"
	"common/pkg/errors"
	"context"
	"time"
)

type TransferService struct {
	repo repository.TransferRepository
}

func NewTransferService(
	repo repository.TransferRepository,
) *TransferService {
	return &TransferService{
		repo: repo,
	}
}

// ExecuteTransfer izvršava transfer između dva računa
func (s *TransferService) ExecuteTransfer(ctx context.Context, req dto.CreateTransferRequest) (*dto.TransferResponse, error) {
	if req.SourceAccountNum == req.DestAccountNum {
		return nil, errors.UnprocessableEntityErr("cannot transfer to yourself")
	}

	// TODO: Validacija da su oba računa od iste klijente
	// TODO: Proveriti dostupnost sredstava na source računu
	// TODO: Implementirati conversion kroz Exchange Office za različite valute
	// TODO: Charge commission (0-1%) za različite valute
	// TODO: Kreirati Transaction record u transaction tabeli

	if err := s.repo.CreateTransfer(ctx, req.SourceAccountNum, req.DestAccountNum, req.Amount, req.Description); err != nil {
		return nil, errors.InternalErr(err)
	}

	return &dto.TransferResponse{
		SourceAccountNum: req.SourceAccountNum,
		DestAccountNum:   req.DestAccountNum,
		Amount:           req.Amount,
		Description:      req.Description,
		Status:           "Pending",
	}, nil
}

// GetTransferHistory vraća historiju transfera za klijentov račun
func (s *TransferService) GetTransferHistory(ctx context.Context, accountNum string, status string, startDate, endDate string, page, pageSize int) (*dto.ListTransfersResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	transfers, total, err := s.repo.GetTransferHistory(ctx, accountNum, status, startDate, endDate, page, pageSize)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	var transferDTOs []dto.TransferResponse
	for _, t := range transfers {
		transferDTOs = append(transferDTOs, dto.TransferResponse{
			TransactionID:    t.TransactionID,
			SourceAccountNum: t.SourceAccountNum,
			DestAccountNum:   t.DestAccountNum,
			Amount:           t.Amount,
			Description:      t.Description,
			Status:           t.Status,
			CreatedAt:        parseTime(t.CreatedAt),
		})
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &dto.ListTransfersResponse{
		Data:       transferDTOs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}
