package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/stretchr/testify/require"
)

// --- Fake OTC repo ---

type fakeOTCRepo struct {
	existing    *model.OTCInfo
	findErr     error
	upsertErr   error
	saved       *model.OTCInfo
	allPublic   []model.OTCInfo
	publicTotal int64
	findAllErr  error
}

func (r *fakeOTCRepo) FindByOwnershipID(_ context.Context, _ uint) (*model.OTCInfo, error) {
	return r.existing, r.findErr
}

func (r *fakeOTCRepo) FindAllPublic(_ context.Context, _, _ int) ([]model.OTCInfo, int64, error) {
	return r.allPublic, r.publicTotal, r.findAllErr
}

func (r *fakeOTCRepo) Upsert(_ context.Context, info *model.OTCInfo) (*model.OTCInfo, error) {
	if r.upsertErr != nil {
		return nil, r.upsertErr
	}
	r.saved = info
	return info, nil
}

// --- Fake user client ---

type fakeOTCUserClient struct{}

func (f *fakeOTCUserClient) GetClientById(_ context.Context, id uint64) (*pb.GetClientByIdResponse, error) {
	return &pb.GetClientByIdResponse{Id: id, FullName: "Test Client"}, nil
}

func (f *fakeOTCUserClient) GetEmployeeById(_ context.Context, id uint64) (*pb.GetEmployeeByIdResponse, error) {
	return &pb.GetEmployeeByIdResponse{Id: id, FullName: "Test Employee"}, nil
}

func (f *fakeOTCUserClient) GetAllClients(_ context.Context, _, _ int32, _, _ string) (*pb.GetAllClientsResponse, error) {
	return &pb.GetAllClientsResponse{}, nil
}

func (f *fakeOTCUserClient) GetAllActuaries(_ context.Context, _, _ int32, _, _ string) (*pb.GetAllActuariesResponse, error) {
	return &pb.GetAllActuariesResponse{}, nil
}

// --- Helpers ---

func makeOwnershipForOTC(id, identityID, assetID uint, ownerType model.OwnerType, amount float64) *model.AssetOwnership {
	return &model.AssetOwnership{
		AssetOwnershipID: id,
		IdentityID:       identityID,
		OwnerType:        ownerType,
		AssetID:          assetID,
		Amount:           amount,
		UpdatedAt:        time.Now(),
	}
}

func newTestOTCService(otcRepo *fakeOTCRepo, ownershipRepo *fakeAssetOwnershipRepo, listingRepo *fakeListingRepo) *OTCService {
	return NewOTCService(otcRepo, ownershipRepo, listingRepo, &fakeOTCUserClient{})
}

// --- PublishAsset tests ---

func TestOTCService_PublishAsset(t *testing.T) {
	cases := []struct {
		name          string
		ownershipRepo *fakeAssetOwnershipRepo
		otcRepo       *fakeOTCRepo
		ownershipID   uint
		identityID    uint
		ownerType     model.OwnerType
		amount        float64
		wantErr       bool
		checkErr      func(t *testing.T, err error)
		check         func(t *testing.T, result *model.OTCInfo)
	}{
		{
			name: "happy path — no existing OTCInfo",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo:     &fakeOTCRepo{existing: nil},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeClient, amount: 5,
			check: func(t *testing.T, result *model.OTCInfo) {
				require.Equal(t, float64(5), result.PublicCount)
				require.Equal(t, float64(0), result.ReservedCount)
				require.Equal(t, uint(1), result.AssetOwnershipID)
				require.Equal(t, uint(5), result.AssetID)
			},
		},
		{
			name: "happy path — existing OTCInfo keeps ReservedCount",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo: &fakeOTCRepo{
				existing: &model.OTCInfo{OTCInfoID: 99, AssetOwnershipID: 1, PublicCount: 3, ReservedCount: 2},
			},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeClient, amount: 10,
			check: func(t *testing.T, result *model.OTCInfo) {
				require.Equal(t, float64(10), result.PublicCount)
				require.Equal(t, float64(2), result.ReservedCount)
				require.Equal(t, uint(99), result.OTCInfoID)
			},
		},
		{
			name: "ownership not found",
			ownershipRepo: &fakeAssetOwnershipRepo{byID: nil},
			otcRepo:       &fakeOTCRepo{},
			ownershipID:   99, identityID: 10, ownerType: model.OwnerTypeClient, amount: 5,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "identity mismatch",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo:     &fakeOTCRepo{},
			ownershipID: 1, identityID: 99, ownerType: model.OwnerTypeClient, amount: 5,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "do not own")
			},
		},
		{
			name: "owner type mismatch",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo:     &fakeOTCRepo{},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeActuary, amount: 5,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "do not own")
			},
		},
		{
			name: "amount < 0",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo:     &fakeOTCRepo{},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeClient, amount: -1,
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "non-negative")
			},
		},
		{
			name: "amount exceeds available",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 10),
			},
			otcRepo: &fakeOTCRepo{
				existing: &model.OTCInfo{ReservedCount: 3},
			},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeClient, amount: 8,
			// available = 10 - 3 = 7, asking for 8
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				require.Contains(t, err.Error(), "exceeds available")
			},
		},
		{
			name: "upsert error",
			ownershipRepo: &fakeAssetOwnershipRepo{
				byID: makeOwnershipForOTC(1, 10, 5, model.OwnerTypeClient, 20),
			},
			otcRepo:     &fakeOTCRepo{upsertErr: errors.New("db error")},
			ownershipID: 1, identityID: 10, ownerType: model.OwnerTypeClient, amount: 5,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestOTCService(tc.otcRepo, tc.ownershipRepo, &fakeListingRepo{})
			result, err := svc.PublishAsset(context.Background(), tc.ownershipID, tc.identityID, tc.ownerType, tc.amount)
			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
			tc.check(t, result)
		})
	}
}

