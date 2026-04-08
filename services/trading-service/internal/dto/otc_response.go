package dto

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type OTCInfoResponse struct {
	OTCInfoID        uint      `json:"otc_info_id"`
	AssetOwnershipID uint      `json:"asset_ownership_id"`
	PublicCount      float64   `json:"public_count"`
	ReservedCount    float64   `json:"reserved_count"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToOTCInfoResponse(o *model.OTCInfo) OTCInfoResponse {
	return OTCInfoResponse{
		OTCInfoID:        o.OTCInfoID,
		AssetOwnershipID: o.AssetOwnershipID,
		PublicCount:      o.PublicCount,
		ReservedCount:    o.ReservedCount,
		UpdatedAt:        o.UpdatedAt,
	}
}

type OTCAssetResponse struct {
	OTCInfoID        uint      `json:"otc_info_id"`
	AssetOwnershipID uint      `json:"asset_ownership_id"`
	Name             string    `json:"name"`
	Ticker           string    `json:"ticker"`
	Price            float64   `json:"price"`
	Currency         string    `json:"currency"`
	AvailableAmount  float64   `json:"available_amount"`
	UpdatedAt        time.Time `json:"updated_at"`
	OwnerName        string    `json:"owner_name"`
}
