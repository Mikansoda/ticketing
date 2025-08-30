package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB,  xenditAPIKey string) *gin.Engine {
	r := gin.Default()

	// use logger middleware (global)
	r.Use(gin.Recovery())

	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	RegisterAuthRoutes(r, db)
	RegisterEventsRoutes(r, db)
	RegisterTicketTypeRoutes(r, db)
	RegisterBookingRoutes(r, db)
	RegisterPaymentRoutes(r, db)
	RegisterReportRoutes(r, db)
	return r
}
