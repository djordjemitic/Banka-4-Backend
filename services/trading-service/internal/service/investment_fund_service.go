package service

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	commonErrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type InvestmentFundService struct {
	fundRepo       repository.InvestmentFundRepository
	positionRepo   repository.ClientFundPositionRepository
	investmentRepo repository.ClientFundInvestmentRepository
	bankingClient  client.BankingClient
	now            func() time.Time
}

func NewInvestmentFundService(
	fundRepo repository.InvestmentFundRepository,
	positionRepo repository.ClientFundPositionRepository,
	investmentRepo repository.ClientFundInvestmentRepository,
	bankingClient client.BankingClient,
) *InvestmentFundService {
	return &InvestmentFundService{
		fundRepo:       fundRepo,
		positionRepo:   positionRepo,
		investmentRepo: investmentRepo,
		bankingClient:  bankingClient,
		now:            time.Now,
	}
}

// CreateFund creates a new investment fund. Only supervisors can call this.
// A bank account is automatically created for the fund via the banking service.
func (s *InvestmentFundService) CreateFund(ctx context.Context, req dto.CreateFundRequest) (*dto.CreateFundResponse, error) {
	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, commonErrors.UnauthorizedErr("not authenticated")
	}

	if authCtx.IdentityType != auth.IdentityEmployee {
		return nil, commonErrors.ForbiddenErr("only employees can create investment funds")
	}

	if authCtx.EmployeeID == nil {
		return nil, commonErrors.UnauthorizedErr("employee identity missing")
	}

	managerID := *authCtx.EmployeeID

	existing, err := s.fundRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	if existing != nil {
		return nil, commonErrors.ConflictErr("fund name is already taken")
	}

	accountNumber, err := s.bankingClient.CreateFundAccount(ctx, req.Name, uint64(managerID))
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	fund := &model.InvestmentFund{
		Name:                req.Name,
		Description:         req.Description,
		MinimumContribution: req.MinimumContribution,
		ManagerID:           managerID,
		LiquidAssets:        0,
		AccountNumber:       accountNumber,
		CreatedAt:           s.now(),
	}

	if err := s.fundRepo.Create(ctx, fund); err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	return &dto.CreateFundResponse{
		FundID:              fund.FundID,
		Name:                fund.Name,
		Description:         fund.Description,
		MinimumContribution: fund.MinimumContribution,
		ManagerID:           fund.ManagerID,
		LiquidAssets:        fund.LiquidAssets,
		AccountNumber:       fund.AccountNumber,
		CreatedAt:           fund.CreatedAt,
	}, nil
}

