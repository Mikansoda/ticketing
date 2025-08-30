package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"ticketing/entity"
	"ticketing/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderController struct {
	service service.BookingService
}

func NewOrderController(s service.BookingService) *OrderController {
	return &OrderController{service: s}
}

// Struct request for creating booking
var createBookingReq struct {
    UserID       string           `json:"user_id"`
    TicketTypeID uint             `json:"ticket_type_id"` 
    Quantity     uint             `json:"quantity"`
    Visitors     []entity.Visitors `json:"visitors"`
}

// GET bookings (admin only)
func (ctl *OrderController) GetBookings(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(offsetStr)

	var (
		bookings []entity.Bookings
		err      error
	)
	if status != "" {
	bookings, err = ctl.service.GetBookingsByStatus(c.Request.Context(), status, limit, offset)
	} else {
	bookings, err = ctl.service.GetBookings(c.Request.Context(), limit, offset)
    }

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch bookings, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// GET bookings (self-requested by user)
func (ctl *OrderController) GetUserBookings(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
			"detail":  err.Error(),
		})
		return
	}
	bookings, err := ctl.service.GetBookingsByUser(c.Request.Context(), userID)
	if err != nil {
		if err.Error() == "bookings not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "No booking found, let's start creating bookings",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch your bookings, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func (ctl *OrderController) GetBookingByID(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid booking ID"})
		return
	}

	currentUserIDStr, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized, not your booking",
		})
		return
	}

	uid, err := uuid.Parse(currentUserIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID",
			"detail":  err.Error(),
		})
		return
	}

	role, _ := c.Get("role")
	isAdmin := role == "admin"

	booking, err := ctl.service.GetBookingByID(
		c.Request.Context(),
		bookingID,
		uid, 
		isAdmin,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Login first to access this page",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, booking)
}


// Create booking (self-created by user)
func (ctl *OrderController) CreateBooking(c *gin.Context) {
    if err := c.ShouldBindJSON(&createBookingReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request", 
			"detail": err.Error(),
		})
        return
    }

    uid, err := uuid.Parse(createBookingReq.UserID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid user ID", 
			"detail": err.Error(),
		})
        return
    }

    booking, err := ctl.service.CreateBooking(c.Request.Context(), uid, createBookingReq.TicketTypeID, createBookingReq.Quantity, createBookingReq.Visitors)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create booking, try again later", 
			"detail": err.Error(),
		})
        return
    }
    c.JSON(http.StatusOK, booking)
}

// Update booking status (admin only)
func (ctl *OrderController) UpdateBookingStatus(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid booking ID",
			"detail":  err.Error(),
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid input data",
			"detail":  err.Error(),
		})
		return
	}
	if err := ctl.service.UpdateBookingStatus(c.Request.Context(), bookingID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to update booking status, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("booking %s status updated to %s", bookingID, req.Status),
	})
}
