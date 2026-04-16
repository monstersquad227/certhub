package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"certhub-backend/internal/database"
	"certhub-backend/internal/models"
	"certhub-backend/internal/utils"

	"gorm.io/gorm"
)

// DNSRecord represents DNS TXT record for ACME DNS-01.
type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GenerateDNSRecord creates a real ACME DNS-01 challenge using Let's Encrypt.
// It contacts the ACME server to create an order and returns the required DNS TXT record.
func GenerateDNSRecord(domain string) (*DNSRecord, error) {
	if domain == "" {
		return nil, errors.New("域名不能为空")
	}
	trimmed := strings.TrimSpace(domain)
	if !ValidateDomainFormat(trimmed) {
		return nil, errors.New("请输入正确的域名格式")
	}

	// Get ACME client
	client, err := GetACMEClient()
	if err != nil {
		return nil, fmt.Errorf("ACME 客户端初始化失败: %w", err)
	}

	// Create order and get DNS-01 challenge
	pending, err := client.CreateOrder(trimmed)
	if err != nil {
		return nil, fmt.Errorf("创建 ACME 订单失败: %w", err)
	}

	return &DNSRecord{
		Type:  "TXT",
		Name:  pending.DNSRecordName,
		Value: pending.DNSRecordValue,
	}, nil
}

// ValidateDomainFormat validates single or wildcard domain.
func ValidateDomainFormat(domain string) bool {
	if strings.HasPrefix(domain, "*.") {
		// wildcard: remove "*." and validate remaining as basic domain
		domain = strings.TrimPrefix(domain, "*.")
	}
	// very simple check: contains dot and no spaces
	if strings.Contains(domain, " ") || !strings.Contains(domain, ".") {
		return false
	}
	return true
}

// CheckBalance checks if user has enough balance for certificate
func CheckBalance(userID uint64, domain string) error {
	isWildcard := strings.HasPrefix(domain, "*.")
	price := GetCertPrice(isWildcard)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if user.Balance < price {
		return errors.New("余额不足")
	}
	return nil
}

// CreatePendingCertificate creates a certificate record with pending status
func CreatePendingCertificate(userID uint64, domain string, dns DNSRecord) (*models.Certificate, error) {
	isWildcard := strings.HasPrefix(domain, "*.")
	
	cert := models.Certificate{
		UserID:     userID,
		Domain:     domain,
		CertType:   map[bool]string{true: "wildcard", false: "single"}[isWildcard],
		CA:         "Let's Encrypt",
		PrivateKey: "",
		PublicKey:  "",
		ExpiresAt:  time.Now().Add(90 * 24 * time.Hour),
		Status:     "pending",
	}
	
	if err := database.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	
	return &cert, nil
}

// ProcessCertificateAsync processes certificate generation asynchronously
func ProcessCertificateAsync(certID uint64, userID uint64, domain string, dns DNSRecord) {
	isWildcard := strings.HasPrefix(domain, "*.")
	price := GetCertPrice(isWildcard)

	// Get ACME client
	acmeClient, err := GetACMEClient()
	if err != nil {
		updateCertificateStatus(certID, "failed", fmt.Sprintf("ACME 客户端初始化失败: %v", err))
		return
	}

	// Get pending challenge
	pending, ok := GetPendingChallenge(domain)
	if !ok {
		updateCertificateStatus(certID, "failed", "未找到 DNS 挑战记录")
		return
	}

	// Verify DNS challenge
	if err := acmeClient.VerifyDNSChallenge(pending); err != nil {
		updateCertificateStatus(certID, "failed", fmt.Sprintf("DNS 挑战验证失败。请确保已正确配置 DNS TXT 记录，并等待 DNS 解析生效（通常需要几分钟）。错误详情：%v", err))
		return
	}

	// Finalize order and get certificate
	certResult, err := acmeClient.FinalizeOrder(pending)
	if err != nil {
		updateCertificateStatus(certID, "failed", fmt.Sprintf("证书生成失败。这可能是因为 DNS 验证未通过或 Let's Encrypt 服务器问题。请检查 DNS 配置后重新申请。错误详情：%v", err))
		return
	}

	// Encrypt private key
	encryptedPriv, err := utils.EncryptAES(certResult.PrivateKey)
	if err != nil {
		updateCertificateStatus(certID, "failed", fmt.Sprintf("私钥加密失败: %v", err))
		return
	}

	// Update certificate with real data
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		// Update certificate
		if err := tx.Model(&models.Certificate{}).Where("id = ?", certID).Updates(map[string]interface{}{
			"ca":          certResult.CA,
			"private_key": encryptedPriv,
			"public_key":  certResult.Certificate,
			"expires_at":  certResult.ExpiresAt,
			"status":      "valid",
		}).Error; err != nil {
			return err
		}

		// Deduct balance
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if user.Balance < price {
			return errors.New("余额不足")
		}

		// Create balance record
		record := models.BalanceRecord{
			UserID:        userID,
			Type:          "consume",
			Amount:        -price,
			OrderNo:       fmt.Sprintf("C%d", time.Now().UnixNano()),
			CertificateID: &certID,
			Description:   fmt.Sprintf("%s证书申请", map[bool]string{true: "泛域名", false: "单域名"}[isWildcard]),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		// Update user balance
		if err := tx.Model(&models.User{}).Where("id = ? AND balance >= ?", userID, price).
			Update("balance", gorm.Expr("balance - ?", price)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		updateCertificateStatus(certID, "failed", fmt.Sprintf("保存证书失败: %v", err))
		return
	}

	// Clean up pending challenge
	challengesMu.Lock()
	delete(pendingChallenges, domain)
	challengesMu.Unlock()
}

// updateCertificateStatus updates certificate status
func updateCertificateStatus(certID uint64, status string, errorMsg string) {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		// Store error message in CA field with a prefix to identify it as an error
		updates["ca"] = "[ERROR] " + errorMsg
		updates["public_key"] = errorMsg // Also store in public_key for better access
	}
	database.DB.Model(&models.Certificate{}).Where("id = ?", certID).Updates(updates)
}

