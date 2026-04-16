package routes

import (
	"certhub-backend/internal/controllers"
	"certhub-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		// auth
		api.POST("/auth/send-code", controllers.SendCode)
		api.POST("/auth/login", controllers.Login)

		// admin auth
		api.POST("/admin/auth/login", controllers.AdminLogin)

		// user protected routes
		user := api.Group("")
		user.Use(middleware.JWTAuth())
		{
			// certificates - user
			user.POST("/certificates/generate-dns", controllers.GenerateDNS)
			user.POST("/certificates/generate", controllers.GenerateCertificate)
			user.GET("/certificates", controllers.ListUserCertificates)
			user.GET("/certificates/:id", controllers.GetCertificateDetail)
			user.POST("/certificates/:id/download-private-key", controllers.DownloadPrivateKey)

			// balance
			user.GET("/balance", controllers.GetBalance)
			user.GET("/balance/records", controllers.GetBalanceRecords)
			user.POST("/balance/recharge", controllers.CreateRechargeOrder)
		}

		// admin protected routes
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.AdminOnly())
		{
			admin.GET("/certificates", controllers.AdminListCertificates)
			admin.POST("/certificates", controllers.AdminCreateCertificate)
			admin.PUT("/certificates/:id", controllers.AdminUpdateCertificate)
			admin.DELETE("/certificates/:id", controllers.AdminDeleteCertificate)
		}
	}

	return r
}


