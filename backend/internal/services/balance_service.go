package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"certhub-backend/internal/config"
	"certhub-backend/internal/database"
	"certhub-backend/internal/models"

	"gorm.io/gorm"
)

// GetUserBalance returns current balance for a user.
func GetUserBalance(userID uint64) (float64, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.Balance, nil
}

// CreateRechargeOrder creates a recharge balance record.
// USDT: verify on-chain tx then credit immediately (status=completed).
// Alipay: create pending order only; balance is credited manually by operators.
func CreateRechargeOrder(userID uint64, amount float64, paymentMethod string, txHash string) (*models.BalanceRecord, error) {
	orderNo := fmt.Sprintf("R%d", time.Now().UnixNano())
	status := "completed"
	description := "充值"

	switch paymentMethod {
	case "alipay":
		status = "pending"
		description = "支付宝待人工确认"
	case "usdt":
		txHash = strings.TrimSpace(txHash)
		if txHash == "" {
			return nil, ErrTxHashInvalid
		}

		var existing models.BalanceRecord
		if err := database.DB.Where("order_no = ?", txHash).First(&existing).Error; err == nil {
			return nil, errors.New("该交易哈希已用于充值")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if _, err := VerifySolanaUSDTTransfer(ctx, txHash, amount); err != nil {
			return nil, err
		}
		orderNo = txHash
	}

	record := &models.BalanceRecord{
		UserID:        userID,
		Type:          "recharge",
		Amount:        amount,
		PaymentMethod: paymentMethod,
		OrderNo:       orderNo,
		Status:        status,
		Description:   description,
	}

	if err := database.DB.Create(record).Error; err != nil {
		if isDuplicateOrderNo(err) {
			return nil, errors.New("该交易哈希已用于充值")
		}
		return nil, err
	}

	// Alipay pending orders do not credit balance; operators update DB manually.
	if status == "completed" {
		if err := applyBalanceChange(userID, amount); err != nil {
			return nil, err
		}
	}
	return record, nil
}

func isDuplicateOrderNo(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "uk_order_no")
}

// ConsumeBalance consumes balance for certificate purchase.
func ConsumeBalance(userID uint64, certificateID uint64, amount float64, desc string) (*models.BalanceRecord, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.Balance < amount {
		return nil, errors.New("余额不足")
	}

	orderNo := fmt.Sprintf("C%d", time.Now().UnixNano())
	record := &models.BalanceRecord{
		UserID:        userID,
		Type:          "consume",
		Amount:        -amount,
		OrderNo:       orderNo,
		Status:        "completed",
		CertificateID: &certificateID,
		Description:   desc,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).
			Where("id = ? AND balance >= ?", userID, amount).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

// applyBalanceChange adjusts user balance by delta.
func applyBalanceChange(userID uint64, delta float64) error {
	return database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", delta)).Error
}

// GetBalanceRecords returns paginated balance records for current user.
func GetBalanceRecords(userID uint64, recordType string, page, pageSize int) ([]models.BalanceRecord, int64, error) {
	var (
		list  []models.BalanceRecord
		total int64
	)
	query := database.DB.Model(&models.BalanceRecord{}).Where("user_id = ?", userID)
	if recordType != "" {
		query = query.Where("type = ?", recordType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetCertPrice returns price according to type.
func GetCertPrice(isWildcard bool) float64 {
	if isWildcard {
		return config.C.Cert.PriceWildcard
	}
	return config.C.Cert.PriceSingle
}
