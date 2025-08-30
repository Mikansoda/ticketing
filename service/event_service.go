package service

import (
	"context"
	"errors"
	"time"

	"ticketing/entity"
	"ticketing/repository"
)

type EventService interface {
	CreateEvents(ctx context.Context, p *entity.Events) error
	GetEventsByID(ctx context.Context, id uint) (*entity.Events, error)
	GetEventsByIDIncludeDeleted(ctx context.Context, id uint) (*entity.Events, error)
	GetEvents(ctx context.Context, search, category, status string, filterDate *time.Time, limit, offset int) ([]entity.Events, error)
	UpdateEvents(ctx context.Context, p *entity.Events) (*entity.Events, error)
	DeleteEvents(ctx context.Context, id uint) error
	RecoverEvents(ctx context.Context, id uint) error
}

type eventService struct {
	repo              repository.EventRepository
	ticketTypeService TicketTypeService
}

func NewEventService(repo repository.EventRepository, ticketTypeService TicketTypeService) EventService {
	return &eventService{repo: repo, ticketTypeService: ticketTypeService}
}

func (s *eventService) CreateEvents(ctx context.Context, p *entity.Events) error {
	exists, err := s.repo.IsEventNameExists(ctx, p.Name, 0)
	if err != nil {
		return errors.New("failed to validate event name: " + err.Error())
	}
	if exists {
		return errors.New("event name already exists")
	}

	p.EventStatus = "active"
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	if err := s.repo.CreateEvents(ctx, p); err != nil {
		return errors.New("failed to create event: " + err.Error())
	}
	return nil
}

func (s *eventService) GetEventsByID(ctx context.Context, id uint) (*entity.Events, error) {
	event, err := s.repo.GetEventsByID(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch event: " + err.Error())
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (s *eventService) GetEventsByIDIncludeDeleted(ctx context.Context, id uint) (*entity.Events, error) {
	event, err := s.repo.GetEventsByIDIncludeDeleted(ctx, id)
	if err != nil {
		return nil, errors.New("failed to fetch events: " + err.Error())
	}
	if event == nil {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (s *eventService) GetEvents(ctx context.Context, search, category, status string, filterDate *time.Time, limit, offset int) ([]entity.Events, error) {
	events, err := s.repo.GetEvents(ctx, search, category, status, filterDate, limit, offset)
	if err != nil {
		return nil, errors.New("failed to fetch events: " + err.Error())
	}
	return events, nil
}

func (s *eventService) UpdateEvents(ctx context.Context, e *entity.Events) (*entity.Events, error) {
	existing, err := s.repo.GetEventsByIDIncludeDeleted(ctx, e.ID)
	if err != nil {
		return nil, errors.New("failed to fetch event: " + err.Error())
	}
	if existing == nil {
		return nil, errors.New("event not found")
	}

	hasBooking, err := s.repo.HasBookings(ctx, e.ID)
	if err != nil {
		return nil, errors.New("failed to validate bookings: " + err.Error())
	}

	// Kalau event udah finished, block update kecuali status diubah
	if existing.EventStatus == "finished" {
		return nil, errors.New("cannot update a finished event")
	}

	// Kalau ada booking
	if hasBooking {
		if e.EventStatus == "ongoing" || e.EventStatus == "finished" {
			existing.EventStatus = e.EventStatus
			existing.UpdatedAt = time.Now()
			if err := s.repo.UpdateEvents(ctx, existing); err != nil {
				return nil, errors.New("failed to update event: " + err.Error())
			}
			if e.EventStatus == "finished" {
				if err := s.ticketTypeService.SyncTicketTypeStatus(ctx, existing.ID); err != nil {
					return nil, errors.New("failed to sync ticket type status: " + err.Error())
				}
			}
			return existing, nil
		}
		return nil, errors.New("cannot update other fields, only allowed to set ongoing or finished")
	}

	// cek nama unik
	if e.Name != "" && e.Name != existing.Name {
		exists, err := s.repo.IsEventNameExists(ctx, e.Name, e.ID)
		if err != nil {
			return nil, errors.New("failed to validate event name: " + err.Error())
		}
		if exists {
			return nil, errors.New("event name already exists")
		}
		existing.Name = e.Name
	}

	if e.CategoryID != 0 {
		existing.CategoryID = e.CategoryID
	}
	if e.Capacity > 0 {
		existing.Capacity = e.Capacity
	}
	if e.EventStatus != "" {
		existing.EventStatus = e.EventStatus
	}
	if e.Description != "" {
		existing.Description = e.Description
	}
	if e.City != "" {
		existing.City = e.City
	}
	if e.Country != "" {
		existing.Country = e.Country
	}
	if !e.StartDate.IsZero() {
		existing.StartDate = e.StartDate
	}
	if !e.EndDate.IsZero() {
		existing.EndDate = e.EndDate
	}

	existing.UpdatedAt = time.Now()

	if err := s.repo.UpdateEvents(ctx, existing); err != nil {
		return nil, errors.New("failed to update event: " + err.Error())
	}
	if existing.EventStatus == "finished" {
		if err := s.ticketTypeService.SyncTicketTypeStatus(ctx, existing.ID); err != nil {
			return nil, errors.New("failed to sync ticket type status: " + err.Error())
		}
	}
	return existing, nil
}

func (s *eventService) DeleteEvents(ctx context.Context, id uint) error {
	existing, err := s.repo.GetEventsByID(ctx, id)
	if err != nil {
		return errors.New("failed to fetch event: " + err.Error())
	}
	if existing == nil {
		return errors.New("event not found")
	}

	// cek udah mulai
	if existing.EventStatus == "ongoing" || existing.EventStatus == "finished" {
		return errors.New("cannot delete event that is ongoing or finished")
	}

	// cek ada booking
	hasBooking, err := s.repo.HasBookings(ctx, id)
	if err != nil {
		return errors.New("failed to validate bookings: " + err.Error())
	}
	if hasBooking {
		return errors.New("cannot delete event that already has bookings")
	}

	if err := s.repo.DeleteEvents(ctx, id); err != nil {
		return errors.New("failed to delete event: " + err.Error())
	}
	return nil
}

func (s *eventService) RecoverEvents(ctx context.Context, id uint) error {
	existing, err := s.repo.GetEventsByIDIncludeDeleted(ctx, id)
	if err != nil {
		return errors.New("failed to fetch event: " + err.Error())
	}
	if existing == nil {
		return errors.New("event not found")
	}
	if err := s.repo.RecoverEvents(ctx, id); err != nil {
		return errors.New("failed to recover event: " + err.Error())
	}
	return nil
}
