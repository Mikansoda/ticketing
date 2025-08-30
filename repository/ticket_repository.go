package repository

import (
	"context"
	"ticketing/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TicketRepository interface {
	Create(ctx context.Context, t *entity.Tickets, tx *gorm.DB) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tickets, error)
	UpdateStatus(ctx context.Context, ticketID uuid.UUID, status string) error
	BulkUpdateStatusByBooking(ctx context.Context, bookingID uuid.UUID, status string, tx *gorm.DB) error
	GetByTicketTypeID(ctx context.Context, ticketTypeID uint) ([]entity.Tickets, error)
}

type ticketRepo struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepo{db: db}
}

func (r *ticketRepo) Create(ctx context.Context, t *entity.Tickets, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(t).Error
}

func (r *ticketRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tickets, error) {
	var ticket entity.Tickets
	if err := r.db.WithContext(ctx).
		Preload("TicketTypes").
		Preload("Visitors").
		Preload("Bookings").
		First(&ticket, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepo) UpdateStatus(ctx context.Context, ticketID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Tickets{}).
		Where("id = ?", ticketID).
		Update("ticket_status", status).Error
}

func (r *ticketRepo) BulkUpdateStatusByBooking(ctx context.Context, bookingID uuid.UUID, status string, tx *gorm.DB) error {
	return tx.WithContext(ctx).
		Model(&entity.Tickets{}).
		Where("booking_id = ?", bookingID).
		Update("ticket_status", status).Error
}

func (r *ticketRepo) GetByTicketTypeID(ctx context.Context, ticketTypeID uint) ([]entity.Tickets, error) {
    var tickets []entity.Tickets
    if err := r.db.WithContext(ctx).
        Where("ticket_type_id = ?", ticketTypeID).
        Find(&tickets).Error; err != nil {
        return nil, err
    }
    return tickets, nil
}
