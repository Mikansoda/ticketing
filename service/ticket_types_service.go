package service

import (
	"context"
	"errors"
	"time"

	"ticketing/entity"
	"ticketing/repository"
)

type TicketTypeService interface {
	CreateTicketType(ctx context.Context, t *entity.TicketTypes) error
	GetTicketTypeByID(ctx context.Context, id uint) (*entity.TicketTypes, error)
	GetTicketTypeByIDIncludeDeleted(ctx context.Context, id uint) (*entity.TicketTypes, error)
	GetTicketTypes(ctx context.Context, event string, limit, offset int) ([]entity.TicketTypes, error)
	UpdateTicketType(ctx context.Context, t *entity.TicketTypes) (*entity.TicketTypes, error)
	DeleteTicketType(ctx context.Context, id uint) error
	RecoverTicketType(ctx context.Context, id uint) error
	SyncTicketTypeStatus(ctx context.Context, eventID uint) error
}

type ticketTypeService struct {
	eventRepo       repository.EventRepository
	ticketTypesRepo repository.TicketTypeRepository
}

func NewTicketTypeService(ticketTypesRepo repository.TicketTypeRepository, eventRepo repository.EventRepository) TicketTypeService {
	return &ticketTypeService{ticketTypesRepo: ticketTypesRepo, eventRepo: eventRepo}
}

func (s *ticketTypeService) CreateTicketType(ctx context.Context, t *entity.TicketTypes) error {
	// ambil event
	event, err := s.eventRepo.GetEventsByID(ctx, t.EventID)
	if err != nil {
		return errors.New("event not found")
	}

	// hitung total quota ticket type yg udah ada
	existingTypes, err := s.ticketTypesRepo.GetTicketTypes(ctx, "", 0, 0)
	if err != nil {
		return errors.New("failed to fetch existing ticket types: " + err.Error())
	}

	totalQuota := uint(0)
	for _, tt := range existingTypes {
		if tt.EventID == t.EventID {
			totalQuota += tt.Quota
		}
	}

	// validasi totalQuota + quotaBaru lebih byk/sama kuk capacity event
	if totalQuota+t.Quota > uint(event.Capacity) {
		return errors.New("ticket quota exceeds event capacity")
	}

	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	if err := s.ticketTypesRepo.CreateTicketType(ctx, t); err != nil {
		return errors.New("failed to create ticket type: " + err.Error())
	}
	return nil
}

func (s *ticketTypeService) GetTicketTypeByID(ctx context.Context, id uint) (*entity.TicketTypes, error) {
	ticketType, err := s.ticketTypesRepo.GetByTicketTypeID(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch ticket type: " + err.Error())
	}
	if ticketType == nil {
		return nil, errors.New("ticket type not found")
	}
	return ticketType, nil
}

func (s *ticketTypeService) GetTicketTypeByIDIncludeDeleted(ctx context.Context, id uint) (*entity.TicketTypes, error) {
	ticketType, err := s.ticketTypesRepo.GetTicketTypeByIDIncludeDeleted(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch ticket type: " + err.Error())
	}
	if ticketType == nil {
		return nil, errors.New("ticket type not found")
	}
	return ticketType, nil
}

func (s *ticketTypeService) GetTicketTypes(ctx context.Context, event string, limit, offset int) ([]entity.TicketTypes, error) {
	ticketTypes, err := s.ticketTypesRepo.GetTicketTypes(ctx, event, limit, offset)
	if err != nil {
		return nil, errors.New("failed to fetch ticket types: " + err.Error())
	}
	return ticketTypes, nil
}

