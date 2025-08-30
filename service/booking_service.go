package service

import (
	"context"
	"fmt"
	"time"

	"ticketing/entity"
	"ticketing/repository"
	"ticketing/helper"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(ctx context.Context, userID uuid.UUID, ticketTypeID uint, qty uint, visitors []entity.Visitors) (*entity.Bookings, error)
	GetBookingByID(ctx context.Context, bookingID uuid.UUID, UserID uuid.UUID, isAdmin bool) (*entity.Bookings, error)
	AutoCancelBookings()
	GetBookings(ctx context.Context, limit, offset int) ([]entity.Bookings, error)
	GetBookingsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Bookings, error)
	UpdateBookingStatus(ctx context.Context, bookingID uuid.UUID, status string) error
	GetBookingsByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Bookings, error)
    CountBookings(ctx context.Context) (int64, error)
    CountBookingsByStatus(ctx context.Context, status string) (int64, error)
}

type bookingService struct {
	ticketTypeRepo repository.TicketTypeRepository
	bookingRepo repository.BookingRepository
	visitorRepo repository.VisitorRepository
	ticketRepo repository.TicketRepository
}

func NewBookingService(
	ticketTypeRepo repository.TicketTypeRepository,
	bookingRepo repository.BookingRepository,
	visitorRepo repository.VisitorRepository,
	ticketRepo repository.TicketRepository,
) BookingService {
	return &bookingService{
		ticketTypeRepo: ticketTypeRepo,
		bookingRepo:    bookingRepo,
		visitorRepo:    visitorRepo,
		ticketRepo:     ticketRepo,
	}
}

func (s *bookingService) CreateBooking(ctx context.Context, userID uuid.UUID, ticketTypeID uint, qty uint, visitors []entity.Visitors) (*entity.Bookings, error) {
    tx := s.bookingRepo.BeginTx()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()

    tt, err := s.ticketTypeRepo.GetByTicketTypeIDForUpdate(ctx, ticketTypeID, tx)
    if err != nil {
        tx.Rollback()
        return nil, fmt.Errorf("failed to fetch ticket type: %w", err)
    }
    if tt.Quota < qty {
        tx.Rollback()
        return nil, fmt.Errorf("not enough quota, only %d left", tt.Quota)
    }
    // block beli kalau ticket type udah finished
    if tt.Status == "finished" {
        tx.Rollback()
        return nil, fmt.Errorf("cannot book ticket: ticket type is finished")
    }
    totalAmount := float64(tt.Price) * float64(qty)
    // create booking
    booking := &entity.Bookings{
        ID:            uuid.New(),
        UserID:        userID,
        Quantity:      qty,
        BookingStatus: "pending",
        TotalAmount:   totalAmount,
        OrderDate:     time.Now(),
        CreatedAt:     time.Now(),
    }
    if err := s.bookingRepo.CreateBooking(ctx, booking, tx); err != nil {
        tx.Rollback()
        return nil, err
    }

    if len(visitors) != int(qty) {
        tx.Rollback()
        return nil, fmt.Errorf("number of visitors must equal to quantity")
    }
    seatNumbers, err := helper.GenerateSeats(ctx, s.ticketRepo, ticketTypeID, int(qty))
	if err != nil {
		tx.Rollback()
		return nil, err
	}
    for i, v := range visitors {
    existing, _ := s.visitorRepo.GetByNationalID(ctx, v.NationalID)
    var visitorID uuid.UUID
    if existing != nil {
        visitorID = existing.ID
    } else {
        v.ID = uuid.New()
        v.BuyerID = userID
        if err := s.visitorRepo.Create(ctx, &v, tx); err != nil {
            tx.Rollback()
            return nil, err
        }
        visitorID = v.ID
    }

    ticket := entity.Tickets{
        ID:           uuid.New(),
        BookingID:    booking.ID,
        TicketTypeID: ticketTypeID, 
        VisitorID:    visitorID,
        SeatNumber:   seatNumbers[i], // pake seat yang udah digenerate
        TicketStatus: "pending",
        CreatedAt:    time.Now(),
    }
    if err := s.ticketRepo.Create(ctx, &ticket, tx); err != nil {
        tx.Rollback()
        return nil, err
    }
}

    // update quota
    tt.Quota -= qty
    if err := s.ticketTypeRepo.UpdateTicketType(ctx, tt, tx); err != nil {
        tx.Rollback()
        return nil, err
    }

    if err := tx.Commit().Error; err != nil {
        return nil, err
    }
     userEmail := ""
        userName := "there"
        if booking.User != nil {
            userEmail = booking.User.Email
            if booking.User.FullName != "" {
                userName = booking.User.FullName
            } else {
                userName = booking.User.Username // fallback kalau gak ada fullname
            }
        }

        subject := "QuickTix - Booking Cancelled"
		body := fmt.Sprintf(
			"Hi %s,\n\n"+
            "Booking %s has been made.\n"+
            "Please make payment within 15 minutes to avoid cancellation\n\n"+ 
            "Our customer support:"+
            "WhatsApp: +62 812 90909090"+
            "Thank you,\n"+
            "QuickTix", 
            userName, booking.ID,
        )
     _ = helper.SendEmail(userEmail, subject, body)
    
    return booking, nil
}
 
