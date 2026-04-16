package services

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"certhub-backend/internal/config"
	"certhub-backend/internal/database"
	"certhub-backend/internal/middleware"
	"certhub-backend/internal/models"

	"gorm.io/gorm"
)

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

// GenerateCode generates a 6-digit verification code.
func GenerateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// ValidateEmailFormat validates email format.
func ValidateEmailFormat(email string) bool {
	return emailRegex.MatchString(email)
}

// RequestVerificationCode handles rate limiting and sends a verification code.
func RequestVerificationCode(email string) error {
	now := time.Now()
	interval := config.C.Security.VerificationCodeIntervalSecond
	
	// Check last code sending time (check the most recent code regardless of status)
	var last models.VerificationCode
	err := database.DB.Where("email = ?", email).
		Order("created_at desc").
		First(&last).Error
	
	if err == nil {
		// Found a recent code, check if it was sent within the interval
		timeSinceLastCode := now.Sub(last.CreatedAt)
		if timeSinceLastCode < time.Duration(interval)*time.Second {
			remainingSeconds := interval - int(timeSinceLastCode.Seconds())
			return fmt.Errorf("请求过于频繁，请 %d 秒后再试", remainingSeconds)
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Database error (not just "not found")
		return fmt.Errorf("查询验证码失败: %w", err)
	}
	// If record not found, it's OK to send a new code

	code := GenerateCode()
	expiresAt := now.Add(time.Duration(config.C.Security.VerificationCodeExpireMinutes) * time.Minute)
	return SendVerificationCode(email, code, expiresAt)
}

// LoginWithCode verifies code and returns JWT token & user info.
func LoginWithCode(email, code string) (string, *models.User, error) {
	var vc models.VerificationCode
	err := database.DB.Where("email = ? AND code = ? AND used = 0", email, code).
		Order("created_at desc").First(&vc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("验证码错误或不存在")
		}
		return "", nil, err
	}

	if time.Now().After(vc.ExpiresAt) {
		return "", nil, errors.New("验证码已过期")
	}

	if vc.ErrorCount >= config.C.Security.VerificationCodeMaxErrors {
		return "", nil, errors.New("验证码错误次数过多，请重新获取")
	}

	// mark used
	vc.Used = true
	if err := database.DB.Save(&vc).Error; err != nil {
		return "", nil, err
	}

	// find or create user
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{
				Email:   email,
				Role:    "user",
				Balance: 0,
			}
			if err := database.DB.Create(&user).Error; err != nil {
				return "", nil, err
			}
		} else {
			return "", nil, err
		}
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// GetUserByID returns user by ID
func GetUserByID(userID uint64) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// VerifyCode verifies a verification code for a user email (without login)
func VerifyCode(email, code string) error {
	// First, try to find the correct code
	var vc models.VerificationCode
	err := database.DB.Where("email = ? AND code = ? AND used = 0", email, code).
		Order("created_at desc").First(&vc).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Code is wrong, increment error count for the most recent unused code
			var recentCode models.VerificationCode
			if err := database.DB.Where("email = ? AND used = 0", email).
				Order("created_at desc").First(&recentCode).Error; err == nil {
				recentCode.ErrorCount++
				database.DB.Save(&recentCode)
			}
			return errors.New("验证码错误或不存在")
		}
		return err
	}

	// Check if expired
	if time.Now().After(vc.ExpiresAt) {
		return errors.New("验证码已过期")
	}

	// Check error count
	if vc.ErrorCount >= config.C.Security.VerificationCodeMaxErrors {
		return errors.New("验证码错误次数过多，请重新获取")
	}

	// Code is correct, mark as used
	vc.Used = true
	if err := database.DB.Save(&vc).Error; err != nil {
		return err
	}

	return nil
}


