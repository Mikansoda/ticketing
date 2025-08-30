package entity

import (
	"time"
    "github.com/google/uuid"
	"gorm.io/gorm"
)

// Ticket = 1 tiket untuk 1 penumpang (visitor)
type Tickets struct {
    ID            uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
    BookingID     uuid.UUID `gorm:"type:char(36);not null" json:"booking_id"`
    TicketTypeID  uint      `gorm:"not null" json:"ticket_type_id"`
    VisitorID     uuid.UUID `gorm:"type:char(36);not null" json:"visitor_id"`
    SeatNumber    string    `gorm:"type:varchar(50)" json:"seat_number"`
    TicketStatus  string    `gorm:"type:enum('pending','valid','cancelled','used');default:'pending';index" json:"status"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
    CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`

    // Relations
    Bookings     *Bookings    `gorm:"foreignKey:BookingID" json:"-"`
    TicketTypes  *TicketTypes `gorm:"foreignKey:TicketTypeID" json:"ticket_type"`
    Visitors     *Visitors    `gorm:"foreignKey:VisitorID" json:"visitor"`
}

// Booking = satu transaksi pembelian tiket
type Bookings struct {
    ID            uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
    UserID        uuid.UUID `gorm:"type:char(36);not null" json:"user_id"`
    Quantity      uint      `gorm:"not null" json:"quantity"`
    BookingStatus string    `gorm:"type:enum('pending','paid','cancelled','refunded');default:'pending';index" json:"status"`
    TotalAmount   float64   `gorm:"type:decimal(12,2);not null" json:"total_amount"`
	OrderDate     time.Time `gorm:"autoCreateTime" json:"order_date"`
    CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

    // Relations
    User         *Users     `gorm:"foreignKey:UserID" json:"-"`
    Tickets     []Tickets   `gorm:"foreignKey:BookingID" json:"tickets,omitempty"`
}

func (t *Tickets) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (b *Bookings) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}