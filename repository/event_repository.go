package repository

import (
	"context"
	"time"

	"ticketing/entity"

	"gorm.io/gorm"
)

type EventRepository interface {
	CreateEvents(ctx context.Context, e *entity.Events) error
	GetEventsByID(ctx context.Context, id uint) (*entity.Events, error)
	GetEventsByIDIncludeDeleted(ctx context.Context, id uint) (*entity.Events, error)
	GetEvents(ctx context.Context, search, category, status string, filterDate *time.Time, limit, offset int) ([]entity.Events, error)
	UpdateEvents(ctx context.Context, e *entity.Events) error
	DeleteEvents(ctx context.Context, id uint) error
	RecoverEvents(ctx context.Context, id uint) error
	IsEventNameExists(ctx context.Context, name string, excludeID uint) (bool, error)
	HasBookings(ctx context.Context, eventID uint) (bool, error)
	CountEvents(ctx context.Context, search, category, status string, filterDate *time.Time) (int64, error)
}

type eventRepo struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepo{db: db}
}

func (r *eventRepo) CreateEvents(ctx context.Context, e *entity.Events) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *eventRepo) GetEventsByID(ctx context.Context, id uint) (*entity.Events, error) {
	var e entity.Events
	if err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Category").
		Preload("TicketTypes").
		First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepo) GetEventsByIDIncludeDeleted(ctx context.Context, id uint) (*entity.Events, error) {
	var e entity.Events
	if err := r.db.WithContext(ctx).
		Unscoped().
		Preload("Images").
		Preload("Category").
		Preload("TicketTypes").
		First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *eventRepo) GetEvents(ctx context.Context, search, category, status string, filterDate *time.Time, limit, offset int) ([]entity.Events, error) {
	var events []entity.Events
	db := r.db.WithContext(ctx).
		Model(&entity.Events{}).
		Preload("Images").
		Preload("Category")

	if search != "" {
		db = db.Where("events.name LIKE ?", "%"+search+"%")
	}
	if category != "" {
		db = db.Joins("JOIN event_categories ec ON ec.id = events.category_id").
			Where("ec.name LIKE ?", "%"+category+"%")
	}
	if status != "" {
		db = db.Where("events.event_status = ?", status)
	}
	if filterDate != nil {
		db = db.Where("events.start_date <= ? AND events.end_date >= ?", *filterDate, *filterDate)
	}

	if err := db.Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepo) UpdateEvents(ctx context.Context, e *entity.Events) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *eventRepo) DeleteEvents(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Events{}, id).Error
}

func (r *eventRepo) RecoverEvents(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.Events{}).
		Unscoped().
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

func (r *eventRepo) IsEventNameExists(ctx context.Context, name string, excludeID uint) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&entity.Events{}).Where("name = ?", name)
	if excludeID != 0 {
		db = db.Where("id <> ?", excludeID)
	}
	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *eventRepo) HasBookings(ctx context.Context, eventID uint) (bool, error) {
	var count int64
	// booking join tickets join ticket_types untuk cek ada transaksi/ga
	err := r.db.WithContext(ctx).
		Table("bookings b").
		Joins("JOIN tickets t ON t.booking_id = b.id").
		Joins("JOIN ticket_types tt ON tt.id = t.ticket_type_id").
		Where("tt.event_id = ?", eventID).
		Where("b.booking_status IN ?", []string{"paid", "pending"}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *eventRepo) CountEvents(ctx context.Context, search, category, status string, filterDate *time.Time) (int64, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&entity.Events{})

	if search != "" {
		db = db.Where("events.name LIKE ?", "%"+search+"%")
	}
	if category != "" {
		db = db.Joins("JOIN event_categories ec ON ec.id = events.category_id").
			Where("ec.name LIKE ?", "%"+category+"%")
	}
	if status != "" {
		db = db.Where("events.event_status = ?", status)
	}
	if filterDate != nil {
		db = db.Where("events.start_date <= ? AND events.end_date >= ?", *filterDate, *filterDate)
	}

	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

