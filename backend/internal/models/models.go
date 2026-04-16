package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel defines common fields for all tables.
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User represents users table.
type User struct {
	BaseModel
	Email   string  `gorm:"type:varchar(255);uniqueIndex:uk_email;not null" json:"email"`
	Role    string  `gorm:"type:varchar(20);not null;default:user;index:idx_role" json:"role"`
	Balance float64 `gorm:"type:decimal(10,2);not null;default:0" json:"balance"`
}

// Certificate represents certificates table.
type Certificate struct {
	BaseModel
	UserID     uint64    `gorm:"not null;index:idx_user_id" json:"user_id"`
	Domain     string    `gorm:"type:varchar(255);not null;index:idx_domain" json:"domain"`
	CertType   string    `gorm:"type:varchar(20);not null" json:"cert_type"` // single/wildcard
	CA         string    `gorm:"type:text;not null" json:"ca"`
	PrivateKey string    `gorm:"type:text;not null" json:"-"`
	PublicKey  string    `gorm:"type:text;not null" json:"public_key"`
	ExpiresAt  time.Time `gorm:"index:idx_expires_at;not null" json:"expires_at"`
	Status     string    `gorm:"type:varchar(20);not null;default:valid;index:idx_status" json:"status"`
}

// BalanceRecord represents balance_records table.
type BalanceRecord struct {
	BaseModel
	UserID        uint64  `gorm:"not null;index:idx_user_id" json:"user_id"`
	Type          string  `gorm:"type:varchar(20);not null;index:idx_type" json:"type"` // recharge/consume
	Amount        float64 `gorm:"type:decimal(10,2);not null" json:"amount"`
	PaymentMethod string  `gorm:"type:varchar(20)" json:"payment_method"` // alipay/wechat
	OrderNo       string  `gorm:"type:varchar(64);uniqueIndex:uk_order_no;not null" json:"order_no"`
	CertificateID *uint64 `gorm:"index" json:"certificate_id,omitempty"`
	Description   string  `gorm:"type:varchar(255)" json:"description"`
}

// VerificationCode represents verification_codes table.
type VerificationCode struct {
	BaseModel
	Email      string    `gorm:"type:varchar(255);index:idx_email_code,priority:1;not null" json:"email"`
	Code       string    `gorm:"type:varchar(6);index:idx_email_code,priority:2;not null" json:"code"`
	ErrorCount int       `gorm:"type:int;not null;default:0" json:"error_count"`
	ExpiresAt  time.Time `gorm:"index:idx_expires_at;not null" json:"expires_at"`
	Used       bool      `gorm:"type:tinyint(1);not null;default:0" json:"used"`
}

// OperationLog represents operation_logs table.
type OperationLog struct {
	BaseModel
	UserID       *uint64 `gorm:"index:idx_user_id" json:"user_id,omitempty"`
	UserEmail    string  `gorm:"type:varchar(255)" json:"user_email"`
	Operation    string  `gorm:"type:varchar(50);index:idx_operation_type;not null" json:"operation_type"`
	ResourceType string  `gorm:"type:varchar(50);index:idx_resource_type;not null" json:"resource_type"`
	ResourceID   *uint64 `json:"resource_id,omitempty"`
	Content      string  `gorm:"type:text" json:"content"`
	IPAddress    string  `gorm:"type:varchar(45)" json:"ip_address"`
}

// SystemConfig represents system_configs table.
type SystemConfig struct {
	BaseModel
	ConfigKey   string `gorm:"type:varchar(100);uniqueIndex:uk_config_key;not null" json:"config_key"`
	ConfigValue string `gorm:"type:text;not null" json:"config_value"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}


