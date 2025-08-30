package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ticketing/controller"
	"ticketing/middleware"
	"ticketing/repository"
	"ticketing/service"
)

func RegisterReportRoutes(r *gin.Engine, db *gorm.DB) {
	// Dependency injection
	reportRepo := repository.NewReportRepository(db)

	reportSvc := service.NewReportService(reportRepo)
	reportCtl := controller.NewReportController(reportSvc)

	// Admin protected routes
	admin := r.Group("/admin", middleware.Auth("admin"))
	{
		// Laporan ringkasan tiket terjual & pendapatan bulanan
		// reports/monthly?month=8&year=2025
		admin.GET("/reports/summary", reportCtl.GetMonthlyReport)
		// Laporan per event
		admin.GET("/reports/event/:id", reportCtl.GetEventReport)
	}
}
