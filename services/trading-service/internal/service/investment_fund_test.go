package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/stretchr/testify/require"
)

// ── Fake Fund Repo ────────────────────────────────────────────────

type fakeFundRepo struct {
	findByIDResult   *model.InvestmentFund
	findByIDErr      error
	findByNameResult *model.InvestmentFund
	findByNameErr    error
	createErr        error
	created          *model.InvestmentFund
}

func (f *fakeFundRepo) FindByName(ctx context.Context, name string) (*model.InvestmentFund, error) {
	return f.findByNameResult, f.findByNameErr
}

func (f *fakeFundRepo) FindByID(ctx context.Context, id uint) (*model.InvestmentFund, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeFundRepo) FindByAccountNumber(ctx context.Context, accountNumber string) (*model.InvestmentFund, error) {
	return nil, nil
}

func (f *fakeFundRepo) Create(ctx context.Context, fund *model.InvestmentFund) error {
	if f.createErr != nil {
		return f.createErr
	}
	fund.FundID = 1
	f.created = fund
	return nil
}

// ── Fake Fund Banking Client ──────────────────────────────────────

type fakeFundBankingClient struct {
	createdAccountNumber string
	createFundAccountErr error
}

func (f *fakeFundBankingClient) GetAccountByNumber(_ context.Context, _ string) (*pb.GetAccountByNumberResponse, error) {
	return nil, nil
}
func (f *fakeFundBankingClient) HasActiveLoan(_ context.Context, _ uint64) (*pb.HasActiveLoanResponse, error) {
	return nil, nil
}
func (f *fakeFundBankingClient) CreatePaymentWithoutVerification(_ context.Context, _ *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	return nil, nil
}
func (f *fakeFundBankingClient) GetAccountsByClientID(_ context.Context, _ uint64) (*pb.GetAccountsByClientIDResponse, error) {
	return nil, nil
}
func (f *fakeFundBankingClient) ConvertCurrency(_ context.Context, amount float64, _, _ string) (float64, error) {
	return amount, nil
}
func (f *fakeFundBankingClient) ExecuteTradeSettlement(_ context.Context, _, _ string, _ pb.TradeSettlementDirection, _ float64) (*pb.ExecuteTradeSettlementResponse, error) {
	return nil, nil
}
func (f *fakeFundBankingClient) GetAccountCurrency(_ context.Context, _ string) (string, error) {
	return "RSD", nil
}
func (f *fakeFundBankingClient) CreateFundAccount(_ context.Context, _ string, _ uint64) (string, error) {
	return f.createdAccountNumber, f.createFundAccountErr
}

// ── Helpers ───────────────────────────────────────────────────────

func fundSupervisorCtx() context.Context {
	employeeID := uint(25)
	return auth.SetAuthOnContext(context.Background(), &auth.AuthContext{
		IdentityID:   200,
		IdentityType: auth.IdentityEmployee,
		EmployeeID:   &employeeID,
	})
}

func fundClientCtx() context.Context {
	clientID := uint(99)
	return auth.SetAuthOnContext(context.Background(), &auth.AuthContext{
		IdentityID:   1,
		IdentityType: auth.IdentityClient,
		ClientID:     &clientID,
	})
}

func validFundRequest() dto.CreateFundRequest {
	return dto.CreateFundRequest{
		Name:                "Alpha Growth Fund",
		Description:         "Fund focused on the IT sector.",
		MinimumContribution: 1000.00,
	}
}

func newTestFundService(fundRepo *fakeFundRepo, bankingClient *fakeFundBankingClient) *InvestmentFundService {
	return NewInvestmentFundService(fundRepo, bankingClient)
}

// ── Tests ─────────────────────────────────────────────────────────

func TestCreateFund_Success(t *testing.T) {
	fundRepo := &fakeFundRepo{}
	bankingClient := &fakeFundBankingClient{createdAccountNumber: "444000112345678901"}
	svc := newTestFundService(fundRepo, bankingClient)

	resp, err := svc.CreateFund(fundSupervisorCtx(), validFundRequest())

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "Alpha Growth Fund", resp.Name)
	require.Equal(t, "444000112345678901", resp.AccountNumber)
	require.Equal(t, uint(25), resp.ManagerID)
	require.Equal(t, 1000.00, resp.MinimumContribution)
	require.Equal(t, 0.0, resp.LiquidAssets)
	require.WithinDuration(t, time.Now(), resp.CreatedAt, 5*time.Second)
}

func TestCreateFund_Unauthenticated(t *testing.T) {
	svc := newTestFundService(&fakeFundRepo{}, &fakeFundBankingClient{})

	_, err := svc.CreateFund(context.Background(), validFundRequest())

	require.Error(t, err)
	require.Contains(t, err.Error(), "not authenticated")
}

func TestCreateFund_NotEmployee(t *testing.T) {
	svc := newTestFundService(&fakeFundRepo{}, &fakeFundBankingClient{})

	_, err := svc.CreateFund(fundClientCtx(), validFundRequest())

	require.Error(t, err)
	require.Contains(t, err.Error(), "only employees")
}

func TestCreateFund_DuplicateName(t *testing.T) {
	fundRepo := &fakeFundRepo{
		findByNameResult: &model.InvestmentFund{Name: "Alpha Growth Fund"},
	}
	svc := newTestFundService(fundRepo, &fakeFundBankingClient{})

	_, err := svc.CreateFund(fundSupervisorCtx(), validFundRequest())

	require.Error(t, err)
	require.Contains(t, err.Error(), "already taken")
}

func TestCreateFund_FindByNameRepoError(t *testing.T) {
	fundRepo := &fakeFundRepo{
		findByNameErr: errors.New("db error"),
	}
	svc := newTestFundService(fundRepo, &fakeFundBankingClient{})

	_, err := svc.CreateFund(fundSupervisorCtx(), validFundRequest())

	require.Error(t, err)
}

func TestCreateFund_BankingClientError(t *testing.T) {
	fundRepo := &fakeFundRepo{}
	bankingClient := &fakeFundBankingClient{
		createFundAccountErr: fmt.Errorf("banking service unavailable"),
	}
	svc := newTestFundService(fundRepo, bankingClient)

	_, err := svc.CreateFund(fundSupervisorCtx(), validFundRequest())

	require.Error(t, err)
}

func TestCreateFund_RepoCreateError(t *testing.T) {
	fundRepo := &fakeFundRepo{
		createErr: errors.New("db error"),
	}
	bankingClient := &fakeFundBankingClient{createdAccountNumber: "444000112345678901"}
	svc := newTestFundService(fundRepo, bankingClient)

	_, err := svc.CreateFund(fundSupervisorCtx(), validFundRequest())

	require.Error(t, err)
}
