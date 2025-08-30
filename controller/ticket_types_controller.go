package controller

import (
	"net/http"
	"strconv"
	"time"

	"ticketing/entity"
	"ticketing/service"

	"github.com/gin-gonic/gin"
)

type TicketTypeController struct {
	service service.TicketTypeService
}

func NewTicketTypeController(s service.TicketTypeService) *TicketTypeController {
	return &TicketTypeController{service: s}
}

// Struct request for endpoint create ticket type and update
type createTicketTypeReq struct {
	EventID uint      `json:"event_ID" binding:"required"`
	Name    string    `json:"name" binding:"required"`
	Price   float64      `json:"price" binding:"required"`
	Quota   uint      `json:"quota" binding:"required"`
	Date    *time.Time `json:"date" binding:"required"`
}

type updateTicketTypeReq struct {
	EventID *uint      `json:"event_ID,omitempty"`
	Name    *string    `json:"name,omitempty"`
	Status  *string    `json:"status,omitempty"`
	Price   *float64      `json:"price,omitempty"`
	Quota   *uint      `json:"quota,omitempty"`
	Date    *time.Time `json:"date,omitempty"`
}

// GET ticket types (e.g. VIP, regular, CAT 1, etc.)
func (ctl *TicketTypeController) GetTicketTypes(c *gin.Context) {
	event := c.Query("event")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(offsetStr)

	ticketTypes, err := ctl.service.GetTicketTypes(c.Request.Context(), event, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch ticket types, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ticketTypes)
}

// GET ticket type by ID
func (ctl *TicketTypeController) GetTicketTypeByID(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ticket type ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)
	ticketType, err := ctl.service.GetTicketTypeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Ticket type not found",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, ticketType)
}

// Create ticket type (admin only)
func (ctl *TicketTypeController) CreateTicketType(c *gin.Context) {
	var req createTicketTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	ticketType := &entity.TicketTypes{
		EventID: req.EventID,
		Name:    req.Name,
		Price:   req.Price,
		Quota:   req.Quota,
		Date:    req.Date,
	}
	if err := ctl.service.CreateTicketType(c.Request.Context(), ticketType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create ticket type, try again later",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket type successfully created",
		"data":    ticketType,
	})
}

// Update ticket type (admin only)
func (ctl *TicketTypeController) UpdateTicketType(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ticket ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	var req updateTicketTypeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	existing, err := ctl.service.GetTicketTypeByIDIncludeDeleted(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Ticket type not found",
			"detail":  err.Error(),
		})
		return
	}

	if req.EventID != nil {
		existing.EventID = *req.EventID
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.Quota != nil {
		existing.Quota = *req.Quota
	}
	if req.Date != nil {
		existing.Date = req.Date
	}
	existing.UpdatedAt = time.Now()

	updatedTicketType, err := ctl.service.UpdateTicketType(c.Request.Context(), existing)
	if err != nil {
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Failed to update ticket type, try again later",
		"detail":  err.Error(),
	})
	return
    }
	c.JSON(http.StatusOK, gin.H{
		"message": "TicketType successfully updated",
		"data":    updatedTicketType,
	})
}

// Delete ticket type (admin only)
func (ctl *TicketTypeController) DeleteTicketType(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ticket type ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.DeleteTicketType(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to delete ticket type, try again later"

		if err.Error() == "ticket type not found" {
			status = http.StatusNotFound
			msg = "Ticket type not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket type successfully deleted",
	})
}

// Recover ticket type
func (ctl *TicketTypeController) RecoverTicketType(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ticket type ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.RecoverTicketType(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to recover ticket type, try again later"

		if err.Error() == "ticket type not found" {
			status = http.StatusNotFound
			msg = "TicketType not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Ticket type successfully recovered",
	})
}
