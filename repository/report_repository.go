package repository

import (
	"context"
	"ticketing/entity"
	"time"

	"gorm.io/gorm"
)

type ReportRepository interface {
	CountTicketsByDateRange(ctx context.Context, start, end time.Time) (int64, error)
	CountTicketsByEvent(ctx context.Context, eventID uint) (int64, error)
	SumTotalAmountByDateRange(ctx context.Context, start, end time.Time) (float64, error)
	SumTotalAmountByEvent(ctx context.Context, eventID uint) (float64, error)
}

type reportRepo struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepo{db: db}
}

// Count tiket by tanggal
func (r *reportRepo) CountTicketsByDateRange(ctx context.Context, start, end time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Tickets{}).
		Where("created_at >= ? AND created_at < ?", start, end).
		Where("ticket_status IN ?", []string{"valid", "used"}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Count tiket by event (liat via booking)
func (r *reportRepo) CountTicketsByEvent(ctx context.Context, eventID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entity.Tickets{}).
		Joins("JOIN bookings ON bookings.id = tickets.booking_id").
		Joins("JOIN ticket_types ON ticket_types.id = tickets.ticket_type_id").
		Where("ticket_types.event_id = ?", eventID).
		Where("tickets.ticket_status IN ?", []string{"valid", "used"}).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Sum total amount by date range (hanya tiket yg statusmya valid/used)
func (r *reportRepo) SumTotalAmountByDateRange(ctx context.Context, start, end time.Time) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).
		Model(&entity.Bookings{}).
		Joins("JOIN tickets ON tickets.booking_id = bookings.id").
		Where("order_date >= ? AND order_date < ?", start, end).
		Where("tickets.ticket_status IN ?", []string{"valid", "used"}).
		Select("SUM(total_amount)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Sum total amount by event
func (r *reportRepo) SumTotalAmountByEvent(ctx context.Context, eventID uint) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).
		Model(&entity.Bookings{}).
		Joins("JOIN tickets ON tickets.booking_id = bookings.id").
		Joins("JOIN ticket_types ON ticket_types.id = tickets.ticket_type_id").
		Where("ticket_types.event_id = ?", eventID).
		Where("tickets.ticket_status IN ?", []string{"valid", "used"}).
		Select("SUM(total_amount)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}