package service

import (
	"banking-service/internal/dto"
	"banking-service/internal/repository"
	"context"
	"testing"
)

type mockTransferRepo struct {
	lastSourceAccount string
	lastDestAccount   string
	lastAmount        float64
	lastDescription   string
}

func (m *mockTransferRepo) CreateTransfer(ctx context.Context, sourceAccount, destAccount string, amount float64, description string) error {
	m.lastSourceAccount = sourceAccount
	m.lastDestAccount = destAccount
	m.lastAmount = amount
	m.lastDescription = description
	return nil
}

func (m *mockTransferRepo) GetTransferHistory(ctx context.Context, accountNum string, status string, startDate, endDate string, page, pageSize int) ([]repository.TransferHistory, int64, error) {
	return []repository.TransferHistory{}, 0, nil
}

func TestExecuteTransfer_SameAccount(t *testing.T) {
	repo := &mockTransferRepo{}
	service := NewTransferService(repo)

	req := dto.CreateTransferRequest{
		SourceAccountNum: "444000112345678911",
		DestAccountNum:   "444000112345678911",
		Amount:           1000,
		Description:      "Test",
	}

	_, err := service.ExecuteTransfer(context.Background(), req)

	if err == nil {
		t.Error("expected error for same account transfer")
	}
}

func TestExecuteTransfer_ValidTransfer(t *testing.T) {
	repo := &mockTransferRepo{}
	service := NewTransferService(repo)

	req := dto.CreateTransferRequest{
		SourceAccountNum: "444000112345678911",
		DestAccountNum:   "444000112345678913",
		Amount:           1000,
		Description:      "Test transfer",
	}

	transfer, err := service.ExecuteTransfer(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transfer.SourceAccountNum != req.SourceAccountNum {
		t.Errorf("expected source %s, got %s", req.SourceAccountNum, transfer.SourceAccountNum)
	}

	if transfer.Amount != req.Amount {
		t.Errorf("expected amount %.2f, got %.2f", req.Amount, transfer.Amount)
	}

	if repo.lastSourceAccount != req.SourceAccountNum {
		t.Errorf("repository received wrong source account")
	}
}

func TestGetTransferHistory_DefaultPagination(t *testing.T) {
	repo := &mockTransferRepo{}
	service := NewTransferService(repo)

	result, err := service.GetTransferHistory(context.Background(), "444000112345678911", "", "", "", 0, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("expected page size 10, got %d", result.PageSize)
	}
}

func TestGetTransferHistory_MaxPageSize(t *testing.T) {
	repo := &mockTransferRepo{}
	service := NewTransferService(repo)

	result, err := service.GetTransferHistory(context.Background(), "444000112345678911", "", "", "", 1, 200)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PageSize != 10 {
		t.Errorf("expected page size capped to 10, got %d", result.PageSize)
	}
}
