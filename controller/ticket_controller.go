package controller

import (
	"net/http"

	"ticketing/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TicketsController struct {
	service service.TicketService
}

func NewTicketsController(s service.TicketService) *TicketsController {
	return &TicketsController{service: s}
}

// GET tickets by ID
func (ctl *TicketsController) GetByID(c *gin.Context) {
	idStr := c.Param("ticketId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ticket ID", "detail": err.Error()})
		return
	}

	ticket, err := ctl.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Ticket not found", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ticket)
}

// Update ticket status
type updateTicketStatusReq struct {
	Status string `json:"status" binding:"required"`
}

func (ctl *TicketsController) UpdateStatus(c *gin.Context) {
	idStr := c.Param("ticketId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ticket ID", 
			"detail": err.Error(),
		})
		return
	}

	var req updateTicketStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data", 
			"detail": err.Error(),
		})
		return
	}

	ticket, err := ctl.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update ticket status", 
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket status updated", 
		"data": ticket,
	})
}
