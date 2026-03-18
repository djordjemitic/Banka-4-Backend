package handler

import (
	"banking-service/internal/dto"
	"banking-service/internal/service"
	"common/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

func (h *AccountHandler) Create(c *gin.Context) {
	var req dto.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	account, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToAccountResponse(account))
}

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	var req dto.ListAccountsQuery

	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	accounts, total, err := h.service.GetAllAccounts(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      accounts,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}
