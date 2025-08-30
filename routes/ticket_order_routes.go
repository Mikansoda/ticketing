package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"

	"ticketing/controller"
	"ticketing/middleware"
	"ticketing/repository"
	"ticketing/service"
)

func RegisterBookingRoutes(r *gin.Engine, db *gorm.DB) {
	// repo
	bookingRepo := repository.NewBookingRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	visitorRepo := repository.NewVisitorRepo(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	// service
	bookingSvc := service.NewBookingService(ticketTypeRepo, bookingRepo, visitorRepo, ticketRepo)
	ticketSvc := service.NewTicketService(ticketRepo)
	// controller
	bookingCtl := controller.NewOrderController(bookingSvc)
	ticketCtl := controller.NewTicketsController(ticketSvc)
	
	// User protected routes
	user := r.Group("/user", middleware.Auth("user"))
	{
		user.GET("/bookings", bookingCtl.GetUserBookings)                // liat booking sendiri
		user.POST("/bookings", middleware.RateLimit(10, time.Minute), bookingCtl.CreateBooking)                // create booking
		user.GET("/tickets/:ticketId", ticketCtl.GetByID)     // liat tiket by ID
	}
	
	// Admin protected routes
	admin := r.Group("/admin", middleware.Auth("admin"))
	{
		admin.GET("/bookings", bookingCtl.GetBookings)                  // list semua booking
		admin.PATCH("/bookings/:id/status", bookingCtl.UpdateBookingStatus) // update status booking
		admin.PATCH("/tickets/:ticketId/status", ticketCtl.UpdateStatus) // update status tiket
	}

	// Routes utk both user & admin
	authAny := r.Group("/", middleware.Auth("user", "admin"))
	{
		authAny.GET("/bookings/:id", bookingCtl.GetBookingByID) // liat detail booking, dengan ticket
	}
}