func (s *ticketTypeService) UpdateTicketType(ctx context.Context, t *entity.TicketTypes) (*entity.TicketTypes, error) {
	existing, err := s.ticketTypesRepo.GetTicketTypeByIDIncludeDeleted(ctx, t.ID)
	if err != nil {
		return nil, errors.New("failed to fetch ticket type: " + err.Error())
	}
	if existing == nil {
		return nil, errors.New("ticket type not found")
	}

	// 1. cek booking paid
	bookings, err := s.ticketTypesRepo.GetBookingsByTicketTypeID(ctx, existing.ID)
	if err != nil {
		return nil, errors.New("failed to fetch bookings: " + err.Error())
	}
	for _, b := range bookings {
		if b.BookingStatus == "paid" {
			return nil, errors.New("ticket type already has paid bookings, cannot be updated")
		}
	}

	// 2. ambil status event
	event, err := s.eventRepo.GetEventsByID(ctx, existing.EventID)
	if err != nil {
		return nil, errors.New("failed to fetch event: " + err.Error())
	}

	// 3. kalau event udah finished, ticket type otomatis finished
	if event.EventStatus == "finished" {
		existing.Status = "finished"
	} else if t.Status != "" {
		if t.Status != "active" && t.Status != "finished" {
			return nil, errors.New("invalid ticket type status")
		}
		existing.Status = t.Status
	}

	// 4. update field lain
	if t.Name != "" {
		existing.Name = t.Name
	}
	if t.Price > 0 {
		existing.Price = t.Price
	}
	if t.Quota > 0 {
		existing.Quota = t.Quota
	}
	if t.Date != nil {
		existing.Date = t.Date
	}
	if t.Quota > 0 {
		// hitung total quota semua ticket type di event (kecuali yang ini)
		ticketTypes, err := s.ticketTypesRepo.GetTicketTypes(ctx, "", 0, 0)
		if err != nil {
			return nil, errors.New("failed to fetch ticket types: " + err.Error())
		}

		totalQuota := uint(0)
		for _, tt := range ticketTypes {
			if tt.EventID == existing.EventID && tt.ID != existing.ID {
				totalQuota += tt.Quota
			}
		}

		// check kalau quota baru + quota lain > capacity
		if totalQuota+t.Quota > uint(event.Capacity) {
			return nil, errors.New("ticket quota exceeds event capacity")
		}

		existing.Quota = t.Quota
	}

	existing.UpdatedAt = time.Now()

	if err := s.ticketTypesRepo.UpdateTicketType(ctx, existing); err != nil {
		return nil, errors.New("failed to update ticket type: " + err.Error())
	}

	return existing, nil
}

func (s *ticketTypeService) DeleteTicketType(ctx context.Context, id uint) error {
	existing, err := s.ticketTypesRepo.GetByTicketTypeID(ctx, id)
	if err != nil {
		return errors.New("failed to fetch ticket type: " + err.Error())
	}
	if existing == nil {
		return errors.New("ticket type not found")
	}

	// cek apakah ada booking yang udah paid
	bookings, err := s.ticketTypesRepo.GetBookingsByTicketTypeID(ctx, existing.ID)
	if err != nil {
		return errors.New("failed to fetch bookings: " + err.Error())
	}
	for _, b := range bookings {
		if b.BookingStatus == "pending" || b.BookingStatus == "paid" {
			return errors.New("cannot delete ticket type: there are pending and paid bookings")
		}
	}

	if err := s.ticketTypesRepo.DeleteTicketType(ctx, id); err != nil {
		return errors.New("failed to delete ticket type: " + err.Error())
	}
	return nil
}

func (s *ticketTypeService) RecoverTicketType(ctx context.Context, id uint) error {
	existing, err := s.ticketTypesRepo.GetTicketTypeByIDIncludeDeleted(ctx, id)
	if err != nil {
		return errors.New("failed to fetch ticket type: " + err.Error())
	}
	if existing == nil {
		return errors.New("ticket type not found")
	}
	if err := s.ticketTypesRepo.RecoverTicketType(ctx, id); err != nil {
		return errors.New("failed to recover ticket type: " + err.Error())
	}
	return nil
}

func (s *ticketTypeService) SyncTicketTypeStatus(ctx context.Context, eventID uint) error {
	tx, err := s.ticketTypesRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = s.ticketTypesRepo.RollbackTx(tx)
		}
	}()

	var ticketTypes []entity.TicketTypes
	ticketTypes, err = s.ticketTypesRepo.GetTicketTypesByEventIDForUpdate(ctx, eventID, tx)
	if err != nil {
		_ = s.ticketTypesRepo.RollbackTx(tx)
		return err
	}

	for i := range ticketTypes {
		if ticketTypes[i].Status != "finished" {
			ticketTypes[i].Status = "finished"
			if err := s.ticketTypesRepo.UpdateTicketType(ctx, &ticketTypes[i], tx); err != nil {
				_ = s.ticketTypesRepo.RollbackTx(tx)
				return err
			}
		}
	}

	if err := s.ticketTypesRepo.CommitTx(tx); err != nil {
		return err
	}
	return nil
}
