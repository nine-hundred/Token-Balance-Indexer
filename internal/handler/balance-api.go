package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"onbloc/internal/response"
	balance_api_service "onbloc/internal/service/balance-api-service"
	"strconv"
	"strings"
)

type BalanceAPIHandler struct {
	service *balance_api_service.Service
}

func NewBalanceAPIHandler(service *balance_api_service.Service) *BalanceAPIHandler {
	return &BalanceAPIHandler{
		service: service,
	}
}

func (b *BalanceAPIHandler) GetTokenBalances(c *gin.Context) {
	wildcard := c.Param("wildcard")
	wildcard = strings.TrimPrefix(wildcard, "/")

	if !strings.HasSuffix(wi)

	address := c.Query("address")
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	if address == "" {
		resp, err := b.service.GetAllTokenBalances(c, offset, limit)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
		}
		c.JSON(http.StatusOK, resp)
	} else {
		resp, err := b.service.GetTokenBalances(c, address)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
		}
		c.JSON(http.StatusOK, resp)
	}
}

func (b BalanceAPIHandler) GetTokenPathBalances(c *gin.Context) {
	tokenPath := c.Param("tokenPath")
	tokenPath = strings.ReplaceAll(tokenPath, "%2F", "/")

	address := c.Query("address")

	var resp response.AccountBalancesResponse
	var err error
	if address != "" {
		resp, err = b.service.GetTokenPathBalanceByAddress(c, tokenPath, address)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
		}
		c.JSON(http.StatusOK, resp)
	} else {
		resp, err = b.service.GetAllTokenPathBalances(c, tokenPath)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
		}
		c.JSON(http.StatusOK, resp)
	}
}