// IssueCertificate issues a real certificate using ACME DNS-01 challenge.
func IssueCertificate(userID uint64, domain string, dns DNSRecord) (*models.Certificate, error) {
	isWildcard := strings.HasPrefix(domain, "*.")
	price := GetCertPrice(isWildcard)

	// Check balance and deduct inside transaction
	var cert models.Certificate
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if user.Balance < price {
			return errors.New("余额不足")
		}

		// Get ACME client
		acmeClient, err := GetACMEClient()
		if err != nil {
			return fmt.Errorf("ACME 客户端初始化失败: %w", err)
		}

		// Get pending challenge
		pending, ok := GetPendingChallenge(domain)
		if !ok {
			return errors.New("未找到 DNS 挑战记录，请先生成 DNS 记录")
		}

		// Verify DNS challenge (assume DNS record has been added by user)
		if err := acmeClient.VerifyDNSChallenge(pending); err != nil {
			return fmt.Errorf("DNS 挑战验证失败: %w", err)
		}

		// Finalize order and get certificate
		certResult, err := acmeClient.FinalizeOrder(pending)
		if err != nil {
			return fmt.Errorf("证书生成失败: %w", err)
		}

		// Encrypt private key
		encryptedPriv, err := utils.EncryptAES(certResult.PrivateKey)
		if err != nil {
			return err
		}

		cert = models.Certificate{
			UserID:     userID,
			Domain:     domain,
			CertType:   map[bool]string{true: "wildcard", false: "single"}[isWildcard],
			CA:         certResult.CA,
			PrivateKey: encryptedPriv,
			PublicKey:  certResult.Certificate,
			ExpiresAt:  certResult.ExpiresAt,
			Status:     "valid",
		}
		if err := tx.Create(&cert).Error; err != nil {
			return err
		}

		// consume balance & record
		record := models.BalanceRecord{
			UserID:        userID,
			Type:          "consume",
			Amount:        -price,
			OrderNo:       fmt.Sprintf("C%d", time.Now().UnixNano()),
			CertificateID: &cert.ID,
			Description:   fmt.Sprintf("%s证书申请", map[bool]string{true: "泛域名", false: "单域名"}[isWildcard]),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("id = ? AND balance >= ?", userID, price).
			Update("balance", gorm.Expr("balance - ?", price)).Error; err != nil {
			return err
		}

		// Clean up pending challenge
		challengesMu.Lock()
		delete(pendingChallenges, domain)
		challengesMu.Unlock()

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetUserCertificates returns paginated certificates for user.
func GetUserCertificates(userID uint64, domain, status string, page, pageSize int) ([]models.Certificate, int64, error) {
	var (
		list  []models.Certificate
		total int64
	)
	query := database.DB.Model(&models.Certificate{}).Where("user_id = ?", userID)
	if domain != "" {
		query = query.Where("domain LIKE ?", "%"+domain+"%")
	}
	
	// Filter by status - but handle special cases
	filterStatus := status
	if status == "expired" || status == "expiring" {
		// These are computed statuses, filter by valid status first
		filterStatus = "valid"
	}
	if filterStatus != "" && filterStatus != "expired" && filterStatus != "expiring" {
		query = query.Where("status = ?", filterStatus)
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
	
	// Compute display status for each certificate
	now := time.Now()
	var filteredList []models.Certificate
	for _, cert := range list {
		// If status is pending or failed, keep it as is
		if cert.Status == "pending" || cert.Status == "failed" {
			if status == "" || status == cert.Status {
				filteredList = append(filteredList, cert)
			}
			continue
		}
		
		// For valid certificates, compute display status based on expiration
		if cert.Status == "valid" {
			daysUntilExpiry := int(cert.ExpiresAt.Sub(now).Hours() / 24)
			displayStatus := cert.Status
			
			if daysUntilExpiry < 0 {
				displayStatus = "expired"
			} else if daysUntilExpiry <= 30 {
				displayStatus = "expiring"
			}
			
			// Apply status filter
			if status == "" || status == displayStatus {
				// Temporarily update status for display
				cert.Status = displayStatus
				filteredList = append(filteredList, cert)
			}
		} else {
			// Other statuses
			if status == "" || status == cert.Status {
				filteredList = append(filteredList, cert)
			}
		}
	}
	
	// Recalculate total if status filter was applied
	if status == "expired" || status == "expiring" {
		total = int64(len(filteredList))
	}
	
	return filteredList, total, nil
}

// GetCertificateDetail returns certificate by id, ensuring ownership if userRole is user.
func GetCertificateDetail(requestUserID uint64, role string, certID uint64) (*models.Certificate, error) {
	var cert models.Certificate
	if err := database.DB.First(&cert, certID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("证书不存在")
		}
		return nil, err
	}
	if role != "admin" && cert.UserID != requestUserID {
		return nil, errors.New("无权访问该证书")
	}
	return &cert, nil
}

// ListCertificatesForAdmin returns paginated certificates for admin filters.
func ListCertificatesForAdmin(userEmail, domain, status, startTime, endTime string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var (
		certs []models.Certificate
		total int64
	)

	query := database.DB.Model(&models.Certificate{})

	if domain != "" {
		query = query.Where("domain LIKE ?", "%"+domain+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}

	if userEmail != "" {
		var user models.User
		if err := database.DB.Where("email = ?", userEmail).First(&user).Error; err == nil {
			query = query.Where("user_id = ?", user.ID)
		} else {
			return []map[string]interface{}{}, 0, nil
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&certs).Error; err != nil {
		return nil, 0, err
	}

	// join with user emails
	var userIDs []uint64
	for _, c := range certs {
		userIDs = append(userIDs, c.UserID)
	}
	var users []models.User
	_ = database.DB.Where("id IN ?", userIDs).Find(&users).Error
	userEmailMap := make(map[uint64]string)
	for _, u := range users {
		userEmailMap[u.ID] = u.Email
	}

	var result []map[string]interface{}
	for _, c := range certs {
		result = append(result, map[string]interface{}{
			"id":         c.ID,
			"user_email": userEmailMap[c.UserID],
			"domain":     c.Domain,
			"cert_type":  c.CertType,
			"created_at": c.CreatedAt,
			"expires_at": c.ExpiresAt,
			"status":     c.Status,
		})
	}
	return result, total, nil
}

// AdminCreateCertificate allows admin to manually create certificate record.
func AdminCreateCertificate(userEmail, domain, ca, privateKey, publicKey string, createdAt, expiresAt time.Time) (*models.Certificate, error) {
	var user models.User
	if err := database.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	encPriv, err := utils.EncryptAES(privateKey)
	if err != nil {
		return nil, err
	}
	cert := models.Certificate{
		UserID:     user.ID,
		Domain:     domain,
		CertType:   "single",
		CA:         ca,
		PrivateKey: encPriv,
		PublicKey:  publicKey,
		ExpiresAt:  expiresAt,
		Status:     "valid",
	}
	if !createdAt.IsZero() {
		cert.CreatedAt = createdAt
	}
	if err := database.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// AdminUpdateCertificate updates certificate fields.
func AdminUpdateCertificate(id uint64, ca, privateKey, publicKey string, expiresAt time.Time, userEmail string) (*models.Certificate, error) {
	var cert models.Certificate
	if err := database.DB.First(&cert, id).Error; err != nil {
		return nil, errors.New("证书不存在")
	}

	if userEmail != "" {
		var user models.User
		if err := database.DB.Where("email = ?", userEmail).First(&user).Error; err != nil {
			return nil, errors.New("用户不存在")
		}
		cert.UserID = user.ID
	}
	if ca != "" {
		cert.CA = ca
	}
	if privateKey != "" {
		enc, err := utils.EncryptAES(privateKey)
		if err != nil {
			return nil, err
		}
		cert.PrivateKey = enc
	}
	if publicKey != "" {
		cert.PublicKey = publicKey
	}
	if !expiresAt.IsZero() {
		cert.ExpiresAt = expiresAt
	}
	if err := database.DB.Save(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// AdminDeleteCertificate deletes certificate by id.
func AdminDeleteCertificate(id uint64) error {
	return database.DB.Delete(&models.Certificate{}, id).Error
}


