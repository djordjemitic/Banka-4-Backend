//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- Seed helpers ---

func seedAssetOwnership(t *testing.T, db *gorm.DB, identityID uint, ownerType model.OwnerType, assetID uint, amount float64) *model.AssetOwnership {
	t.Helper()
	o := &model.AssetOwnership{
		IdentityID: identityID,
		OwnerType:  ownerType,
		AssetID:    assetID,
		Amount:     amount,
		UpdatedAt:  time.Now(),
	}
	require.NoError(t, db.Create(o).Error)
	return o
}

func seedOTCInfo(t *testing.T, db *gorm.DB, ownershipID, assetID uint, public, reserved float64) *model.OTCInfo {
	t.Helper()
	info := &model.OTCInfo{
		AssetOwnershipID: ownershipID,
		AssetID:          assetID,
		PublicCount:      public,
		ReservedCount:    reserved,
		UpdatedAt:        time.Now(),
	}
	require.NoError(t, db.Create(info).Error)
	return info
}

// --- Publish endpoint tests ---

func TestOTCHandler_PublishAsset_ClientSuccess(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 100.0)
	// identityID=50, clientID=50 (RequireClientSelf checks JWT clientID == path clientID)
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 20)

	path := fmt.Sprintf("/api/client/50/assets/%d/publish/5", ownership.AssetOwnershipID)
	rec := performRequest(t, router, http.MethodPatch, path, nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[dto.OTCInfoResponse](t, rec)
	require.Equal(t, float64(5), resp.PublicCount)
	require.Equal(t, float64(0), resp.ReservedCount)
	require.Equal(t, ownership.AssetOwnershipID, resp.AssetOwnershipID)
}

func TestOTCHandler_PublishAsset_ActuarySuccess(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 50.0)
	// Agent has identityID=200 (from authHeaderForAgent which sets IdentityID=200, EmployeeID=20)
	ownership := seedAssetOwnership(t, db, 200, model.OwnerTypeActuary, listing.AssetID, 15)

	path := fmt.Sprintf("/api/actuary/20/assets/%d/publish/3", ownership.AssetOwnershipID)
	rec := performRequest(t, router, http.MethodPatch, path, nil, authHeaderForAgent(t))
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[dto.OTCInfoResponse](t, rec)
	require.Equal(t, float64(3), resp.PublicCount)
}

func TestOTCHandler_PublishAsset_Unauthenticated(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodPatch, "/api/client/50/assets/1/publish/5", nil, "")
	requireStatus(t, rec, http.StatusUnauthorized)
}

func TestOTCHandler_PublishAsset_WrongOwner(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 100.0)
	// ownership belongs to identityID=50
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 20)

	// requester is identityID=99, clientID=99 — does not own this asset
	path := fmt.Sprintf("/api/client/99/assets/%d/publish/5", ownership.AssetOwnershipID)
	rec := performRequest(t, router, http.MethodPatch, path, nil, authHeaderForClient(t, 99, 99))
	requireStatus(t, rec, http.StatusForbidden)
}

func TestOTCHandler_PublishAsset_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodPatch, "/api/client/50/assets/99999/publish/1", nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusNotFound)
}

func TestOTCHandler_PublishAsset_AmountExceedsOwned(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 100.0)
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 10)

	path := fmt.Sprintf("/api/client/50/assets/%d/publish/9999", ownership.AssetOwnershipID)
	rec := performRequest(t, router, http.MethodPatch, path, nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusBadRequest)
}

func TestOTCHandler_PublishAsset_UpdatesExistingOTCInfo(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 100.0)
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 20)
	seedOTCInfo(t, db, ownership.AssetOwnershipID, listing.AssetID, 3, 0)

	path := fmt.Sprintf("/api/client/50/assets/%d/publish/8", ownership.AssetOwnershipID)
	rec := performRequest(t, router, http.MethodPatch, path, nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[dto.OTCInfoResponse](t, rec)
	require.Equal(t, float64(8), resp.PublicCount)
}

// --- GetPublicOTCAssets tests ---

func TestOTCHandler_GetPublicOTCAssets_ReturnsList(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	exchange := seedExchange(t, db, uniqueValue(t, "MIC")[:10])
	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, exchange.MicCode, model.AssetTypeStock, 120.0)
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 10)
	seedOTCInfo(t, db, ownership.AssetOwnershipID, listing.AssetID, 6, 1)

	rec := performRequest(t, router, http.MethodGet, "/api/otc/public?page=1&page_size=10", nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusOK)

	body := decodeResponse[map[string]interface{}](t, rec)
	data, ok := body["data"].([]interface{})
	require.True(t, ok)
	require.GreaterOrEqual(t, len(data), 1)

	entry := data[0].(map[string]interface{})
	require.Equal(t, float64(5), entry["available_amount"]) // 6 - 1
	require.NotEmpty(t, entry["ticker"])
	require.NotEmpty(t, entry["name"])
}

func TestOTCHandler_GetPublicOTCAssets_UnpublishedNotIncluded(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ticker := uniqueValue(t, "TICK")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, 100.0)
	ownership := seedAssetOwnership(t, db, 50, model.OwnerTypeClient, listing.AssetID, 10)
	// public_count = 0 — should not appear
	seedOTCInfo(t, db, ownership.AssetOwnershipID, listing.AssetID, 0, 0)

	rec := performRequest(t, router, http.MethodGet, "/api/otc/public?page=1&page_size=10", nil, authHeaderForClient(t, 50, 50))
	requireStatus(t, rec, http.StatusOK)

	body := decodeResponse[map[string]interface{}](t, rec)
	data := body["data"].([]interface{})
	// the 0-public entry must not be in the results
	for _, item := range data {
		entry := item.(map[string]interface{})
		require.NotEqual(t, ticker, entry["ticker"])
	}
}

func TestOTCHandler_GetPublicOTCAssets_Unauthenticated(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodGet, "/api/otc/public", nil, "")
	requireStatus(t, rec, http.StatusUnauthorized)
}
