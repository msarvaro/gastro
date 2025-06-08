package service

import (
	"context"
	"fmt"
	"log"
	"restaurant-management/internal/domain/waiter"
)

type WaiterService struct {
	waiterRepo waiter.Repository
}

func NewWaiterService(waiterRepo waiter.Repository) waiter.Service {
	return &WaiterService{
		waiterRepo: waiterRepo,
	}
}

func (s *WaiterService) GetWaiterProfile(ctx context.Context, waiterID int, businessID int) (*waiter.WaiterProfile, error) {
	if waiterID <= 0 {
		return nil, fmt.Errorf("invalid waiter ID: %d", waiterID)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	profile, err := s.waiterRepo.GetWaiterProfile(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error retrieving waiter profile for waiter %d: %v", waiterID, err)
		return nil, err
	}

	return profile, nil
}

func (s *WaiterService) GetWaiterCurrentAndUpcomingShifts(ctx context.Context, waiterID int, businessID int) (*waiter.ShiftWithEmployees, []waiter.ShiftWithEmployees, error) {
	if waiterID <= 0 {
		return nil, nil, fmt.Errorf("invalid waiter ID: %d", waiterID)
	}
	if businessID <= 0 {
		return nil, nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	currentShift, upcomingShifts, err := s.waiterRepo.GetWaiterCurrentAndUpcomingShifts(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error retrieving shifts for waiter %d: %v", waiterID, err)
		return nil, nil, err
	}

	if upcomingShifts == nil {
		upcomingShifts = []waiter.ShiftWithEmployees{}
	}

	return currentShift, upcomingShifts, nil
}

func (s *WaiterService) GetTablesAssignedToWaiter(ctx context.Context, waiterID int, businessID int) ([]waiter.Table, error) {
	if waiterID <= 0 {
		return nil, fmt.Errorf("invalid waiter ID: %d", waiterID)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	tables, err := s.waiterRepo.GetTablesAssignedToWaiter(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error retrieving assigned tables for waiter %d: %v", waiterID, err)
		return nil, err
	}

	if tables == nil {
		tables = []waiter.Table{}
	}

	return tables, nil
}

func (s *WaiterService) GetWaiterOrderStats(ctx context.Context, waiterID int, businessID int) (waiter.OrderStatusCounts, error) {
	if waiterID <= 0 {
		return waiter.OrderStatusCounts{}, fmt.Errorf("invalid waiter ID: %d", waiterID)
	}
	if businessID <= 0 {
		return waiter.OrderStatusCounts{}, fmt.Errorf("invalid business ID: %d", businessID)
	}

	stats, err := s.waiterRepo.GetWaiterOrderStats(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error retrieving order stats for waiter %d: %v", waiterID, err)
		return waiter.OrderStatusCounts{}, err
	}

	return stats, nil
}

func (s *WaiterService) GetWaiterPerformanceMetrics(ctx context.Context, waiterID int, businessID int) (waiter.PerformanceMetrics, error) {
	if waiterID <= 0 {
		return waiter.PerformanceMetrics{}, fmt.Errorf("invalid waiter ID: %d", waiterID)
	}
	if businessID <= 0 {
		return waiter.PerformanceMetrics{}, fmt.Errorf("invalid business ID: %d", businessID)
	}

	metrics, err := s.waiterRepo.GetWaiterPerformanceMetrics(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error retrieving performance metrics for waiter %d: %v", waiterID, err)
		return waiter.PerformanceMetrics{}, err
	}

	return metrics, nil
}
