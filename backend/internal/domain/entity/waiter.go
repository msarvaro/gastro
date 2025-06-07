package entity

import "time"

// WaiterStats represents aggregated statistics for a waiter
type WaiterStats struct {
	UserID            int
	TotalOrders       int
	TotalRevenue      float64
	AverageOrderValue float64
	TotalTips         float64
	CustomerRating    float64 // Average rating from customers
	ResponseTime      float64 // Average response time in minutes
	Period            string  // daily, weekly, monthly
	Date              time.Time
}

// WaiterPerformance represents performance metrics
type WaiterPerformance struct {
	UserID          int
	ShiftID         int
	TablesServed    int
	OrdersCompleted int
	Revenue         float64
	Tips            float64
	Complaints      int
	Compliments     int
	AvgServiceTime  float64 // in minutes
	Date            time.Time
}

// Business methods
func (w *WaiterStats) CalculateEfficiency() float64 {
	if w.TotalOrders == 0 {
		return 0
	}
	// Simple efficiency calculation based on revenue per order and response time
	revenueScore := w.AverageOrderValue / 100 * 50   // 50% weight
	responseScore := (20 - w.ResponseTime) / 20 * 50 // 50% weight, assuming 20 min is baseline

	efficiency := revenueScore + responseScore
	if efficiency < 0 {
		return 0
	}
	if efficiency > 100 {
		return 100
	}
	return efficiency
}
