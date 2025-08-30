package entity

import (
	"time"
	"github.com/google/uuid"
    "gorm.io/gorm"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Entity table for users
type Users struct {
    ID                uuid.UUID     `gorm:"type:char(36);primaryKey" json:"id"`
    FullName          string        `gorm:"type:varchar(100);not null" json:"full_name"`
    Username          string        `gorm:"type:varchar(100);unique;not null" json:"username"`
    Email             string        `gorm:"type:varchar(100);unique;not null" json:"email"`
    PasswordHash      string        `gorm:"type:text;not null" json:"-"`
    Role              Role          `gorm:"type:enum('user','admin');default:'user';not null" json:"role"`
    IsActive          bool          `gorm:"default:false"`
    OTPHash           string        `gorm:"size:191"` // hash of last OTP
	OTPExpiresAt     *time.Time
    RefreshTokenHash  string        `gorm:"size:191"`
	RefreshExpiresAt *time.Time
    CreatedAt         time.Time     `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt         time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

    // Relations
	Visitors         []Visitors       `gorm:"foreignKey:BuyerID"`
}

// Entity table for user adresses
type Visitors struct {
    ID                uuid.UUID     `gorm:"type:char(36);primaryKey" json:"id"`
    BuyerID           uuid.UUID     `gorm:"type:char(36);not null" json:"buyer_id"`
	Title             string        `gorm:"type:enum('mr','mrs','ms');not null" json:"title"`
    FullName          string        `gorm:"type:varchar(100);not null" json:"visitor_name"`
    PhoneNumber       string        `gorm:"type:varchar(100);not null" json:"phone_number"`
    Nationality       string        `gorm:"type:varchar(100);not null" json:"nationality"`
    NationalID        string        `gorm:"type:varchar(500);uniqueIndex;not null" json:"national_id"`
    CreatedAt         time.Time     `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt         time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
    DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`  // for soft deletion

    // Relation
    User             *Users         `gorm:"foreignKey:BuyerID" json:"-"` 
}

func (u *Users) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *Visitors) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}