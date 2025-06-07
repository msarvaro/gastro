package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"restaurant-management/internal/presentation/http/middleware"
	"restaurant-management/internal/presentation/http/utils"
	"strconv"
	"time"
)

// WaiterHandler handles waiter-related requests
type WaiterHandler struct {
	waiterService  services.WaiterService
	tableService   services.TableService
	orderService   services.OrderService
	profileService services.UserService
}

// NewWaiterHandler creates a new waiter handler
func NewWaiterHandler(
	waiterService services.WaiterService,
	tableService services.TableService,
	orderService services.OrderService,
	profileService services.UserService,
) *WaiterHandler {
	return &WaiterHandler{
		waiterService:  waiterService,
		tableService:   tableService,
		orderService:   orderService,
		profileService: profileService,
	}
}

// GetTables returns all tables assigned to the waiter
func (h *WaiterHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	waiterID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	// Get all tables for the business
	tables, err := h.tableService.GetTablesByBusinessID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting tables: %v", err)
		http.Error(w, "Failed to retrieve tables", http.StatusInternalServerError)
		return
	}

	// Get assigned tables for the waiter
	assignedTables, err := h.waiterService.GetAssignedTables(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting assigned tables: %v", err)
		// Continue with all tables if we can't get assigned tables
	}

	// Create a map of assigned table IDs for quick lookup
	assignedTableIDs := make(map[int]bool)
	for _, table := range assignedTables {
		assignedTableIDs[table.ID] = true
	}

	// Prepare the response
	var response responses.TablesListResponse

	// Count stats and calculate revenue
	stats := responses.TablesStatsResponse{}
	var totalRevenue float64

	// Map tables to response
	for _, table := range tables {
		// Count stats
		stats.Total++
		switch table.Status {
		case "free":
			stats.Free++
		case "occupied":
			stats.Occupied++
		case "reserved":
			stats.Reserved++
		}

		// Create enhanced table response
		tableResponse := mapTableToEnhancedResponse(table)

		// Add orders if any and calculate revenue
		activeOrders, err := h.orderService.GetOrdersByTable(r.Context(), table.ID)
		if err == nil && len(activeOrders) > 0 {
			var tableRevenue float64
			for _, order := range activeOrders {
				if order.Status != "completed" && order.Status != "cancelled" {
					tableResponse.Orders = append(tableResponse.Orders, &responses.OrderShortResponse{
						ID:     order.ID,
						Status: utils.FormatStatus(order.Status, "order"),
						Time:   utils.FormatDate(order.CreatedAt),
					})
				}
				if order.Status == "completed" {
					tableRevenue += float64(order.TotalAmount)
				}
			}
			if tableRevenue > 0 {
				revenueResponse := utils.FormatMoney(int(tableRevenue * 100)) // Convert to cents
				tableResponse.Revenue = &revenueResponse
				totalRevenue += tableRevenue
			}
		}

		// Add available actions based on table status
		tableResponse.AvailableActions = utils.CreateAvailableActions("table", table.Status, "waiter")

		response.Tables = append(response.Tables, tableResponse)
	}

	// Calculate final stats
	if stats.Total > 0 {
		stats.OccupancyRate = float64(stats.Occupied) / float64(stats.Total) * 100
	}
	if len(tables) > 0 {
		stats.AverageRevenue = utils.FormatMoney(int(totalRevenue / float64(len(tables)) * 100)) // Convert to cents
	}
	stats.TurnoverRate = "2.3 оборота/день" // TODO: Calculate from actual data

	response.Stats = stats
	response.LastUpdated = utils.FormatDate(time.Now())

	// Add filter options
	response.Filters = []responses.FilterOptionResponse{
		{Value: "all", Label: "Все статусы", Count: stats.Total, Selected: true},
		{Value: "free", Label: "Свободные", Count: stats.Free, Selected: false},
		{Value: "occupied", Label: "Занятые", Count: stats.Occupied, Selected: false},
		{Value: "reserved", Label: "Забронированные", Count: stats.Reserved, Selected: false},
	}

	// Add available actions
	response.Actions = []responses.ActionResponse{
		{ID: "refresh", Label: "Обновить", Variant: "primary"},
		{ID: "assign_tables", Label: "Назначить столы", Variant: "secondary"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateTableStatus updates the status of a table
func (h *WaiterHandler) UpdateTableStatus(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract table ID from URL
	tableIDStr := r.URL.Path[len("/api/waiter/tables/"):]
	tableIDStr = tableIDStr[:len(tableIDStr)-len("/status")]
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		http.Error(w, "Invalid table ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req requests.UpdateTableStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update table status
	err = h.tableService.UpdateTableStatus(r.Context(), tableID, req.Status)
	if err != nil {
		log.Printf("Error updating table status: %v", err)
		http.Error(w, "Failed to update table status", http.StatusInternalServerError)
		return
	}

	// Get updated table
	table, err := h.tableService.GetTableByID(r.Context(), tableID)
	if err != nil {
		log.Printf("Error getting updated table: %v", err)
		http.Error(w, "Table status updated but couldn't fetch details", http.StatusInternalServerError)
		return
	}

	response := mapTableToEnhancedResponse(table)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrders returns all orders for the waiter
func (h *WaiterHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	waiterID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all active orders for the waiter
	orders, err := h.waiterService.GetWaiterOrders(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting waiter orders: %v", err)
		http.Error(w, "Failed to retrieve orders", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	var response responses.OrdersListResponse

	// Count stats and calculate revenue
	stats := responses.OrdersStatsResponse{}
	var totalRevenue float64
	var totalOrders int

	// Map orders to response
	for _, order := range orders {
		// Count stats by status
		stats.TotalActiveOrders++
		totalOrders++
		totalRevenue += float64(order.TotalAmount)

		switch order.Status {
		case "new":
			stats.New++
		case "accepted":
			stats.Accepted++
		case "preparing":
			stats.Preparing++
		case "ready":
			stats.Ready++
		case "served":
			stats.Served++
		}

		response.Orders = append(response.Orders, mapOrderToEnhancedResponse(order))
	}

	// Calculate final stats
	stats.TotalRevenue = utils.FormatMoney(int(totalRevenue * 100)) // Convert to cents
	if totalOrders > 0 {
		stats.AverageOrderValue = utils.FormatMoney(int(totalRevenue / float64(totalOrders) * 100)) // Convert to cents
	}
	stats.PendingTime = "Средне ожидание: 12 мин" // TODO: Calculate from actual data

	response.Stats = stats
	response.LastUpdated = utils.FormatDate(time.Now())

	// Add filter options
	response.Filters = []responses.FilterOptionResponse{
		{Value: "all", Label: "Все заказы", Count: stats.TotalActiveOrders, Selected: true},
		{Value: "new", Label: "Новые", Count: stats.New, Selected: false},
		{Value: "preparing", Label: "Готовятся", Count: stats.Preparing, Selected: false},
		{Value: "ready", Label: "Готовы", Count: stats.Ready, Selected: false},
	}

	// Add available actions
	response.Actions = []responses.ActionResponse{
		{ID: "refresh", Label: "Обновить", Variant: "primary"},
		{ID: "create_order", Label: "Создать заказ", Variant: "primary"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateOrder creates a new order
func (h *WaiterHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	waiterID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req requests.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create order entity
	order := &entity.Order{
		TableID:   req.TableID,
		WaiterID:  waiterID,
		Status:    "new",
		Comment:   req.Comment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add items to order
	for _, item := range req.Items {
		order.Items = append(order.Items, &entity.OrderItem{
			DishID:   item.DishID,
			Quantity: item.Quantity,
			Notes:    item.Notes,
		})
	}

	// Create order
	err := h.waiterService.TakeOrder(r.Context(), waiterID, order)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	// Get created order
	createdOrder, err := h.orderService.GetOrderByID(r.Context(), order.ID)
	if err != nil {
		log.Printf("Error getting created order: %v", err)
		http.Error(w, "Order created but couldn't fetch details", http.StatusInternalServerError)
		return
	}

	response := mapOrderToEnhancedResponse(createdOrder)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateOrderStatus updates the status of an order
func (h *WaiterHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract order ID from URL
	orderIDStr := r.URL.Path[len("/api/waiter/orders/"):]
	orderIDStr = orderIDStr[:len(orderIDStr)-len("/status")]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req requests.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update order status
	err = h.orderService.UpdateOrderStatus(r.Context(), orderID, req.Status)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	// Get updated order
	updatedOrder, err := h.orderService.GetOrderByID(r.Context(), orderID)
	if err != nil {
		log.Printf("Error getting updated order: %v", err)
		http.Error(w, "Order status updated but couldn't fetch details", http.StatusInternalServerError)
		return
	}

	response := mapOrderToEnhancedResponse(updatedOrder)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrderHistory returns the order history for the waiter
func (h *WaiterHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	waiterID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all orders for the waiter, then filter completed and cancelled
	allOrders, err := h.waiterService.GetWaiterOrders(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting waiter orders: %v", err)
		http.Error(w, "Failed to retrieve order history", http.StatusInternalServerError)
		return
	}

	// Filter completed and cancelled orders
	var completedOrders []*entity.Order
	var cancelledOrders []*entity.Order

	for _, order := range allOrders {
		if order.Status == "completed" {
			completedOrders = append(completedOrders, order)
		} else if order.Status == "cancelled" {
			cancelledOrders = append(cancelledOrders, order)
		}
	}

	// Prepare the response
	var response responses.OrderHistoryListResponse

	// Count stats with enhanced calculations
	stats := responses.HistoryStatsResponse{
		CompletedTotal: len(completedOrders),
		CancelledTotal: len(cancelledOrders),
	}

	var completedAmountTotal, cancelledAmountTotal float64
	var totalServiceTime time.Duration
	var serviceTimeCount int

	// Calculate total amounts and service times for completed orders
	for _, order := range completedOrders {
		completedAmountTotal += float64(order.TotalAmount)
		response.Orders = append(response.Orders, mapOrderToHistoryResponse(order))

		// Calculate service time if completion time is available
		if order.CompletedAt != nil {
			serviceTime := order.CompletedAt.Sub(order.CreatedAt)
			totalServiceTime += serviceTime
			serviceTimeCount++
		}
	}

	// Calculate cancelled orders
	for _, order := range cancelledOrders {
		cancelledAmountTotal += float64(order.TotalAmount)
		response.Orders = append(response.Orders, mapOrderToHistoryResponse(order))
	}

	// Set enhanced stats
	stats.CompletedAmountTotal = utils.FormatMoney(int(completedAmountTotal * 100)) // Convert to cents
	stats.CancelledAmountTotal = utils.FormatMoney(int(cancelledAmountTotal * 100)) // Convert to cents

	// Calculate success rate
	totalOrders := len(completedOrders) + len(cancelledOrders)
	if totalOrders > 0 {
		stats.SuccessRate = float64(len(completedOrders)) / float64(totalOrders) * 100
	}

	// Calculate average service time
	if serviceTimeCount > 0 {
		avgServiceTime := totalServiceTime / time.Duration(serviceTimeCount)
		startTime := time.Now().Add(-avgServiceTime)
		stats.AverageServiceTime = utils.FormatDuration(startTime, time.Now())
	} else {
		stats.AverageServiceTime = "Нет данных"
	}

	stats.TopPerformingPeriod = "14:00-16:00" // TODO: Calculate from actual data

	response.Stats = stats
	response.LastUpdated = utils.FormatDate(time.Now())

	// Add filter options
	response.Filters = []responses.FilterOptionResponse{
		{Value: "all", Label: "Все заказы", Count: totalOrders, Selected: true},
		{Value: "completed", Label: "Завершенные", Count: len(completedOrders), Selected: false},
		{Value: "cancelled", Label: "Отмененные", Count: len(cancelledOrders), Selected: false},
	}

	// Add available actions
	response.Actions = []responses.ActionResponse{
		{ID: "export", Label: "Экспорт", Variant: "secondary"},
		{ID: "analytics", Label: "Аналитика", Variant: "primary"},
	}

	// Add pagination
	response.Pagination = responses.PaginationResponse{
		CurrentPage:  1,
		TotalPages:   1,
		TotalItems:   len(response.Orders),
		ItemsPerPage: 50,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetProfile returns the waiter's profile information
func (h *WaiterHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	waiterID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user profile
	profile, err := h.profileService.GetUserProfile(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		http.Error(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}

	// Get assigned tables
	assignedTables, err := h.waiterService.GetAssignedTables(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting assigned tables: %v", err)
		// Continue without assigned tables
	}

	// Get active orders
	orders, err := h.waiterService.GetWaiterOrders(r.Context(), waiterID)
	if err != nil {
		log.Printf("Error getting waiter orders: %v", err)
		// Continue without orders
	}

	// Get performance stats
	stats, err := h.waiterService.GetPerformanceStats(r.Context(), waiterID, time.Now())
	if err != nil {
		log.Printf("Error getting performance stats: %v", err)
		// Continue without performance stats
	}

	// Count order stats
	orderStats := responses.OrdersStatsResponse{}
	for _, order := range orders {
		orderStats.TotalActiveOrders++
		switch order.Status {
		case "new":
			orderStats.New++
		case "accepted":
			orderStats.Accepted++
		case "preparing":
			orderStats.Preparing++
		case "ready":
			orderStats.Ready++
		case "served":
			orderStats.Served++
		}
	}

	// Create enhanced response
	response := &responses.WaiterProfileResponse{
		ID:         profile.User.ID,
		Name:       profile.User.Name,
		Username:   profile.User.Username,
		Email:      profile.User.Email,
		Role:       utils.TranslateRole(profile.User.Role),
		Status:     utils.FormatStatus(profile.User.Status, "user"),
		OrderStats: orderStats,
		AvailableActions: []responses.ActionResponse{
			{ID: "edit_profile", Label: "Редактировать профиль", Variant: "secondary"},
			{ID: "view_schedule", Label: "Просмотр расписания", Variant: "primary"},
		},
	}

	// Add assigned tables with enhanced formatting
	for _, table := range assignedTables {
		response.AssignedTables = append(response.AssignedTables, mapTableToEnhancedResponse(table))
	}

	// Add current shift if available with enhanced formatting
	if profile.CurrentShift != nil {
		shift := profile.CurrentShift
		response.CurrentShift = &responses.ShiftResponse{
			ID:        shift.ID,
			Date:      shift.Date.Format("2006-01-02"),
			StartTime: shift.StartTime.Format("15:04"),
			EndTime:   shift.EndTime.Format("15:04"),
			Duration:  utils.FormatDuration(shift.StartTime, time.Now()),
			Status:    utils.FormatStatus("active", "shift"),
			Manager:   "Менеджер", // TODO: Get actual manager name
		}
	}

	// Add upcoming shifts if available with enhanced formatting
	for _, shift := range profile.UpcomingShifts {
		shiftResponse := &responses.ShiftResponse{
			ID:        shift.ID,
			Date:      shift.Date.Format("2006-01-02"),
			StartTime: shift.StartTime.Format("15:04"),
			EndTime:   shift.EndTime.Format("15:04"),
			Duration:  utils.FormatDuration(shift.StartTime, shift.EndTime),
			Status:    utils.FormatStatus("upcoming", "shift"),
			Manager:   "Менеджер", // TODO: Get actual manager name
		}
		response.UpcomingShifts = append(response.UpcomingShifts, shiftResponse)
	}

	// Add performance data if available with enhanced formatting
	if stats != nil {
		response.PerformanceData = responses.PerformanceResponse{
			TablesServed:    int(stats.TotalOrders), // Use TotalOrders as proxy for tables served
			OrdersCompleted: int(stats.TotalOrders),
			Revenue:         utils.FormatMoney(int(stats.TotalRevenue * 100)), // Convert to cents
			Tips:            utils.FormatMoney(int(stats.TotalTips * 100)),    // Convert to cents
			EfficiencyScore: stats.CalculateEfficiency(),
			CustomerRating:  stats.CustomerRating,
			Ranking:         "Топ 15%", // TODO: Calculate actual ranking
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to map order entity to enhanced response DTO
func mapOrderToEnhancedResponse(order *entity.Order) *responses.OrderResponse {
	resp := &responses.OrderResponse{
		ID:               order.ID,
		TableNumber:      order.TableID, // TODO: Get actual table number
		Status:           utils.FormatStatus(order.Status, "order"),
		TotalAmount:      utils.FormatMoney(order.TotalAmount),
		Comment:          order.Comment,
		CreatedAt:        utils.FormatDate(order.CreatedAt),
		UpdatedAt:        utils.FormatDate(order.UpdatedAt),
		Duration:         utils.FormatDuration(order.CreatedAt, time.Now()),
		AvailableActions: utils.CreateAvailableActions("order", order.Status, "waiter"),
		Priority:         "normal", // TODO: Calculate based on business logic
	}

	// Calculate estimated ready time based on preparation time
	estimatedReady := order.CreatedAt.Add(25 * time.Minute) // Default estimation
	estimatedReadyFormatted := utils.FormatDate(estimatedReady)
	resp.EstimatedReady = &estimatedReadyFormatted

	// Map order items with enhanced formatting
	for _, item := range order.Items {
		itemName := ""
		if item.Dish != nil {
			itemName = item.Dish.Name
		}

		orderItem := &responses.OrderItemResponse{
			ID:         item.ID,
			Name:       itemName,
			Quantity:   item.Quantity,
			Price:      utils.FormatMoney(int(item.Price * 100)),                          // Convert to cents
			TotalPrice: utils.FormatMoney(int(item.Price * float64(item.Quantity) * 100)), // Convert to cents
			Notes:      item.Notes,
			Status:     utils.FormatStatus("pending", "order_item"),
		}

		resp.Items = append(resp.Items, orderItem)
	}

	return resp
}

// Helper function to map order entity to enhanced history response DTO
func mapOrderToHistoryResponse(order *entity.Order) *responses.OrderHistoryResponse {
	resp := &responses.OrderHistoryResponse{
		ID:          order.ID,
		TableNumber: order.TableID, // TODO: Get actual table number
		Status:      utils.FormatStatus(order.Status, "order"),
		TotalAmount: utils.FormatMoney(order.TotalAmount),
		CreatedAt:   utils.FormatDate(order.CreatedAt),
	}

	// Map completion or cancellation time with enhanced formatting
	if order.Status == "completed" && order.CompletedAt != nil {
		completedDate := utils.FormatDate(*order.CompletedAt)
		resp.CompletedAt = &completedDate
		resp.Duration = utils.FormatDuration(order.CreatedAt, *order.CompletedAt)
	} else if order.Status == "cancelled" && order.CancelledAt != nil {
		cancelledDate := utils.FormatDate(*order.CancelledAt)
		resp.CancelledAt = &cancelledDate
		resp.Duration = utils.FormatDuration(order.CreatedAt, *order.CancelledAt)
	}

	// Map order items with enhanced formatting
	for _, item := range order.Items {
		itemName := ""
		if item.Dish != nil {
			itemName = item.Dish.Name
		}

		orderItem := &responses.OrderItemResponse{
			ID:         item.ID,
			Name:       itemName,
			Quantity:   item.Quantity,
			Price:      utils.FormatMoney(int(item.Price * 100)),                          // Convert to cents
			TotalPrice: utils.FormatMoney(int(item.Price * float64(item.Quantity) * 100)), // Convert to cents
			Notes:      item.Notes,
			Status:     utils.FormatStatus("completed", "order_item"),
		}

		resp.Items = append(resp.Items, orderItem)
	}

	return resp
}

// Helper function to map table entity to enhanced response DTO
func mapTableToEnhancedResponse(table *entity.Table) *responses.TableResponse {
	resp := &responses.TableResponse{
		ID:     table.ID,
		Number: table.Number,
		Seats:  table.Seats,
		Status: utils.FormatStatus(table.Status, "table"),
	}

	// Add last activity if table is occupied
	if table.OccupiedAt != nil {
		lastActivityFormatted := utils.FormatDate(*table.OccupiedAt)
		resp.LastActivity = &lastActivityFormatted
		occupiedDuration := utils.FormatDuration(*table.OccupiedAt, time.Now())
		resp.OccupiedDuration = &occupiedDuration
	}

	return resp
}
