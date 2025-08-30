package service

import (
	"context"
	"errors"
	"time"

	"ticketing/repository"
)

type ReportService interface {
	// laporan ringkasan bulanan/summary
	GetMonthlySummary(ctx context.Context, month, year int) (totalTickets int64, totalRevenue float64, err error)
	// laporan tiket per event
	GetTicketsByEvent(ctx context.Context, eventID uint) (totalTickets int64, totalRevenue float64, err error)
}

type reportService struct {
	reportRepo  repository.ReportRepository
}

func NewReportService(reportRepo  repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

// GetMonthlySummary hitung total tiket dan revenue di bulan tertentu
func (s *reportService) GetMonthlySummary(ctx context.Context, month, year int) (int64, float64, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0) // awal bulan berikutnya

	totalTickets, err := s.reportRepo.CountTicketsByDateRange(ctx, startDate, endDate)
	if err != nil {
		return 0, 0, errors.New("failed to count tickets: " + err.Error())
	}

	totalRevenue, err := s.reportRepo.SumTotalAmountByDateRange(ctx, startDate, endDate)
	if err != nil {
		return 0, 0, errors.New("failed to sum revenue: " + err.Error())
	}

	return totalTickets, totalRevenue, nil
}

// GetTicketsByEvent ngehitung tiket dan revenue untuk event tertentu
func (s *reportService) GetTicketsByEvent(ctx context.Context, eventID uint) (int64, float64, error) {
	totalTickets, err := s.reportRepo.CountTicketsByEvent(ctx, eventID)
	if err != nil {
		return 0, 0, errors.New("failed to count tickets: " + err.Error())
	}

	totalRevenue, err := s.reportRepo.SumTotalAmountByEvent(ctx, eventID)
	if err != nil {
		return 0, 0, errors.New("failed to sum revenue: " + err.Error())
	}

	return totalTickets, totalRevenue, nil
}
