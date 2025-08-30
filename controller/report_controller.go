package controller

import (
	"net/http"
	"strconv"

	"ticketing/service"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	service service.ReportService
}

func NewReportController(s service.ReportService) *ReportController {
	return &ReportController{service: s}
}

// GET report by period
func (ctl *ReportController) GetMonthlyReport(c *gin.Context) {
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid month"})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid year"})
		return
	}

	totalTickets, totalRevenue, err := ctl.service.GetMonthlySummary(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch monthly report",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"month":         month,
		"year":          year,
		"total_tickets": totalTickets,
		"total_revenue": totalRevenue,
	})
}

// GET report by event
func (ctl *ReportController) GetEventReport(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid event ID"})
		return
	}

	totalTickets, totalRevenue, err := ctl.service.GetTicketsByEvent(c.Request.Context(), uint(eventID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch event report",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":      eventID,
		"total_tickets": totalTickets,
		"total_revenue": totalRevenue,
	})
}
