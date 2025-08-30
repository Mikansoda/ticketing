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

func RegisterEventsRoutes(r *gin.Engine, db *gorm.DB) {
	// Dependency injection
	// Ticket Type
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	ticketTypeSvc := service.NewTicketTypeService(ticketTypeRepo, repository.NewEventRepository(db))

	// Events
	eventRepo := repository.NewEventRepository(db)
	eventSvc := service.NewEventService(eventRepo, ticketTypeSvc)
	eventCtl := controller.NewEventController(eventSvc)

	// Categories
	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryCtl := controller.NewCategoryController(categorySvc)

	// Image
	imageRepo := repository.NewEventImageRepository(db)
	imageSvc := service.NewEventImagesService(imageRepo)
	imageCtl := controller.NewEventImageController(imageSvc)

	// Publik
	r.GET("/events", eventCtl.GetEvents)
	r.GET("/events/:eventId", eventCtl.GetEventsByID)
	r.GET("/categories", categoryCtl.GetCategories)

	// Admin protected routes
	admin := r.Group("/admin", middleware.Auth("admin"))
	{
		admin.POST("/categories", categoryCtl.CreateCategory)
		admin.PATCH("/categories/:id", categoryCtl.UpdateCategory)
		admin.DELETE("/categories/:id", categoryCtl.DeleteCategory)
		admin.POST("/categories/:id/recover", categoryCtl.RecoverCategory)

		admin.POST("/events", eventCtl.CreateEvents)
		admin.PATCH("/events/:eventId", eventCtl.UpdateEvents)
		admin.DELETE("/events/:eventId", eventCtl.DeleteEvents)
		admin.POST("/events/:eventId/recover", eventCtl.RecoverEvents)
		// 5 req/menit
		admin.POST("/events/:eventId/images", middleware.RateLimit(5, time.Minute), imageCtl.UploadImage)
		admin.DELETE("/images/:imageId", middleware.RateLimit(10, time.Minute), imageCtl.DeleteImage)
		admin.POST("/images/:imageId/recover", middleware.RateLimit(10, time.Minute), imageCtl.RecoverImage)
	}
}
