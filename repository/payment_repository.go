package repository

import (
	"context"
	"fmt"
	"ticketing/entity"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payments) error
	GetByBookingID(ctx context.Context, orderID uuid.UUID) (*entity.Payments, error)
	GetPendingOlderThan(ctx context.Context, hours int) ([]entity.Payments, error)
	GetByUserID(ctx context.Context, userID string, out *[]entity.Payments) error
	GetPayments(ctx context.Context, out *[]entity.Payments) error
	GetByInvoiceID(ctx context.Context, invoiceID string) (*entity.Payments, error)
	Update(ctx context.Context, payment *entity.Payments) error
}

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, payment *entity.Payments) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepo) GetByBookingID(ctx context.Context, orderID uuid.UUID) (*entity.Payments, error) {
	var payment entity.Payments
	err := r.db.WithContext(ctx).Where("booking_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepo) GetPendingOlderThan(ctx context.Context, minutes int) ([]entity.Payments, error) {
	var payments []entity.Payments
	cutoff := time.Now().Add(-time.Duration(minutes) * time.Minute)
	err := r.db.WithContext(ctx).Where("status = ? AND created_at < ?", "pending", cutoff).Find(&payments).Error
	return payments, err
}

func (r *paymentRepo) GetByUserID(ctx context.Context, userID string, out *[]entity.Payments) error {
	return r.db.WithContext(ctx).
		Joins("JOIN bookings ON bookings.id = payments.booking_id").
		Where("bookings.user_id = ?", userID).
		Order("payments.created_at DESC").
		Find(out).Error
}

func (r *paymentRepo) GetPayments(ctx context.Context, out *[]entity.Payments) error {
	return r.db.WithContext(ctx).Find(out).Error
}

func (r *paymentRepo) GetByInvoiceID(ctx context.Context, invoiceID string) (*entity.Payments, error) {
	fmt.Println("Searching payment with invoice_id:", invoiceID)
	var payment entity.Payments
	err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepo) Update(ctx context.Context, payment *entity.Payments) error {
	return r.db.WithContext(ctx).Save(payment).Error
}