// --- GetPublicOTCAssets tests ---

func TestOTCService_GetPublicOTCAssets(t *testing.T) {
	asset := model.Asset{AssetID: 5, Ticker: "AAPL", Name: "Apple Inc."}
	exchange := &model.Exchange{MicCode: "XNYS", Currency: "USD"}
	listing := model.Listing{
		ListingID:   1,
		AssetID:     5,
		Price:       150.0,
		Exchange:    exchange,
	}

	ownership := model.AssetOwnership{
		AssetOwnershipID: 1,
		IdentityID:       10,
		OwnerType:        model.OwnerTypeClient,
		AssetID:          5,
		Asset:            asset,
	}

	otcInfo := model.OTCInfo{
		OTCInfoID:        7,
		AssetOwnershipID: 1,
		AssetID:          5,
		PublicCount:      8,
		ReservedCount:    2,
		AssetOwnership:   ownership,
		UpdatedAt:        time.Now(),
	}

	cases := []struct {
		name        string
		otcRepo     *fakeOTCRepo
		listingRepo *fakeListingRepo
		wantErr     bool
		check       func(t *testing.T, results []interface{}, total int64)
	}{
		{
			name: "happy path with listings",
			otcRepo: &fakeOTCRepo{
				allPublic:   []model.OTCInfo{otcInfo},
				publicTotal: 1,
			},
			listingRepo: &fakeListingRepo{byAssetIDs: []model.Listing{listing}},
			check: func(t *testing.T, _ []interface{}, total int64) {
				require.Equal(t, int64(1), total)
			},
		},
		{
			name: "empty result",
			otcRepo: &fakeOTCRepo{
				allPublic:   []model.OTCInfo{},
				publicTotal: 0,
			},
			listingRepo: &fakeListingRepo{},
			check: func(t *testing.T, _ []interface{}, total int64) {
				require.Equal(t, int64(0), total)
			},
		},
		{
			name:        "repo error",
			otcRepo:     &fakeOTCRepo{findAllErr: errors.New("db error")},
			listingRepo: &fakeListingRepo{},
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newTestOTCService(tc.otcRepo, &fakeAssetOwnershipRepo{}, tc.listingRepo)
			results, total, err := svc.GetPublicOTCAssets(context.Background(), 1, 10)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			// convert to []interface{} for the check func signature
			iface := make([]interface{}, len(results))
			for i, r := range results {
				iface[i] = r
			}
			tc.check(t, iface, total)
		})
	}
}

func TestOTCService_GetPublicOTCAssets_FieldMapping(t *testing.T) {
	asset := model.Asset{AssetID: 5, Ticker: "AAPL", Name: "Apple Inc."}
	exchange := &model.Exchange{MicCode: "XNYS", Currency: "USD"}
	listing := model.Listing{ListingID: 1, AssetID: 5, Price: 150.0, Exchange: exchange}
	ownership := model.AssetOwnership{
		AssetOwnershipID: 1,
		IdentityID:       10,
		OwnerType:        model.OwnerTypeClient,
		AssetID:          5,
		Asset:            asset,
	}
	otcInfo := model.OTCInfo{
		OTCInfoID:        7,
		AssetOwnershipID: 1,
		AssetID:          5,
		PublicCount:      8,
		ReservedCount:    2,
		AssetOwnership:   ownership,
	}

	svc := newTestOTCService(
		&fakeOTCRepo{allPublic: []model.OTCInfo{otcInfo}, publicTotal: 1},
		&fakeAssetOwnershipRepo{},
		&fakeListingRepo{byAssetIDs: []model.Listing{listing}},
	)

	results, total, err := svc.GetPublicOTCAssets(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, results, 1)

	r := results[0]
	require.Equal(t, "AAPL", r.Ticker)
	require.Equal(t, "Apple Inc.", r.Name)
	require.Equal(t, float64(150.0), r.Price)
	require.Equal(t, "USD", r.Currency)
	require.Equal(t, float64(6), r.AvailableAmount) // 8 - 2
	require.Equal(t, "Test Client", r.OwnerName)
}
