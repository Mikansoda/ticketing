package repository

import (
	"context"
	"time"

	"ticketing/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookingRepository interface {
	CreateBooking(ctx context.Context, booking *entity.Bookings, tx *gorm.DB) error
	GetBookings(ctx context.Context, limit, offset int) ([]entity.Bookings, error)
	GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Bookings, error)
	GetBookingsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Bookings, error)
	GetBookingsByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Bookings, error)
	GetPendingBookingsOlderThan(ctx context.Context, duration time.Duration) ([]entity.Bookings, error)
	Update(ctx context.Context, booking *entity.Bookings, tx *gorm.DB) error
	GetBookingByIDForUpdate(ctx context.Context, id uuid.UUID, tx *gorm.DB) (*entity.Bookings, error)
	BeginTx() *gorm.DB
}

type bookingRepo struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepo{db: db}
}

func (r *bookingRepo) CreateBooking(ctx context.Context, booking *entity.Bookings, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(booking).Error
}

func (r *bookingRepo) GetBookings(ctx context.Context, limit, offset int) ([]entity.Bookings, error) {
	var bookings []entity.Bookings
	err := r.db.WithContext(ctx).
	    Preload("User").
		Preload("Tickets").
        Preload("Tickets.TicketTypes").
		Limit(limit).Offset(offset).
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepo) GetBookingByID(ctx context.Context, id uuid.UUID) (*entity.Bookings, error) {
	var booking entity.Bookings
	err := r.db.WithContext(ctx).
	    Preload("User").
		Preload("Tickets").
        Preload("Tickets.TicketTypes").
        Preload("Tickets.Visitors").
		First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepo) GetBookingsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Bookings, error) {
	var bookings []entity.Bookings
	err := r.db.WithContext(ctx).
	    Preload("User").
		Preload("Tickets").
        Preload("Tickets.TicketTypes").
		Where("user_id = ?", userID).
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepo) GetBookingsByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Bookings, error) {
	var bookings []entity.Bookings
	if err := r.db.WithContext(ctx).
	    Preload("Tickets").
		Preload("Tickets.TicketTypes").
		Where("booking_status = ?", status).
		Limit(limit).Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *bookingRepo) GetPendingBookingsOlderThan(ctx context.Context, duration time.Duration) ([]entity.Bookings, error) {
	var bookings []entity.Bookings
	cutoff := time.Now().Add(-duration)
	err := r.db.WithContext(ctx).
		Preload("Tickets").
		Where("booking_status = ? AND created_at < ?", "pending", cutoff).
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepo) Update(ctx context.Context, booking *entity.Bookings, tx *gorm.DB) error {
	return tx.WithContext(ctx).Save(booking).Error
}

func (r *bookingRepo) GetBookingByIDForUpdate(ctx context.Context, id uuid.UUID, tx *gorm.DB) (*entity.Bookings, error) {
	var booking entity.Bookings
	err := tx.WithContext(ctx).
		Preload("Tickets").
		Preload("Tickets.TicketTypes").
		Preload("Tickets.Visitors").
		Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).
		First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepo) BeginTx() *gorm.DB {
	return r.db.Begin()
}
