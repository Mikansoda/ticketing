package entity

import (
	"time"

	"gorm.io/gorm"
)

type Events struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	CategoryID  uint           `gorm:"not null" json:"category_id"`
	Capacity    uint           `gorm:"not null" json:"capacity"`
	EventStatus string         `gorm:"type:enum('active','ongoing','finished');default:'active'" json:"event_status"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	City        string         `gorm:"type:varchar(255);not null" json:"city"`
	Country     string         `gorm:"type:varchar(255);not null" json:"country"`
	StartDate   time.Time      `gorm:"not null" json:"start_date"`
	EndDate     time.Time      `gorm:"not null" json:"end_date"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // soft delete

	// Foreign key
	Category    EventCategories `gorm:"foreignKey:CategoryID" json:"category"`

	// Relations
	Images      []EventImages   `gorm:"foreignKey:EventID" json:"images,omitempty"`
	TicketTypes []TicketTypes   `gorm:"foreignKey:EventID" json:"ticket_types,omitempty"`
}

// Entity table for event images
type EventImages struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	EventID     uint           `gorm:"not null" json:"event_id"`
	ImageURL    string         `gorm:"type:text;not null" json:"image_url"`
	IsPrimary   bool           `gorm:"default:false" json:"is_primary"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Event      *Events         `gorm:"foreignKey:EventID" json:"-"`
}

// Entity table for event categories
type EventCategories struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Events    []Events       `gorm:"foreignKey:CategoryID" json:"events,omitempty"`
}
