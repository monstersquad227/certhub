package services

import (
	"errors"
	"fmt"
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

// CreateRechargeOrder creates a recharge balance record (simulate payment).
func CreateRechargeOrder(userID uint64, amount float64, paymentMethod string) (*models.BalanceRecord, error) {
	orderNo := fmt.Sprintf("R%d", time.Now().UnixNano())

	record := &models.BalanceRecord{
		UserID:        userID,
		Type:          "recharge",
		Amount:        amount,
		PaymentMethod: paymentMethod,
		OrderNo:       orderNo,
		Description:   "充值",
	}

	if err := database.DB.Create(record).Error; err != nil {
		return nil, err
	}
	// Immediately apply recharge for V1.0.0
	if err := applyBalanceChange(userID, amount); err != nil {
		return nil, err
	}
	return record, nil
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
		CertificateID: &certificateID,
		Description:   desc,
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		// deduct
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


