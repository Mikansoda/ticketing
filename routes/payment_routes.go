package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ticketing/controller"
	"ticketing/middleware"
	"ticketing/repository"
	"ticketing/service"
)

func RegisterPaymentRoutes(r *gin.Engine, db *gorm.DB) {
	// Dependency injection
	bookingRepo := repository.NewBookingRepository(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	eventsRepo := repository.NewEventRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	ticketRepo := repository.NewTicketRepository(db)
	ticketTypeSvc := service.NewTicketTypeService(ticketTypeRepo, eventsRepo)
	paymentSvc := service.NewPaymentService(paymentRepo, bookingRepo, ticketRepo, ticketTypeSvc, db)
	paymentCtl := controller.NewPaymentController(paymentSvc)

	// User protected routes
	user := r.Group("/user", middleware.Auth("user"))
	{
		user.POST("/payments/xendit", middleware.RateLimit(10, time.Minute), paymentCtl.CreatePayment)
		user.GET("/payments", middleware.RateLimit(10, time.Minute), paymentCtl.GetUserPayments)
	}

	// Admin protected routes
	admin := r.Group("/admin", middleware.Auth("admin"))
	{
		admin.GET("/payments", paymentCtl.GetPayments)
		admin.POST("/payments/webhook/xendit", paymentCtl.XenditWebhook)
	}
}
