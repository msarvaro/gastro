package repository

import (
	"context"
	"database/sql"
	"restaurant-management/internal/domain/entity"
	"time"
)

type WaiterRepository struct {
	db *sql.DB
}

func NewWaiterRepository(db *sql.DB) *WaiterRepository {
	return &WaiterRepository{db: db}
}

// Stats operations
func (r *WaiterRepository) GetStats(ctx context.Context, userID int, period string, date time.Time) (*entity.WaiterStats, error) {
	var query string
	var dateFilter string
	
	switch period {
	case "daily":
		dateFilter = "DATE(orders.created_at) = DATE($2)"
	case "weekly":
		dateFilter = "DATE(orders.created_at) >= DATE($2 - INTERVAL '7 days') AND DATE(orders.created_at) <= DATE($2)"
	case "monthly":
		dateFilter = "EXTRACT(YEAR FROM orders.created_at) = EXTRACT(YEAR FROM $2) AND EXTRACT(MONTH FROM orders.created_at) = EXTRACT(MONTH FROM $2)"
	default:
		dateFilter = "DATE(orders.created_at) = DATE($2)"
	}
	
	query = `
		SELECT 
			COUNT(orders.id) as total_orders,
			COALESCE(SUM(orders.total_amount), 0) as total_revenue,
			COALESCE(AVG(orders.total_amount), 0) as average_order_value,
			0 as total_tips,
			0 as customer_rating,
			0 as response_time
		FROM orders 
		WHERE orders.waiter_id = $1 AND ` + dateFilter + ` AND orders.status = 'completed'
	`
	
	stats := &entity.WaiterStats{
		UserID: userID,
		Period: period,
		Date:   date,
	}
	
	err := r.db.QueryRowContext(ctx, query, userID, date).Scan(
		&stats.TotalOrders,
		&stats.TotalRevenue,
		&stats.AverageOrderValue,
		&stats.TotalTips,
		&stats.CustomerRating,
		&stats.ResponseTime,
	)
	
	return stats, err
}

