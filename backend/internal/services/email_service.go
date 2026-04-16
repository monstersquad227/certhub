package services

import (
	"fmt"
	"time"

	"certhub-backend/internal/config"
	"certhub-backend/internal/database"
	"certhub-backend/internal/models"

	gomail "github.com/go-gomail/gomail"
)

// SendVerificationCode sends a 6-digit verification code to the given email
// and saves it to the database with expiration and error count.
func SendVerificationCode(email, code string, expiresAt time.Time) error {
	cfg := config.C.Email

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "CertHub 验证码")
	m.SetBody("text/plain", fmt.Sprintf("您的验证码是：%s，有效期 %d 分钟。", code, config.C.Security.VerificationCodeExpireMinutes))

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.Username, cfg.Password)
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	vc := &models.VerificationCode{
		Email:     email,
		Code:      code,
		ExpiresAt: expiresAt,
	}
	return database.DB.Create(vc).Error
}


