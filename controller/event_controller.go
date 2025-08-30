package controller

import (
	"net/http"
	"strconv"
	"time"

	"ticketing/entity"
	"ticketing/service"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	service service.EventService
}

func NewEventController(s service.EventService) *EventController {
	return &EventController{service: s}
}

// Struct request for endpoint create event and update
type createEventsReq struct {
	Name        string    `json:"name" binding:"required"`
	CategoryID  uint      `json:"category_id" binding:"required"`
	Capacity    uint      `json:"capacity" binding:"required"`
	Description string    `json:"description" binding:"required"`
	City        string    `json:"city" binding:"required"`
	Country     string    `json:"country" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
}

type updateEventsReq struct {
	Name        *string    `json:"name,omitempty"`
	CategoryID  *uint      `json:"category_id,omitempty"`
	Capacity    *uint      `json:"capacity,omitempty"`
	EventStatus *string    `json:"event_status,omitempty"`
	Description *string    `json:"description,omitempty"`
	City        *string    `json:"city,omitempty"`
	Country     *string    `json:"country,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

// GET events
func (ctl *EventController) GetEvents(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")
	status := c.Query("status")
	dateStr := c.Query("date")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(offsetStr)

	var filterDate *time.Time
	if dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			filterDate = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid date format. Use YYYY-MM-DD",
				"detail":  err.Error(),
			})
			return
		}
	}

	events, err := ctl.service.GetEvents(c.Request.Context(), search, category, status, filterDate, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch events, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, events)
}

// GET event by ID
func (ctl *EventController) GetEventsByID(c *gin.Context) {
	idStr := c.Param("eventId")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	event, err := ctl.service.GetEventsByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Events not found",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, event)
}

// Create event (admin only)
func (ctl *EventController) CreateEvents(c *gin.Context) {
	var req createEventsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}

	event := &entity.Events{
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		Capacity:    req.Capacity,
		Description: req.Description,
		City:        req.City,
		Country:     req.Country,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	if err := ctl.service.CreateEvents(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create event, try again later",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Events successfully created",
		"data":    event,
	})
}

// Update event (admin only)
func (ctl *EventController) UpdateEvents(c *gin.Context) {
	idStr := c.Param("eventId")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	var req updateEventsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid input data",
			"detail":  err.Error(),
		})
		return
	}
	event := &entity.Events{ID: id}

	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.CategoryID != nil {
		event.CategoryID = *req.CategoryID
	}
	if req.Capacity != nil {
		event.Capacity = *req.Capacity
	}
	if req.EventStatus != nil {
		event.EventStatus = *req.EventStatus
	}
	if req.Description != nil {
		event.Description = *req.Description
	}
	if req.City != nil {
		event.City = *req.City
	}
	if req.Country != nil {
		event.Country = *req.Country
	}
	if req.StartDate != nil {
		event.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		event.EndDate = *req.EndDate
	}

	updatedEvents, err := ctl.service.UpdateEvents(c.Request.Context(), event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update event",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Events successfully updated",
		"data":    updatedEvents,
	})
}

// Delete event (admin only)
func (ctl *EventController) DeleteEvents(c *gin.Context) {
	idStr := c.Param("eventId")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.DeleteEvents(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to delete event, try again later"

		if err.Error() == "event not found" {
			status = http.StatusNotFound
			msg = "Events not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Events successfully deleted",
	})
}

// Recover event (admin only)
func (ctl *EventController) RecoverEvents(c *gin.Context) {
	idStr := c.Param("eventId")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
			"detail":  err.Error(),
		})
		return
	}
	id := uint(idUint64)

	if err := ctl.service.RecoverEvents(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := "Failed to recover event, try again later"

		if err.Error() == "event not found" {
			status = http.StatusNotFound
			msg = "Event not found"
		}

		c.JSON(status, gin.H{
			"message": msg,
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Event successfully recovered"})
}