func (r *WaiterRepository) GetStatsByBusinessID(ctx context.Context, businessID int, period string, date time.Time) ([]*entity.WaiterStats, error) {
	var query string
	var dateFilter string
	
	switch period {
	case "daily":
		dateFilter = "DATE(orders.created_at) = DATE($2)"
	case "weekly":
		dateFilter = "DATE(orders.created_at) >= DATE($2 - INTERVAL '7 days') AND DATE(orders.created_at) <= DATE($2)"
	case "monthly":
		dateFilter = "EXTRACT(YEAR FROM orders.created_at) = EXTRACT(YEAR FROM $2) AND EXTRACT(MONTH FROM orders.created_at) = EXTRACT(MONTH FROM $2)"
	default:
		dateFilter = "DATE(orders.created_at) = DATE($2)"
	}
	
	query = `
		SELECT 
			orders.waiter_id as user_id,
			COUNT(orders.id) as total_orders,
			COALESCE(SUM(orders.total_amount), 0) as total_revenue,
			COALESCE(AVG(orders.total_amount), 0) as average_order_value,
			0 as total_tips,
			0 as customer_rating,
			0 as response_time
		FROM orders 
		JOIN users ON orders.waiter_id = users.id
		WHERE users.business_id = $1 AND ` + dateFilter + ` AND orders.status = 'completed'
		GROUP BY orders.waiter_id
	`
	
	rows, err := r.db.QueryContext(ctx, query, businessID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var statsList []*entity.WaiterStats
	for rows.Next() {
		stats := &entity.WaiterStats{
			Period: period,
			Date:   date,
		}
		
		err := rows.Scan(
			&stats.UserID,
			&stats.TotalOrders,
			&stats.TotalRevenue,
			&stats.AverageOrderValue,
			&stats.TotalTips,
			&stats.CustomerRating,
			&stats.ResponseTime,
		)
		if err != nil {
			return nil, err
		}
		
		statsList = append(statsList, stats)
	}
	
	return statsList, nil
}

func (r *WaiterRepository) CreateOrUpdateStats(ctx context.Context, stats *entity.WaiterStats) error {
	query := `
		INSERT INTO waiter_stats (user_id, total_orders, total_revenue, average_order_value,
		total_tips, customer_rating, response_time, period, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, period, date)
		DO UPDATE SET
			total_orders = EXCLUDED.total_orders,
			total_revenue = EXCLUDED.total_revenue,
			average_order_value = EXCLUDED.average_order_value,
			total_tips = EXCLUDED.total_tips,
			customer_rating = EXCLUDED.customer_rating,
			response_time = EXCLUDED.response_time
	`
	
	_, err := r.db.ExecContext(ctx, query,
		stats.UserID, stats.TotalOrders, stats.TotalRevenue, stats.AverageOrderValue,
		stats.TotalTips, stats.CustomerRating, stats.ResponseTime, stats.Period, stats.Date,
	)
	
	return err
}

// Performance operations
func (r *WaiterRepository) GetPerformance(ctx context.Context, userID int, shiftID int) (*entity.WaiterPerformance, error) {
	query := `
		SELECT user_id, shift_id, tables_served, orders_completed, revenue, tips,
		complaints, compliments, avg_service_time, date
		FROM waiter_performance
		WHERE user_id = $1 AND shift_id = $2
	`
	
	performance := &entity.WaiterPerformance{}
	err := r.db.QueryRowContext(ctx, query, userID, shiftID).Scan(
		&performance.UserID, &performance.ShiftID, &performance.TablesServed,
		&performance.OrdersCompleted, &performance.Revenue, &performance.Tips,
		&performance.Complaints, &performance.Compliments, &performance.AvgServiceTime,
		&performance.Date,
	)
	
	return performance, err
}

func (r *WaiterRepository) GetPerformanceByDateRange(ctx context.Context, userID int, start, end time.Time) ([]*entity.WaiterPerformance, error) {
	query := `
		SELECT user_id, shift_id, tables_served, orders_completed, revenue, tips,
		complaints, compliments, avg_service_time, date
		FROM waiter_performance
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var performances []*entity.WaiterPerformance
	for rows.Next() {
		performance := &entity.WaiterPerformance{}
		err := rows.Scan(
			&performance.UserID, &performance.ShiftID, &performance.TablesServed,
			&performance.OrdersCompleted, &performance.Revenue, &performance.Tips,
			&performance.Complaints, &performance.Compliments, &performance.AvgServiceTime,
			&performance.Date,
		)
		if err != nil {
			return nil, err
		}
		performances = append(performances, performance)
	}
	
	return performances, nil
}

func (r *WaiterRepository) CreatePerformanceRecord(ctx context.Context, performance *entity.WaiterPerformance) error {
	query := `
		INSERT INTO waiter_performance (user_id, shift_id, tables_served, orders_completed,
		revenue, tips, complaints, compliments, avg_service_time, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		performance.UserID, performance.ShiftID, performance.TablesServed,
		performance.OrdersCompleted, performance.Revenue, performance.Tips,
		performance.Complaints, performance.Compliments, performance.AvgServiceTime,
		performance.Date,
	)
	
	return err
}

func (r *WaiterRepository) UpdatePerformanceRecord(ctx context.Context, performance *entity.WaiterPerformance) error {
	query := `
		UPDATE waiter_performance SET
		tables_served = $1, orders_completed = $2, revenue = $3, tips = $4,
		complaints = $5, compliments = $6, avg_service_time = $7
		WHERE user_id = $8 AND shift_id = $9
	`
	
	_, err := r.db.ExecContext(ctx, query,
		performance.TablesServed, performance.OrdersCompleted, performance.Revenue,
		performance.Tips, performance.Complaints, performance.Compliments,
		performance.AvgServiceTime, performance.UserID, performance.ShiftID,
	)
	
	return err
} 