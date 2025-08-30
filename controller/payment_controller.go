package controller

import (
	"bytes"
	"ticketing/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentController struct {
	service service.PaymentService
}

func NewPaymentController(s service.PaymentService) *PaymentController {
	return &PaymentController{service: s}
}

// Request body struct
type createPaymentReq struct {
	BookingID uuid.UUID `json:"booking_id" binding:"required"`
}

// GET payments (admin only)
func (ctl *PaymentController) GetPayments(c *gin.Context) {
	payments, err := ctl.service.GetAllPayments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch payments, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// GET all payments made (self-requested by user)
func (ctl *PaymentController) GetUserPayments(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	payments, err := ctl.service.GetPaymentsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch payments, try again later",
			"detail":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, payments)
}

// User create payment via Xendit
func (ctl *PaymentController) CreatePayment(c *gin.Context) {
	var req createPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid order ID",
			"detail":  err.Error(),
		})
		return
	}

	booking, err := ctl.service.GetBookingByID(c.Request.Context(), req.BookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "order not found",
			"detail":  err.Error(),
		})
		return
	}

	// Panggil Xendit API (sandbox)
	xenditKey := os.Getenv("XENDIT_API_KEY")
	if xenditKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Xendit API key not set",
		})
		return
	}
    payerEmail := ""
	if booking.User != nil {
		payerEmail = booking.User.Email
	}
	payload := map[string]interface{}{
		"external_id":          booking.ID.String(),
		"amount":               booking.TotalAmount,
		"payer_email":          payerEmail,
		"description":          fmt.Sprintf("Payment for order %s", booking.ID.String()),
		"success_redirect_url": "https://example.com/success",
		"failure_redirect_url": "https://example.com/failure",
	}

	data, _ := json.Marshal(payload)
	reqXendit, _ := http.NewRequest("POST", "https://api.xendit.co/v2/invoices", bytes.NewBuffer(data))
	reqXendit.SetBasicAuth(xenditKey, "")
	reqXendit.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(reqXendit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to call payment gateway, try again later",
			"detail":  err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Invalid response from payment gateway",
			"detail":  err.Error(),
		})
		return
	}

	invoiceID, ok := result["id"].(string)
	invoiceURL, _ := result["invoice_url"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Invalid invoice id from payment gateway",
		})
		return
	}

	// Save to DB
	payment, err := ctl.service.CreatePayment(c.Request.Context(), booking, invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to save payment",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Payment created",
		"payment_id":  payment.ID,
		"xendit_id":   invoiceID,
		"invoice_url": invoiceURL,
		"status":      payment.Status,
	})
}

// Webhook Xendit
func (ctl *PaymentController) XenditWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid payload",
			"detail":  err.Error(),
		})
		return
	}

	invoiceID, _ := payload["id"].(string)
	status, _ := payload["status"].(string)

	if err := ctl.service.UpdatePaymentStatus(c.Request.Context(), invoiceID, status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update payment",
			"detail":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment successfully updated"})
}
