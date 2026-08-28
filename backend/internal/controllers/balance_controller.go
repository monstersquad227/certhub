package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"certhub-backend/internal/services"
	"certhub-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// GET /api/v1/balance
func GetBalance(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	balance, err := services.GetUserBalance(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取余额失败")
		return
	}
	utils.Success(c, gin.H{"balance": balance})
}

// GET /api/v1/balance/records
func GetBalanceRecords(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	recordType := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := services.GetBalanceRecords(userID, recordType, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取余额记录失败")
		return
	}
	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type rechargeRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required"`
	TxHash        string  `json:"tx_hash"`
}

// POST /api/v1/balance/recharge
func CreateRechargeOrder(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	var req rechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}
	if req.PaymentMethod == "usdt" && strings.TrimSpace(req.TxHash) == "" {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "缺少链上交易哈希")
		return
	}

	record, err := services.CreateRechargeOrder(userID, req.Amount, req.PaymentMethod, req.TxHash)
	if err != nil {
		status, code, msg := mapRechargeError(err)
		utils.Error(c, status, code, msg)
		return
	}
	utils.Success(c, gin.H{
		"order_no":       record.OrderNo,
		"amount":         record.Amount,
		"payment_method": record.PaymentMethod,
		"status":         record.Status,
		"payment_url":    "",
	})
}

func mapRechargeError(err error) (httpStatus int, bizCode int, message string) {
	msg := err.Error()
	switch {
	case errors.Is(err, services.ErrUSDTDisabled),
		errors.Is(err, services.ErrUSDTConfigIncomplete):
		return http.StatusServiceUnavailable, http.StatusServiceUnavailable, msg
	case errors.Is(err, services.ErrTxHashInvalid),
		errors.Is(err, services.ErrTxNotFound),
		errors.Is(err, services.ErrTxFailed),
		errors.Is(err, services.ErrTxNotUSDTTransfer),
		errors.Is(err, services.ErrTxAmountMismatch):
		return http.StatusBadRequest, http.StatusBadRequest, msg
	case strings.Contains(msg, "该交易哈希已用于充值"):
		return http.StatusConflict, http.StatusConflict, msg
	case strings.Contains(msg, "Solana"):
		return http.StatusBadGateway, http.StatusBadGateway, "链上交易校验失败，请稍后重试"
	default:
		return http.StatusInternalServerError, http.StatusInternalServerError, "创建充值订单失败"
	}
}
