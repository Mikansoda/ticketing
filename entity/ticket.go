package entity

import (
	"time"
	"gorm.io/gorm"
)

type TicketTypes struct {
	ID        uint       `gorm:"not null;primaryKey" json:"id"`
	EventID   uint       `gorm:"not null" json:"event_id"`
	Name      string     `gorm:"type:varchar(100);not null" json:"name"` // VIP, Regular, dll
    Status    string     `gorm:"type:enum('active','finished');default:'active'" json:"status"`
	Price     float64    `gorm:"type:decimal(12,2);not null" json:"price"`
	Quota     uint       `gorm:"not null" json:"quota"`
	Date     *time.Time  `gorm:"type:date" json:"date,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Event     Events     `gorm:"foreignKey:EventID" json:"-"`
}
