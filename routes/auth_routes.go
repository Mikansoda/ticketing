package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ticketing/controller"
	"ticketing/middleware"
	"ticketing/repository"
	"ticketing/service"
)

func RegisterAuthRoutes(r *gin.Engine, db *gorm.DB) {
	authRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(authRepo)
	authCtl := controller.NewAuthController(authSvc)

	authApi := r.Group("/auth")
	{
		authApi.POST("/register", authCtl.Register)
		authApi.POST("/verify-otp", authCtl.VerifyOTP)
		authApi.POST("/login", authCtl.Login)
		authApi.POST("/refresh", authCtl.Refresh)
		authApi.POST("/logout", authCtl.Logout)

		authApi.GET("/profile", middleware.Auth("user", "admin"), authCtl.Profile)
		authApi.GET("/admin/dashboard", middleware.Auth("admin"), authCtl.AdminDashboard)
	}
}
