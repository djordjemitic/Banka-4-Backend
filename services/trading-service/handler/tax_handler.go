package handler

import (
	"net/http"
	"strings"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
	"github.com/gin-gonic/gin"
)

type TaxHandler struct {
	taxService *service.TaxService
	userClient client.UserServiceClient
}

func NewTaxHandler(taxService *service.TaxService, userClient client.UserServiceClient) *TaxHandler {
	return &TaxHandler{taxService: taxService, userClient: userClient}
}

type UserTaxEntry struct {
	ID         uint64  `json:"id"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Email      string  `json:"email"`
	UserType   string  `json:"userType"`
	TaxOwedRSD float64 `json:"taxOwedRsd"`
}

func (h *TaxHandler) ListTaxUsers(c *gin.Context) {
	ctx := c.Request.Context()

	userTypeFilter := c.Query("userType")
	firstNameFilter := c.Query("first_name")
	lastNameFilter := c.Query("last_name")
	page := int32(1)
	pageSize := int32(100)

	var entries []UserTaxEntry

	if userTypeFilter == "" || userTypeFilter == "client" {
		clientsResp, err := h.userClient.GetAllClients(ctx, page, pageSize, firstNameFilter, lastNameFilter)
		if err != nil {
			_ = c.Error(errors.InternalErr(err))
			return
		}
		for _, cl := range clientsResp.Clients {
			tax, err := h.taxService.GetClientTotalTax(ctx, cl.Id)
			if err != nil {
				_ = c.Error(err)
				return
			}
			entries = append(entries, UserTaxEntry{
				ID:         cl.Id,
				FirstName:  cl.FirstName,
				LastName:   cl.LastName,
				Email:      cl.Email,
				UserType:   "client",
				TaxOwedRSD: tax,
			})
		}
	}

	if userTypeFilter == "" || userTypeFilter == "actuary" {
		actuariesResp, err := h.userClient.GetAllActuaries(ctx, page, pageSize, firstNameFilter, lastNameFilter)
		if err != nil {
			_ = c.Error(errors.InternalErr(err))
			return
		}
		for _, act := range actuariesResp.Actuaries {
			tax, err := h.taxService.GetEmployeeTotalTax(ctx, uint(act.Id))
			if err != nil {
				_ = c.Error(err)
				return
			}
			entries = append(entries, UserTaxEntry{
				ID:         act.Id,
				FirstName:  act.FirstName,
				LastName:   act.LastName,
				Email:      act.Email,
				UserType:   "actuary",
				TaxOwedRSD: tax,
			})
		}
	}

	if entries == nil {
		entries = []UserTaxEntry{}
	}

	c.JSON(http.StatusOK, entries)
}

func (h *TaxHandler) CollectTaxes(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.taxService.CollectTaxes(ctx); err != nil {
		_ = c.Error(errors.InternalErr(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tax collection completed"})
}

func nameMatches(firstName, lastName, query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(firstName), q) ||
		strings.Contains(strings.ToLower(lastName), q)
}
