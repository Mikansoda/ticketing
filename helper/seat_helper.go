package helper

import (
	"context"
	"fmt"
	"ticketing/repository"
)

// GenerateSeats otomatis assign seat number buat tiket baru
func GenerateSeats(ctx context.Context, ticketRepo repository.TicketRepository, ticketTypeID uint, qty int) ([]string, error) {
	// ambil semua tiket yg udah ada buat ticket type ini
	existingTickets, err := ticketRepo.GetByTicketTypeID(ctx, ticketTypeID)
	if err != nil {
		return nil, err
	}

	// bikin map seat yg udh dipake
	usedSeats := map[string]bool{}
	for _, t := range existingTickets {
		usedSeats[t.SeatNumber] = true
	}

	// contoh seat: A1, A2,dst, B1, B2, dst
	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := []string{}
	letterIdx := 0
	num := 1

	for len(result) < qty {
		seat := fmt.Sprintf("%c%d", letters[letterIdx], num)
		if !usedSeats[seat] {
			result = append(result, seat)
		}
		num++
		if num > 100 { // ini pake max 100 seat per row
			num = 1
			letterIdx++
			if letterIdx >= len(letters) {
				return nil, fmt.Errorf("ran out of seats")
			}
		}
	}

	return result, nil
}
