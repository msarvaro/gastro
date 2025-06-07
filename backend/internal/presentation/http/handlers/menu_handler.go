package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"restaurant-management/internal/presentation/http/middleware"

	"github.com/gorilla/mux"
)

// MenuHandler handles menu-related HTTP requests
type MenuHandler struct {
	menuService services.MenuService
}

// NewMenuHandler creates a new menu handler
func NewMenuHandler(menuService services.MenuService) *MenuHandler {
	return &MenuHandler{
		menuService: menuService,
	}
}

// GetMenu retrieves the active menu for a business
func (h *MenuHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	menu, err := h.menuService.GetMenuByBusinessID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		http.Error(w, "Failed to get menu", http.StatusInternalServerError)
		return
	}

	menuResponse := mapMenuToResponse(menu)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuResponse)
}

// GetMenuItems retrieves menu items with optional filtering
func (h *MenuHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	// Get available dishes (this serves as getting menu items)
	dishes, err := h.menuService.GetAvailableDishes(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting menu items: %v", err)
		http.Error(w, "Failed to get menu items", http.StatusInternalServerError)
		return
	}

	dishResponses := make([]*responses.DishResponse, len(dishes))
	for i, dish := range dishes {
		dishResponses[i] = mapDishToResponse(dish)
	}

	response := responses.DishListResponse{
		Dishes: dishResponses,
		Total:  len(dishResponses),
		Page:   1,
		Limit:  len(dishResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMenuItem retrieves a specific menu item by ID
func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}

	// For now, get all dishes and find the one with the matching ID
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	dishes, err := h.menuService.GetAvailableDishes(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting menu items: %v", err)
		http.Error(w, "Failed to get menu item", http.StatusInternalServerError)
		return
	}

	var dish *entity.Dish
	for _, d := range dishes {
		if d.ID == id {
			dish = d
			break
		}
	}

	if dish == nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}

	dishResponse := mapDishToResponse(dish)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dishResponse)
}

// CreateMenuItem creates a new menu item
func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	var req requests.CreateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	dish := &entity.Dish{
		CategoryID:      req.CategoryID,
		BusinessID:      businessID,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		ImageURL:        req.ImageURL,
		IsAvailable:     req.IsAvailable != nil && *req.IsAvailable,
		PreparationTime: req.PreparationTime,
		Calories:        req.Calories,
		Allergens:       req.Allergens,
	}

	err := h.menuService.AddDish(r.Context(), dish)
	if err != nil {
		log.Printf("Error creating menu item: %v", err)
		http.Error(w, "Failed to create menu item", http.StatusInternalServerError)
		return
	}

	dishResponse := mapDishToResponse(dish)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dishResponse)
}

// UpdateMenuItem updates an existing menu item
func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}

	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	var req requests.UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	dish := &entity.Dish{
		ID:              id,
		CategoryID:      req.CategoryID,
		BusinessID:      businessID,
		Name:            req.Name,
		Description:     req.Description,
		Price:           req.Price,
		ImageURL:        req.ImageURL,
		IsAvailable:     req.IsAvailable != nil && *req.IsAvailable,
		PreparationTime: req.PreparationTime,
		Calories:        req.Calories,
		Allergens:       req.Allergens,
	}

	err = h.menuService.UpdateDish(r.Context(), dish)
	if err != nil {
		log.Printf("Error updating menu item: %v", err)
		http.Error(w, "Failed to update menu item", http.StatusInternalServerError)
		return
	}

	dishResponse := mapDishToResponse(dish)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dishResponse)
}

