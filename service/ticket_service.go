package service

import (
	"context"
	"errors"
	"time"

	"ticketing/entity"
	"ticketing/repository"

	"github.com/google/uuid"
)

type TicketService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tickets, error)
	UpdateStatus(ctx context.Context, ticketID uuid.UUID, status string) (*entity.Tickets, error)
	BulkUpdateStatusByBooking(ctx context.Context, bookingID uuid.UUID, status string) error
}

type ticketService struct {
	repo repository.TicketRepository
}

func NewTicketService(repo repository.TicketRepository) TicketService {
	return &ticketService{repo: repo}
}

func (s *ticketService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tickets, error) {
	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch ticket: " + err.Error())
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}
	return ticket, nil
}

func (s *ticketService) UpdateStatus(ctx context.Context, ticketID uuid.UUID, status string) (*entity.Tickets, error) {
	allowed := map[string]bool{
		"pending":   true,
		"valid":     true,
		"cancelled": true,
		"used":      true,
	}
	if !allowed[status] {
		return nil, errors.New("invalid ticket status")
	}

	ticket, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("failed to fetch ticket: " + err.Error())
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}
	if ticket.TicketStatus == "used" || ticket.TicketStatus == "cancelled" {
		return nil, errors.New("ticket already final and cannot be updated")
	}
	if status == "used" {
		// tiket harus masih valid
		if ticket.TicketStatus != "valid" {
			return nil, errors.New("only valid tickets can be used")
		}
		// booking harus paid
		if ticket.Bookings == nil || ticket.Bookings.BookingStatus != "paid" {
			return nil, errors.New("ticket's booking must be paid before using")
		}
		now := time.Now()
		ticket.CheckedInAt = &now
	}
	if err := s.repo.UpdateStatus(ctx, ticketID, status); err != nil {
		return nil, errors.New("failed to update ticket status: " + err.Error())
	}
	ticket.TicketStatus = status

	return ticket, nil
}

func (s *ticketService) BulkUpdateStatusByBooking(ctx context.Context, bookingID uuid.UUID, status string) error {
	if err := s.repo.BulkUpdateStatusByBooking(ctx, bookingID, status, nil); err != nil {
		return errors.New("failed to bulk update tickets: " + err.Error())
	}
	return nil
}