func (s *bookingService) GetBookingByID(ctx context.Context, bookingID uuid.UUID, UserID uuid.UUID, isAdmin bool) (*entity.Bookings, error) {
    booking, err := s.bookingRepo.GetBookingByID(ctx, bookingID)
    if err != nil {
        return nil, err
    }

    // user biasa cuma bisa liat booking dirinya sendiri
    if !isAdmin && booking.UserID != UserID {
        return nil, fmt.Errorf("unauthorized: not your booking")
    }

    return booking, nil
}


// auto cancel pending > 15 mins
func (s *bookingService) AutoCancelBookings() {
    ctx := context.Background()
    bookings, _ := s.bookingRepo.GetPendingBookingsOlderThan(ctx, 15*time.Minute)
    for _, booking := range bookings {
        tx, err := s.ticketTypeRepo.BeginTx(ctx)
        if err != nil {
            fmt.Println("failed to start transaction:", err)
            continue
        }

        o, err := s.bookingRepo.GetBookingByIDForUpdate(ctx, booking.ID, tx)
        if err != nil {
            tx.Rollback()
            continue
        }

        // skip kalau udah final
        if o.BookingStatus == "cancelled" || o.BookingStatus == "refunded" {
            tx.Rollback()
            continue
        }

        // update booking
        o.BookingStatus = "cancelled"
        o.UpdatedAt = time.Now()
        if err := s.bookingRepo.Update(ctx, o, tx); err != nil {
            tx.Rollback()
            continue
        }

        // balikin quota tiket
        for _, t := range o.Tickets {
            tt, err := s.ticketTypeRepo.GetByTicketTypeIDForUpdate(ctx, t.TicketTypeID, tx)
            if err != nil {
                tx.Rollback()
                continue
            }
            tt.Quota += 1
            if err := s.ticketTypeRepo.UpdateTicketType(ctx, tt, tx); err != nil {
                tx.Rollback()
                continue
            }
        }
        s.ticketTypeRepo.CommitTx(tx)

        // ambil data user buat email
        userEmail := ""
        userName := "there"
        if o.User != nil {
            userEmail = o.User.Email
            if o.User.FullName != "" {
                userName = o.User.FullName
            } else {
                userName = o.User.Username // fallback kalau gak ada fullname
            }
        }

        subject := "QuickTix - Booking Cancelled"
		body := fmt.Sprintf(
			"Hi %s,\n\n"+
            "Booking %s has been cancelled because it exceeded the 15-minutes payment window.\n"+
            "If you have already made payment, please contact our support:\n\n"+ 
            "WhatsApp: +62 812 90909090"+
            "Thank you,\n"+
            "QuickTix", 
            userName, booking.ID,
        )
        _ = helper.SendEmail(userEmail, subject, body)
    }
}

