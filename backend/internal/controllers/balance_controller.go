package controllers

import (
	"net/http"
	"strconv"

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

	record, err := services.CreateRechargeOrder(userID, req.Amount, req.PaymentMethod)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "创建充值订单失败")
		return
	}
	utils.Success(c, gin.H{
		"order_no":       record.OrderNo,
		"amount":         record.Amount,
		"payment_method": record.PaymentMethod,
		"payment_url":    "", // 模拟支付流程
	})
}


