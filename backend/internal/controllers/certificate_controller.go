package controllers

import (
	"net/http"
	"strconv"
	"time"

	"certhub-backend/internal/services"
	"certhub-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type generateDNSRequest struct {
	Domain string `json:"domain" binding:"required"`
}

type generateCertRequest struct {
	Domain    string                `json:"domain" binding:"required"`
	DNSRecord services.DNSRecord    `json:"dns_record" binding:"required"`
}

// POST /api/v1/certificates/generate-dns
func GenerateDNS(c *gin.Context) {
	var req generateDNSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}
	record, err := services.GenerateDNSRecord(req.Domain)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"dns_record": record,
		"domain":     req.Domain,
	})
}

// POST /api/v1/certificates/generate
func GenerateCertificate(c *gin.Context) {
	var req generateCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}

	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	// 先检查余额
	if err := services.CheckBalance(userID, req.Domain); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	// 先创建处理中的证书记录
	cert, err := services.CreatePendingCertificate(userID, req.Domain, req.DNSRecord)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	// 异步处理证书生成
	go func() {
		services.ProcessCertificateAsync(cert.ID, userID, req.Domain, req.DNSRecord)
	}()

	utils.Success(c, gin.H{
		"status":      "processing",
		"certificate": cert,
		"message":     "证书正在生成中，请稍后查看",
	})
}

// GET /api/v1/certificates
func ListUserCertificates(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)

	domain := c.Query("domain")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := services.GetUserCertificates(userID, domain, status, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取证书列表失败")
		return
	}

	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GET /api/v1/certificates/:id
func GetCertificateDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)
	roleAny, _ := c.Get("userRole")
	role := roleAny.(string)

	cert, err := services.GetCertificateDetail(userID, role, id)
	if err != nil {
		if err.Error() == "证书不存在" {
			utils.Error(c, http.StatusNotFound, http.StatusNotFound, err.Error())
		} else if err.Error() == "无权访问该证书" {
			utils.Error(c, http.StatusForbidden, http.StatusForbidden, err.Error())
		} else {
			utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取证书详情失败")
		}
		return
	}
	utils.Success(c, gin.H{
		"certificate": cert,
	})
}

// POST /api/v1/certificates/:id/download-private-key
type downloadPrivateKeyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

func DownloadPrivateKey(c *gin.Context) {
	var req downloadPrivateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误：验证码为6位数字")
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	userIDAny, _ := c.Get("userID")
	userID := userIDAny.(uint64)
	roleAny, _ := c.Get("userRole")
	role := roleAny.(string)

	// Get certificate first to check ownership
	cert, err := services.GetCertificateDetail(userID, role, id)
	if err != nil {
		if err.Error() == "证书不存在" {
			utils.Error(c, http.StatusNotFound, http.StatusNotFound, err.Error())
		} else {
			utils.Error(c, http.StatusForbidden, http.StatusForbidden, err.Error())
		}
		return
	}

	// Get user email to verify code
	user, err := services.GetUserByID(userID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	// Verify the verification code
	if err := services.VerifyCode(user.Email, req.Code); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	// Decrypt and return private key
	private, err := utils.DecryptAES(cert.PrivateKey)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "私钥解密失败")
		return
	}

	utils.Success(c, gin.H{
		"private_key": private,
	})
}

// Admin endpoints

type adminCreateCertRequest struct {
	UserEmail string `json:"user_email" binding:"required"`
	Domain    string `json:"domain" binding:"required"`
	CA        string `json:"ca" binding:"required"`
	Private   string `json:"private_key" binding:"required"`
	Public    string `json:"public_key" binding:"required"`
	ExpiresAt string `json:"expires_at" binding:"required"`
	CreatedAt string `json:"created_at"`
}

// POST /api/v1/admin/certificates
func AdminCreateCertificate(c *gin.Context) {
	var req adminCreateCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}
	expires, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "过期时间格式错误")
		return
	}
	var created time.Time
	if req.CreatedAt != "" {
		created, err = time.Parse(time.RFC3339, req.CreatedAt)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "申请时间格式错误")
			return
		}
	}
	cert, err := services.AdminCreateCertificate(req.UserEmail, req.Domain, req.CA, req.Private, req.Public, created, expires)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"certificate": cert})
}

type adminUpdateCertRequest struct {
	CA        string `json:"ca"`
	Private   string `json:"private_key"`
	Public    string `json:"public_key"`
	ExpiresAt string `json:"expires_at"`
	UserEmail string `json:"user_email"`
}

// PUT /api/v1/admin/certificates/:id
func AdminUpdateCertificate(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	var req adminUpdateCertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}
	var expires time.Time
	var err error
	if req.ExpiresAt != "" {
		expires, err = time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "过期时间格式错误")
			return
		}
	}
	cert, err := services.AdminUpdateCertificate(id, req.CA, req.Private, req.Public, expires, req.UserEmail)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}
	utils.Success(c, gin.H{"certificate": cert})
}

// DELETE /api/v1/admin/certificates/:id
func AdminDeleteCertificate(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)
	if err := services.AdminDeleteCertificate(id); err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "删除证书失败")
		return
	}
	utils.Success(c, nil)
}

// GET /api/v1/admin/certificates
func AdminListCertificates(c *gin.Context) {
	userEmail := c.Query("user_email")
	domain := c.Query("domain")
	status := c.Query("status")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	list, total, err := services.ListCertificatesForAdmin(userEmail, domain, status, startTime, endTime, page, pageSize)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, http.StatusInternalServerError, "获取证书列表失败")
		return
	}
	utils.Success(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}


