package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	pkgerrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
)

type OTCHandler struct {
	service *service.OTCService
}

func NewOTCHandler(service *service.OTCService) *OTCHandler {
	return &OTCHandler{service: service}
}

// PublishAsset godoc
// @Summary Publish stocks for OTC trading
// @Description Sets the number of stocks the caller makes publicly visible on the OTC portal.
// The amount replaces the current public count. Must be non-negative and cannot exceed
// owned minus reserved stocks. Accessible by the owning client or actuary only.
// @Tags otc
// @Produce json
// @Param id path int true "Asset ownership ID"
// @Param amount path number true "Number of stocks to make public"
// @Success 200 {object} dto.OTCInfoResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Security BearerAuth
// @Router /api/client/{clientId}/assets/{id}/publish/{amount} [patch]
func (h *OTCHandler) PublishAsset(c *gin.Context) {
	ownershipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid asset ownership id"))
		return
	}

	amount, err := strconv.ParseFloat(c.Param("amount"), 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid amount"))
		return
	}

	authCtx := auth.GetAuth(c)
	if authCtx == nil {
		c.Error(pkgerrors.UnauthorizedErr("not authenticated"))
		return
	}

	ownerType := model.OwnerTypeClient
	if authCtx.IdentityType == auth.IdentityEmployee {
		ownerType = model.OwnerTypeActuary
	}

	saved, svcErr := h.service.PublishAsset(c.Request.Context(), uint(ownershipID), authCtx.IdentityID, ownerType, amount)
	if svcErr != nil {
		c.Error(svcErr)
		return
	}

	c.JSON(http.StatusOK, dto.ToOTCInfoResponse(saved))
}

// GetOTCInfo godoc
// @Summary Get OTC info for an owned asset
// @Description Returns the OTC info (public count, reserved count) for a specific asset ownership.
// Only the owning client or actuary can access this.
// @Tags otc
// @Produce json
// @Param id path int true "Asset ownership ID"
// @Success 200 {object} dto.OTCInfoResponse
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Security BearerAuth
// @Router /api/client/{clientId}/assets/{id}/otc [get]
func (h *OTCHandler) GetOTCInfo(c *gin.Context) {
	ownershipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid asset ownership id"))
		return
	}

	authCtx := auth.GetAuth(c)
	if authCtx == nil {
		c.Error(pkgerrors.UnauthorizedErr("not authenticated"))
		return
	}

	ownerType := model.OwnerTypeClient
	if authCtx.IdentityType == auth.IdentityEmployee {
		ownerType = model.OwnerTypeActuary
	}

	info, svcErr := h.service.GetOTCInfo(c.Request.Context(), uint(ownershipID), authCtx.IdentityID, ownerType)
	if svcErr != nil {
		c.Error(svcErr)
		return
	}

	c.JSON(http.StatusOK, dto.ToOTCInfoResponse(info))
}

// GetPublicOTCAssets godoc
// @Summary List all publicly available OTC assets
// @Description Returns a paginated list of stocks that have been marked public by their owners on the OTC portal.
// Each entry includes name, ticker, current price, currency, available amount, last updated timestamp, and owner name.
// @Tags otc
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 10)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Security BearerAuth
// @Router /api/otc/public [get]
func (h *OTCHandler) GetPublicOTCAssets(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	assets, total, err := h.service.GetPublicOTCAssets(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      assets,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
