package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"ticketing/controller"
	"ticketing/middleware"
	"ticketing/repository"
	"ticketing/service"
)

func RegisterTicketTypeRoutes(r *gin.Engine, db *gorm.DB) {
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	eventRepo := repository.NewEventRepository(db)

	ticketTypeSvc := service.NewTicketTypeService(ticketTypeRepo, eventRepo)
	ticketTypeCtl := controller.NewTicketTypeController(ticketTypeSvc)

	// Public routes
	r.GET("/ticket-types", ticketTypeCtl.GetTicketTypes)
	r.GET("/ticket-types/:id", ticketTypeCtl.GetTicketTypeByID)

	// Admin protected routes
	admin := r.Group("/admin", middleware.Auth("admin"))
	{
		admin.POST("/ticket-types", ticketTypeCtl.CreateTicketType)
		admin.PATCH("/ticket-types/:id", ticketTypeCtl.UpdateTicketType)
		admin.DELETE("/ticket-types/:id", ticketTypeCtl.DeleteTicketType)
		admin.POST("/ticket-types/:id/recover", ticketTypeCtl.RecoverTicketType)
	}
}
