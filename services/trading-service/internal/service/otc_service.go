package service

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
)

type OTCService struct {
	otcRepo       repository.OTCRepository
	ownershipRepo repository.AssetOwnershipRepository
	listingRepo   repository.ListingRepository
	userClient    client.UserServiceClient
}

func NewOTCService(
	otcRepo repository.OTCRepository,
	ownershipRepo repository.AssetOwnershipRepository,
	listingRepo repository.ListingRepository,
	userClient client.UserServiceClient,
) *OTCService {
	return &OTCService{
		otcRepo:       otcRepo,
		ownershipRepo: ownershipRepo,
		listingRepo:   listingRepo,
		userClient:    userClient,
	}
}

func (s *OTCService) PublishAsset(ctx context.Context, ownershipID, identityID uint, ownerType model.OwnerType, amount float64) (*model.OTCInfo, error) {
	ownership, err := s.ownershipRepo.FindByID(ctx, ownershipID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if ownership == nil {
		return nil, errors.NotFoundErr("asset ownership not found")
	}

	if ownership.IdentityID != identityID || ownership.OwnerType != ownerType {
		return nil, errors.ForbiddenErr("you do not own this asset")
	}

	existing, err := s.otcRepo.FindByOwnershipID(ctx, ownershipID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	var reservedCount float64
	if existing != nil {
		reservedCount = existing.ReservedCount
	}

	if amount < 0 {
		return nil, errors.BadRequestErr("amount must be non-negative")
	}
	if amount > ownership.Amount-reservedCount {
		return nil, errors.BadRequestErr("amount exceeds available (owned minus reserved) stocks")
	}

	info := &model.OTCInfo{
		AssetOwnershipID: ownershipID,
		AssetID:          ownership.AssetID,
		PublicCount:      amount,
		ReservedCount:    reservedCount,
	}
	if existing != nil {
		info.OTCInfoID = existing.OTCInfoID
	}

	saved, err := s.otcRepo.Upsert(ctx, info)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	return saved, nil
}

func (s *OTCService) GetOTCInfo(ctx context.Context, ownershipID, identityID uint, ownerType model.OwnerType) (*model.OTCInfo, error) {
	ownership, err := s.ownershipRepo.FindByID(ctx, ownershipID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if ownership == nil {
		return nil, errors.NotFoundErr("asset ownership not found")
	}

	if ownership.IdentityID != identityID || ownership.OwnerType != ownerType {
		return nil, errors.ForbiddenErr("you do not own this asset")
	}

	info, err := s.otcRepo.FindByOwnershipID(ctx, ownershipID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if info == nil {
		return nil, errors.NotFoundErr("no OTC info for this asset ownership")
	}
	return info, nil
}

func (s *OTCService) GetPublicOTCAssets(ctx context.Context, page, pageSize int) ([]dto.OTCAssetResponse, int64, error) {
	infos, total, err := s.otcRepo.FindAllPublic(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.InternalErr(err)
	}

	if len(infos) == 0 {
		return []dto.OTCAssetResponse{}, 0, nil
	}

	assetIDs := make([]uint, 0, len(infos))
	for _, info := range infos {
		assetIDs = append(assetIDs, info.AssetID)
	}

	listings, err := s.listingRepo.FindByAssetIDs(ctx, assetIDs)
	if err != nil {
		return nil, 0, errors.InternalErr(err)
	}

	listingByAssetID := make(map[uint]*model.Listing, len(listings))
	for i := range listings {
		listingByAssetID[listings[i].AssetID] = &listings[i]
	}

	responses := make([]dto.OTCAssetResponse, 0, len(infos))
	for _, info := range infos {
		resp := dto.OTCAssetResponse{
			OTCInfoID:        info.OTCInfoID,
			AssetOwnershipID: info.AssetOwnershipID,
			Name:             info.AssetOwnership.Asset.Name,
			Ticker:           info.AssetOwnership.Asset.Ticker,
			AvailableAmount:  info.PublicCount - info.ReservedCount,
			UpdatedAt:        info.UpdatedAt,
		}

		if listing, ok := listingByAssetID[info.AssetID]; ok {
			resp.Price = listing.Price
			if listing.Exchange != nil {
				resp.Currency = listing.Exchange.Currency
			}
		}

		var ownerName string
		if info.AssetOwnership.OwnerType == model.OwnerTypeClient {
			userResp, userErr := s.userClient.GetClientById(ctx, uint64(info.AssetOwnership.IdentityID))
			if userErr == nil && userResp != nil {
				ownerName = userResp.FullName
			}
		} else {
			userResp, userErr := s.userClient.GetEmployeeById(ctx, uint64(info.AssetOwnership.IdentityID))
			if userErr == nil && userResp != nil {
				ownerName = userResp.FullName
			}
		}
		resp.OwnerName = ownerName

		responses = append(responses, resp)
	}

	return responses, total, nil
}
