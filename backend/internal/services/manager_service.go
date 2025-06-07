package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// ManagerService implements manager service operations
type ManagerService struct {
	userRepo     repository.UserRepository
	orderRepo    repository.OrderRepository
	businessRepo repository.BusinessRepository
}

func createManagerService(
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	businessRepo repository.BusinessRepository,
) *ManagerService {
	return &ManagerService{
		userRepo:     userRepo,
		orderRepo:    orderRepo,
		businessRepo: businessRepo,
	}
}

// GetStaffList retrieves all staff members for a business
func (s *ManagerService) GetStaffList(ctx context.Context, businessID int) ([]*entity.User, error) {
	return s.userRepo.GetByBusinessID(ctx, businessID)
}

// CreateStaffMember creates a new staff member
func (s *ManagerService) CreateStaffMember(ctx context.Context, user *entity.User) error {
	return s.userRepo.Create(ctx, user)
}

// UpdateStaffMember updates an existing staff member
func (s *ManagerService) UpdateStaffMember(ctx context.Context, user *entity.User) error {
	return s.userRepo.Update(ctx, user)
}

// DeactivateStaffMember deactivates a staff member
func (s *ManagerService) DeactivateStaffMember(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Status = "inactive"
	return s.userRepo.Update(ctx, user)
}

// GetDailyReport retrieves daily report for a business
func (s *ManagerService) GetDailyReport(ctx context.Context, businessID int, date time.Time) (interface{}, error) {
	// Get orders for the specific date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	orders, err := s.orderRepo.GetByDateRange(ctx, businessID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// Calculate basic statistics
	totalOrders := len(orders)
	var totalRevenue float64
	var completedOrders int

	for _, order := range orders {
		if order.IsCompleted() {
			completedOrders++
			totalRevenue += float64(order.TotalAmount)
		}
	}

	report := map[string]interface{}{
		"date":             date.Format("2006-01-02"),
		"total_orders":     totalOrders,
		"completed_orders": completedOrders,
		"total_revenue":    totalRevenue,
		"average_order": func() float64 {
			if completedOrders > 0 {
				return totalRevenue / float64(completedOrders)
			}
			return 0
		}(),
	}

	return report, nil
}

// GetRevenueReport retrieves revenue report for a date range
func (s *ManagerService) GetRevenueReport(ctx context.Context, businessID int, start, end time.Time) (interface{}, error) {
	orders, err := s.orderRepo.GetByDateRange(ctx, businessID, start, end)
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var completedOrders int
	dailyRevenue := make(map[string]float64)

	for _, order := range orders {
		if order.IsCompleted() {
			completedOrders++
			revenue := float64(order.TotalAmount)
			totalRevenue += revenue

			// Group by day
			dayKey := order.CreatedAt.Format("2006-01-02")
			dailyRevenue[dayKey] += revenue
		}
	}

	report := map[string]interface{}{
		"period":           map[string]string{"start": start.Format("2006-01-02"), "end": end.Format("2006-01-02")},
		"total_revenue":    totalRevenue,
		"completed_orders": completedOrders,
		"daily_breakdown":  dailyRevenue,
		"average_daily": func() float64 {
			days := int(end.Sub(start).Hours() / 24)
			if days > 0 {
				return totalRevenue / float64(days)
			}
			return 0
		}(),
	}

	return report, nil
}

// GetStaffPerformanceReport retrieves staff performance report
func (s *ManagerService) GetStaffPerformanceReport(ctx context.Context, businessID int, period string) (interface{}, error) {
	// Get all staff for the business
	staff, err := s.userRepo.GetByBusinessID(ctx, businessID)
	if err != nil {
		return nil, err
	}

	performance := make([]map[string]interface{}, 0)

	for _, user := range staff {
		if user.Role == "waiter" {
			// Get waiter statistics
			orderStats, _ := s.orderRepo.GetWaiterOrderStatistics(ctx, user.ID)
			tablesServed, _ := s.orderRepo.GetWaiterTablesServedCount(ctx, user.ID)
			ordersCompleted, _ := s.orderRepo.GetWaiterCompletedOrdersCount(ctx, user.ID)

			userPerformance := map[string]interface{}{
				"user_id":          user.ID,
				"name":             user.Name,
				"role":             user.Role,
				"tables_served":    tablesServed,
				"orders_completed": ordersCompleted,
				"order_stats":      orderStats,
			}

			performance = append(performance, userPerformance)
		}
	}

	report := map[string]interface{}{
		"period":            period,
		"business_id":       businessID,
		"staff_performance": performance,
	}

	return report, nil
}

// UpdateBusinessHours updates business operating hours
func (s *ManagerService) UpdateBusinessHours(ctx context.Context, businessID int, openTime, closeTime string) error {
	// This would require extending the business entity to include operating hours
	// For now, we'll just verify the business exists
	_, err := s.businessRepo.GetByID(ctx, businessID)
	return err
}

// GetBusinessStatistics retrieves business statistics
func (s *ManagerService) GetBusinessStatistics(ctx context.Context, businessID int) (interface{}, error) {
	// Get staff count
	staff, err := s.userRepo.GetByBusinessID(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Get recent orders (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentOrders, err := s.orderRepo.GetByDateRange(ctx, businessID, thirtyDaysAgo, time.Now())
	if err != nil {
		return nil, err
	}

	var totalRevenue float64
	var completedOrders int

	for _, order := range recentOrders {
		if order.IsCompleted() {
			completedOrders++
			totalRevenue += float64(order.TotalAmount)
		}
	}

	stats := map[string]interface{}{
		"staff_count":          len(staff),
		"total_orders_30d":     len(recentOrders),
		"completed_orders_30d": completedOrders,
		"revenue_30d":          totalRevenue,
		"average_order_30d": func() float64 {
			if completedOrders > 0 {
				return totalRevenue / float64(completedOrders)
			}
			return 0
		}(),
	}

	return stats, nil
}
