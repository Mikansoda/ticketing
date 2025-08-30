package repository

import (
	"context"

	"errors"
	"ticketing/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TicketTypeRepository interface {
	CreateTicketType(ctx context.Context, p *entity.TicketTypes, tx ...*gorm.DB) error
	GetTicketTypes(ctx context.Context, event string, limit, offset int) ([]entity.TicketTypes, error) 
	GetByTicketTypeID(ctx context.Context, id uint) (*entity.TicketTypes, error)
	GetTicketTypeByIDIncludeDeleted(ctx context.Context, id uint) (*entity.TicketTypes, error)
	UpdateTicketType(ctx context.Context, p *entity.TicketTypes, tx ...*gorm.DB) error
	DeleteTicketType(ctx context.Context, id uint) error
	RecoverTicketType(ctx context.Context, id uint) error
	BeginTx(ctx context.Context) (*gorm.DB, error)
	CommitTx(tx *gorm.DB) error
	RollbackTx(tx *gorm.DB) error
	GetByTicketTypeIDForUpdate(ctx context.Context, id uint, tx *gorm.DB) (*entity.TicketTypes, error)
	GetTicketTypesByEventIDForUpdate(ctx context.Context, eventID uint, tx *gorm.DB) ([]entity.TicketTypes, error)
	GetBookingsByTicketTypeID(ctx context.Context, ticketTypeID uint) ([]entity.Bookings, error)
}

type ticketTypeRepo struct {
	db *gorm.DB
}

func NewTicketTypeRepository(db *gorm.DB) TicketTypeRepository {
	return &ticketTypeRepo{db: db}
}

func (r *ticketTypeRepo) CreateTicketType(ctx context.Context, p *entity.TicketTypes, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(p).Error
}

func (r *ticketTypeRepo) GetTicketTypes(ctx context.Context, event string, limit, offset int) ([]entity.TicketTypes, error) {
	var ticketTypes []entity.TicketTypes
	query := r.db.WithContext(ctx).
		Model(&entity.TicketTypes{}).
		Preload("Event")

	if event != "" {
		query = query.Joins("JOIN events ON events.id = ticket_types.event_id").
			Where("events.name LIKE ?", "%"+event+"%")
	}

	if err := query.Limit(limit).Offset(offset).Find(&ticketTypes).Error; err != nil {
		return nil, err
	}
	return ticketTypes, nil
}

func (r *ticketTypeRepo) GetByTicketTypeID(ctx context.Context, id uint) (*entity.TicketTypes, error) {
	var ticketType entity.TicketTypes
	if err := r.db.WithContext(ctx).
		Preload("Event").
		First(&ticketType, id).Error; err != nil {
		return nil, err
	}
	return &ticketType, nil
}

func (r *ticketTypeRepo) GetTicketTypeByIDIncludeDeleted(ctx context.Context, id uint) (*entity.TicketTypes, error) {
	var ticketType entity.TicketTypes
	if err := r.db.WithContext(ctx).
		Unscoped().
		Preload("Event").
		First(&ticketType, id).Error; err != nil {
		return nil, err
	}
	return &ticketType, nil
}

func (r *ticketTypeRepo) UpdateTicketType(ctx context.Context, ticketType *entity.TicketTypes, tx ...*gorm.DB) error {
	db := r.db
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Save(ticketType).Error
}

func (r *ticketTypeRepo) DeleteTicketType(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.TicketTypes{}, id).Error
}

func (r *ticketTypeRepo) RecoverTicketType(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.TicketTypes{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *ticketTypeRepo) BeginTx(ctx context.Context) (*gorm.DB, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

func (r *ticketTypeRepo) CommitTx(tx *gorm.DB) error {
	if tx == nil {
		return errors.New("no active transaction")
	}
	return tx.Commit().Error
}

func (r *ticketTypeRepo) RollbackTx(tx *gorm.DB) error {
	if tx == nil {
		return errors.New("no active transaction")
	}
	return tx.Rollback().Error
}

func (r *ticketTypeRepo) GetByTicketTypeIDForUpdate(ctx context.Context, id uint, tx *gorm.DB) (*entity.TicketTypes, error) {
	var ticketType entity.TicketTypes
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ticketType, id).Error; err != nil {
		return nil, err
	}
	return &ticketType, nil
}

func (r *ticketTypeRepo) GetTicketTypesByEventIDForUpdate(ctx context.Context, eventID uint, tx *gorm.DB) ([]entity.TicketTypes, error) {
	var ticketTypes []entity.TicketTypes
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("event_id = ?", eventID).Find(&ticketTypes).Error; err != nil {
		return nil, err
	}
	return ticketTypes, nil
}

func (r *ticketTypeRepo) GetBookingsByTicketTypeID(ctx context.Context, ticketTypeID uint) ([]entity.Bookings, error) {
	var bookings []entity.Bookings

	err := r.db.WithContext(ctx).
	    Distinct("bookings.*").
		Joins("JOIN tickets ON tickets.booking_id = bookings.id").
		Where("tickets.ticket_type_id = ?", ticketTypeID).
		Preload("Tickets", "ticket_type_id = ?", ticketTypeID).
		Find(&bookings).Error

	if err != nil {
		return nil, err
	}

	return bookings, nil
}
