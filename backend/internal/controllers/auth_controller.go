package controllers

import (
	"net/http"

	"certhub-backend/internal/config"
	"certhub-backend/internal/services"
	"certhub-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type sendCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

type loginRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// SendCode handles POST /api/v1/auth/send-code
func SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}

	if !services.ValidateEmailFormat(req.Email) {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "请输入正确的邮箱格式")
		return
	}

	if err := services.RequestVerificationCode(req.Email); err != nil {
		utils.Error(c, http.StatusTooManyRequests, http.StatusTooManyRequests, err.Error())
		return
	}

	utils.Success(c, nil)
}

// Login handles POST /api/v1/auth/login
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}

	token, user, err := services.LoginWithCode(req.Email, req.Code)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

// AdminLogin reuses LoginWithCode but restricts to admin email.
// POST /api/v1/admin/auth/login
func AdminLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, "参数错误")
		return
	}

	adminEmail := config.C.Server.Mode // placeholder to avoid empty usage
	_ = adminEmail
	// In a real deployment, admin emails should be stored in system_configs or env.
	// Here we simply treat any existing user with role=admin as admin.

	token, user, err := services.LoginWithCode(req.Email, req.Code)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	if user.Role != "admin" {
		utils.Error(c, http.StatusForbidden, http.StatusForbidden, "该邮箱不是管理员邮箱")
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}