// InvestInFund handles a client or supervisor investing into a fund.
//
// Rules:
//   - The amount must meet the fund's MinimumContribution.
//   - Clients must use one of their own accounts.
//   - Supervisors must use a bank account.
//   - The account is debited via ExecuteTradeSettlement (BUY direction).
//   - A ClientFundInvestment record is always created.
//   - The ClientFundPosition is created if it does not exist, or updated otherwise.
func (s *InvestmentFundService) InvestInFund(ctx context.Context, fundID uint, req dto.InvestInFundRequest) (*dto.InvestInFundResponse, error) {
	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, commonErrors.UnauthorizedErr("not authenticated")
	}

	fund, err := s.fundRepo.FindByID(ctx, fundID)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	if fund == nil {
		return nil, commonErrors.NotFoundErr("fund not found")
	}

	if req.Amount < fund.MinimumContribution {
		return nil, commonErrors.BadRequestErr(
			fmt.Sprintf("amount %.2f is below the fund's minimum contribution of %.2f",
				req.Amount, fund.MinimumContribution),
		)
	}

	callerID, ownerType, err := resolveCallerIdentity(authCtx)
	if err != nil {
		return nil, err
	}

	account, err := s.validateFundAccount(ctx, req.AccountNumber, authCtx)
	if err != nil {
		return nil, err
	}

	currencyCode := account.GetCurrencyCode()

	if authCtx.IdentityType == auth.IdentityEmployee {
		// Supervisor: direktan transfer od bank accounta ka fund accountu
		_, err = s.bankingClient.CreatePaymentWithoutVerification(ctx, &pb.CreatePaymentRequest{
			PayerAccountNumber:     req.AccountNumber,
			RecipientAccountNumber: fund.AccountNumber,
			RecipientName:          fund.Name,
			Amount:                 req.Amount,
			ReferenceNumber:        "",
			PaymentCode:            "289",
			Purpose:                fmt.Sprintf("Investment into fund %s", fund.Name),
		})
	} else {
		// Client: standardni trade settlement
		_, err = s.bankingClient.ExecuteTradeSettlement(
			ctx,
			req.AccountNumber,
			currencyCode,
			pb.TradeSettlementDirection_TRADE_SETTLEMENT_DIRECTION_BUY,
			req.Amount,
		)
	}

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, commonErrors.NotFoundErr(st.Message())
			case codes.FailedPrecondition:
				return nil, commonErrors.BadRequestErr(st.Message())
			}
		}
		return nil, commonErrors.ServiceUnavailableErr(err)
	}

	now := s.now()

	investment := &model.ClientFundInvestment{
		ClientID:      callerID,
		OwnerType:     ownerType,
		FundID:        fundID,
		AccountNumber: req.AccountNumber,
		Amount:        req.Amount,
		CurrencyCode:  currencyCode,
		CreatedAt:     now,
	}
	if err := s.investmentRepo.Create(ctx, investment); err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	position, err := s.positionRepo.FindByClientAndFund(ctx, callerID, ownerType, fundID)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	if position == nil {
		position = &model.ClientFundPosition{
			ClientID:            callerID,
			OwnerType:           ownerType,
			FundID:              fundID,
			TotalInvestedAmount: req.Amount,
			UpdatedAt:           now,
		}
	} else {
		position.TotalInvestedAmount += req.Amount
		position.UpdatedAt = now
	}
	if err := s.positionRepo.Upsert(ctx, position); err != nil {
		return nil, commonErrors.InternalErr(err)
	}

	return &dto.InvestInFundResponse{
		FundID:        fund.FundID,
		FundName:      fund.Name,
		InvestedNow:   req.Amount,
		CurrencyCode:  currencyCode,
		TotalInvested: position.TotalInvestedAmount,
		CreatedAt:     now,
	}, nil
}

func (s *InvestmentFundService) validateFundAccount(ctx context.Context, accountNumber string, authCtx *auth.AuthContext) (*pb.GetAccountByNumberResponse, error) {
	account, err := s.bankingClient.GetAccountByNumber(ctx, accountNumber)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, commonErrors.NotFoundErr("account not found")
		}
		return nil, commonErrors.ServiceUnavailableErr(err)
	}

	switch authCtx.IdentityType {
	case auth.IdentityClient:
		if authCtx.ClientID == nil || uint64(*authCtx.ClientID) != account.GetClientId() {
			return nil, commonErrors.ForbiddenErr("account does not belong to you")
		}
	case auth.IdentityEmployee:
		if account.GetAccountType() != "Bank" {
			return nil, commonErrors.BadRequestErr("supervisors must use a bank account for fund investments")
		}
	}

	return account, nil
}

func resolveCallerIdentity(authCtx *auth.AuthContext) (uint, model.OwnerType, error) {
	switch authCtx.IdentityType {
	case auth.IdentityClient:
		if authCtx.ClientID == nil {
			return 0, "", commonErrors.UnauthorizedErr("not authenticated")
		}
		return *authCtx.ClientID, model.OwnerTypeClient, nil
	case auth.IdentityEmployee:
		if authCtx.EmployeeID == nil {
			return 0, "", commonErrors.UnauthorizedErr("not authenticated")
		}
		return *authCtx.EmployeeID, model.OwnerTypeActuary, nil
	default:
		return 0, "", commonErrors.UnauthorizedErr("unknown identity type")
	}
}
