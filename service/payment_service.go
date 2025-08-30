package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ticketing/entity"
	"ticketing/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, booking *entity.Bookings, invoiceID string) (*entity.Payments, error)
	GetBookingByID(ctx context.Context, bookingID uuid.UUID) (*entity.Bookings, error)
	GetPaymentsByUserID(ctx context.Context, userID string) ([]entity.Payments, error)
	GetAllPayments(ctx context.Context) ([]entity.Payments, error)
	AutoCancelPendingPayments()
	UpdatePaymentStatus(ctx context.Context, invoiceID string, status string) error
}

type paymentService struct {
	paymentRepo   repository.PaymentRepository
	bookingRepo   repository.BookingRepository
	ticketRepo    repository.TicketRepository
	ticketTypeSvc TicketTypeService
	db            *gorm.DB
}

func NewPaymentService(paymentRepo repository.PaymentRepository, bookingRepo repository.BookingRepository, ticketRepo   repository.TicketRepository, ticketTypeSvc TicketTypeService, db *gorm.DB) PaymentService {
	return &paymentService{paymentRepo, bookingRepo, ticketRepo, ticketTypeSvc, db}
}

// Create payment from Xendit invoice
func (s *paymentService) CreatePayment(ctx context.Context, booking *entity.Bookings, invoiceID string) (*entity.Payments, error) {
	if booking == nil {
		return nil, errors.New("booking is nil")
	}

	payment := &entity.Payments{
		ID:          uuid.New(),
		BookingID:   booking.ID,
		InvoiceID:   invoiceID,
		PaymentType: "xendit_invoice",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *paymentService) GetBookingByID(ctx context.Context, bookingID uuid.UUID) (*entity.Bookings, error) {
	booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	return booking, nil
}

func (s *paymentService) GetPaymentsByUserID(ctx context.Context, userID string) ([]entity.Payments, error) {
	var payments []entity.Payments
	err := s.paymentRepo.GetByUserID(ctx, userID, &payments)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

func (s *paymentService) GetAllPayments(ctx context.Context) ([]entity.Payments, error) {
	var payments []entity.Payments
	err := s.paymentRepo.GetPayments(ctx, &payments)
	if err != nil {
		return nil, err
	}
	return payments, nil
}

// Auto cancel pending payment > 15 minutes
func (s *paymentService) AutoCancelPendingPayments() {
	ctx := context.Background()
	payments, _ := s.paymentRepo.GetPendingOlderThan(ctx, 15)
	for _, p := range payments {
		tx := s.db.Begin()
		if tx.Error != nil {
			continue
		}
		// update payment
		p.Status = "failed"
		p.UpdatedAt = time.Now()
		if err := s.paymentRepo.Update(ctx, &p); err != nil {
			tx.Rollback()
			continue
		}
		// cancel booking
		booking, err := s.bookingRepo.GetBookingByIDForUpdate(ctx, p.BookingID, tx)
		if err != nil {
			tx.Rollback()
			continue
		}
		booking.BookingStatus = "cancelled"
		booking.UpdatedAt = time.Now()
		if err := s.bookingRepo.Update(ctx, booking, tx); err != nil {
			tx.Rollback()
			continue
		}

		if len(booking.Tickets) > 0 {
			firstTicket := booking.Tickets[0] 
			prod, err := s.ticketTypeSvc.GetTicketTypeByID(ctx, firstTicket.TicketTypeID)
			if err == nil {
				prod.Quota += booking.Quantity
				s.ticketTypeSvc.UpdateTicketType(ctx, prod)
			}
		}

		tx.Commit()
	}
}

// Update payment status from Xendit webhook
func (s *paymentService) UpdatePaymentStatus(ctx context.Context, invoiceID string, status string) error {
	payment, err := s.paymentRepo.GetByInvoiceID(ctx, invoiceID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	payment.Status = status
	payment.UpdatedAt = time.Now()
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	booking, err := s.bookingRepo.GetBookingByIDForUpdate(ctx, payment.BookingID, tx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("booking not found: %w", err)
	}

	switch status {
	case "PAID", "paid":
		booking.BookingStatus = "paid"
		booking.UpdatedAt = time.Now()
		if err := s.bookingRepo.Update(ctx, booking, tx); err != nil {
			tx.Rollback()
			return err
		}
	if err := s.ticketRepo.BulkUpdateStatusByBooking(ctx, booking.ID, "valid", tx); err != nil {
			tx.Rollback()
			return err
	}

	case "FAILED", "failed":
		booking.BookingStatus = "cancelled"
		booking.UpdatedAt = time.Now()
		if err := s.bookingRepo.Update(ctx, booking, tx); err != nil {
			tx.Rollback()
			return err
	}
	if err := s.ticketRepo.BulkUpdateStatusByBooking(ctx, booking.ID, "cancelled", tx); err != nil {
			tx.Rollback()
			return err
		}
	}
	// rollback kuota tiket
	if len(booking.Tickets) > 0 {
		firstTicket := booking.Tickets[0]
		prod, err := s.ticketTypeSvc.GetTicketTypeByID(ctx, firstTicket.TicketTypeID)
		if err == nil {
			prod.Quota += booking.Quantity
			s.ticketTypeSvc.UpdateTicketType(ctx, prod)
		}
	}
	return tx.Commit().Error
}