func (s *bookingService) GetBookings(ctx context.Context, limit, offset int) ([]entity.Bookings, error) {
	return s.bookingRepo.GetBookings(ctx, limit, offset)
}

func (s *bookingService) GetBookingsByUser(ctx context.Context, userID uuid.UUID) ([]entity.Bookings, error) {
	return s.bookingRepo.GetBookingsByUser(ctx, userID)
}

func (s *bookingService) GetBookingsByStatus(ctx context.Context, status string, limit, offset int) ([]entity.Bookings, error) {
	return s.bookingRepo.GetBookingsByStatus(ctx, status, limit, offset)
}

func (s *bookingService) UpdateBookingStatus(ctx context.Context, bookingID uuid.UUID, status string) error {
	tx := s.bookingRepo.BeginTx()
	if tx.Error != nil {
		return tx.Error
	}

	o, err := s.bookingRepo.GetBookingByIDForUpdate(ctx, bookingID, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// validate status baru
	allowed := map[string]bool{
		"pending":   true,
		"paid":      true,
		"cancelled": true,
		"refunded":  true,
	}
	if !allowed[status] {
		tx.Rollback()
		return fmt.Errorf("invalid booking status")
	}

	// kalau udah final, ga bisa diubah lagi
	if o.BookingStatus == "refunded" || o.BookingStatus == "cancelled" {
		tx.Rollback()
		return fmt.Errorf("booking sudah %s, tidak bisa diubah lagi", o.BookingStatus)
	}

	// cek dulu sebelum update 
	if status == "refunded" && o.BookingStatus != "paid" {
		tx.Rollback()
		return fmt.Errorf("only paid bookings can be refunded")
	}

	// update booking
	o.BookingStatus = status
	o.UpdatedAt = time.Now()
	if err := s.bookingRepo.Update(ctx, o, tx); err != nil {
		tx.Rollback()
		return err
	}

	// update ticket status
	ticketStatus := ""
	switch status {
	case "paid":
		ticketStatus = "valid"
	case "cancelled", "refunded":
		ticketStatus = "cancelled"
	case "pending":
		ticketStatus = "pending"
	}
	if ticketStatus != "" {
		if err := s.ticketRepo.BulkUpdateStatusByBooking(ctx, bookingID, ticketStatus, tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	// rollback quota kalau dibatalin/refund
	if status == "cancelled" || status == "refunded" {
		for _, t := range o.Tickets {
			tt, _ := s.ticketTypeRepo.GetByTicketTypeIDForUpdate(ctx, t.TicketTypeID, tx)
			tt.Quota += 1
			if err := s.ticketTypeRepo.UpdateTicketType(ctx, tt, tx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update quota for ticket type %d: %w", t.TicketTypeID, err)
			}
		}
	}

	// kirim email
	userEmail := ""
	userName := "there"
	if o.User != nil {
		userEmail = o.User.Email
		if o.User.FullName != "" {
			userName = o.User.FullName
		} else {
			userName = o.User.Username
		}
	}
	subject := "QuickTix - Booking Refund"
    body := fmt.Sprintf(
        "Hi %s,\n\n"+
            "Booking %s has been refunded.\n"+
            "If you have any questions, please contact our support:\n"+
            "WhatsApp: +62 812 90909090\n\n"+
            "Thank you,\n"+
            "QuickTix", 
        userName, o.ID,
    )
    _ = helper.SendEmail(userEmail, subject, body)

	return tx.Commit().Error
}

func (s *bookingService) CountBookings(ctx context.Context) (int64, error) {
	return s.bookingRepo.CountBookings(ctx)
}

func (s *bookingService) CountBookingsByStatus(ctx context.Context, status string) (int64, error) {
	return s.bookingRepo.CountBookingsByStatus(ctx, status)
}
