package entity

import (
	"time"
	"github.com/google/uuid"
)

// Entity table for payments
type Payments struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	BookingID   uuid.UUID `gorm:"type:char(36);not null" json:"booking_id"`
	InvoiceID   string    `gorm:"type:varchar(255);uniqueIndex" json:"xendit_invoice_id"`
	PaymentType string    `gorm:"type:enum('xendit_invoice');not null" json:"payment_type"`
	Status      string    `gorm:"type:enum('pending','paid','failed');default:'pending'" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Bookings   *Bookings  `gorm:"foreignKey:BookingID" json:"booking_ID,omitempty"`
}
