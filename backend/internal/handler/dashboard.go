package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/business"
	"restaurant-management/internal/domain/inventory"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/domain/table"
	"restaurant-management/internal/domain/user"
	"restaurant-management/internal/middleware"
)

type DashboardController struct {
	businessService  business.Service
	userService      user.Service
	orderService     order.Service
	tableService     table.Service
	inventoryService inventory.Service
}

func NewDashboardController(
	businessService business.Service,
	userService user.Service,
	orderService order.Service,
	tableService table.Service,
	inventoryService inventory.Service,
) *DashboardController {
	return &DashboardController{
		businessService:  businessService,
		userService:      userService,
		orderService:     orderService,
		tableService:     tableService,
		inventoryService: inventoryService,
	}
}

// GetOverviewStats provides general statistics for dashboard
func (c *DashboardController) GetOverviewStats(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	// Get order stats
	orderStats, err := c.orderService.GetOrderStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting order stats: %v", err)
		orderStats = nil // Continue with other stats
	}

	// Get table stats
	tableStats, err := c.tableService.GetTableStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting table stats: %v", err)
		tableStats = nil // Continue with other stats
	}

	// Get inventory count (simple count)
	inventory, err := c.inventoryService.GetAllInventory(r.Context(), businessID)
	inventoryCount := 0
	lowStockCount := 0
	if err == nil {
		inventoryCount = len(inventory)
		for _, item := range inventory {
			if item.Quantity <= item.MinQuantity {
				lowStockCount++
			}
		}
	}

	// Get user count for this business
	users, err := c.userService.GetUsers(r.Context(), businessID)
	userCount := 0
	if err == nil {
		userCount = len(users)
	}

	response := map[string]interface{}{
		"orders": orderStats,
		"tables": tableStats,
		"inventory": map[string]int{
			"total":     inventoryCount,
			"low_stock": lowStockCount,
		},
		"users": map[string]int{
			"total": userCount,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBusinessOverview provides business-level overview (admin only)
func (c *DashboardController) GetBusinessOverview(w http.ResponseWriter, r *http.Request) {
	// Get all businesses stats
	businesses, stats, err := c.businessService.GetAllBusinesses(r.Context())
	if err != nil {
		log.Printf("Error getting business overview: %v", err)
		http.Error(w, "Failed to fetch business overview", http.StatusInternalServerError)
		return
	}

	// Get user stats
	userStats, err := c.userService.GetUserStats(r.Context())
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		userStats = nil // Continue with other stats
	}

	response := map[string]interface{}{
		"businesses": map[string]interface{}{
			"list":  businesses,
			"stats": stats,
		},
		"users": userStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