// DeleteMenuItem deletes a menu item
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}

	err = h.menuService.RemoveDish(r.Context(), id)
	if err != nil {
		log.Printf("Error deleting menu item: %v", err)
		http.Error(w, "Failed to delete menu item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCategories retrieves all categories for the business
func (h *MenuHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	// Get the active menu and its categories
	menu, err := h.menuService.GetMenuByBusinessID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}

	categoryResponses := make([]*responses.CategoryResponse, len(menu.Categories))
	for i, category := range menu.Categories {
		categoryResponses[i] = mapCategoryToResponse(category)
	}

	response := responses.CategoryListResponse{
		Categories: categoryResponses,
		Total:      len(categoryResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateCategory creates a new category
func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	var req requests.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	category := &entity.Category{
		MenuID:     req.MenuID,
		Name:       req.Name,
		BusinessID: businessID,
	}

	err := h.menuService.AddCategory(r.Context(), category)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		http.Error(w, "Failed to create category", http.StatusInternalServerError)
		return
	}

	categoryResponse := mapCategoryToResponse(category)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(categoryResponse)
}

// GetMenuSummary retrieves a summary of the menu
func (h *MenuHandler) GetMenuSummary(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	menu, err := h.menuService.GetMenuByBusinessID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting menu: %v", err)
		http.Error(w, "Failed to get menu summary", http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	totalCategories := len(menu.Categories)
	var totalDishes, availableDishes int
	var totalPrice float64

	for _, category := range menu.Categories {
		totalDishes += len(category.Dishes)
		for _, dish := range category.Dishes {
			if dish.IsAvailable {
				availableDishes++
			}
			totalPrice += dish.Price
		}
	}

	avgPrice := float64(0)
	if totalDishes > 0 {
		avgPrice = totalPrice / float64(totalDishes)
	}

	summary := responses.MenuSummaryResponse{
		Menu:            mapMenuToResponse(menu),
		TotalCategories: totalCategories,
		TotalDishes:     totalDishes,
		AvailableDishes: availableDishes,
		AveragePrice:    avgPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// Helper functions for mapping entities to responses
func mapMenuToResponse(menu *entity.Menu) *responses.MenuEntityResponse {
	categoryResponses := make([]*responses.CategoryResponse, len(menu.Categories))
	for i, category := range menu.Categories {
		categoryResponses[i] = mapCategoryToResponse(category)
	}

	return &responses.MenuEntityResponse{
		ID:          menu.ID,
		BusinessID:  menu.BusinessID,
		Name:        menu.Name,
		Description: menu.Description,
		IsActive:    menu.IsActive,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
		Categories:  categoryResponses,
	}
}

func mapCategoryToResponse(category *entity.Category) *responses.CategoryResponse {
	dishResponses := make([]*responses.DishResponse, len(category.Dishes))
	for i, dish := range category.Dishes {
		dishResponses[i] = mapDishToResponse(dish)
	}

	return &responses.CategoryResponse{
		ID:         category.ID,
		MenuID:     category.MenuID,
		Name:       category.Name,
		BusinessID: category.BusinessID,
		CreatedAt:  category.CreatedAt,
		UpdatedAt:  category.UpdatedAt,
		Dishes:     dishResponses,
		DishCount:  len(category.Dishes),
	}
}

func mapDishToResponse(dish *entity.Dish) *responses.DishResponse {
	optionResponses := make([]*responses.MenuItemOptionResponse, len(dish.Options))
	for i, option := range dish.Options {
		optionResponses[i] = &responses.MenuItemOptionResponse{
			ID:            option.ID,
			DishID:        option.DishID,
			Name:          option.Name,
			PriceModifier: option.PriceModifier,
			IsAvailable:   option.IsAvailable,
		}
	}

	return &responses.DishResponse{
		ID:              dish.ID,
		CategoryID:      dish.CategoryID,
		BusinessID:      dish.BusinessID,
		Name:            dish.Name,
		Description:     dish.Description,
		Price:           dish.Price,
		ImageURL:        dish.ImageURL,
		IsAvailable:     dish.IsAvailable,
		PreparationTime: dish.PreparationTime,
		Calories:        dish.Calories,
		Allergens:       dish.Allergens,
		CreatedAt:       dish.CreatedAt,
		UpdatedAt:       dish.UpdatedAt,
		Options:         optionResponses,
		CanBeOrdered:    dish.CanBeOrdered(),
	}
}
