package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"restaurant-management/internal/database"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/models"

	"github.com/gorilla/mux"
)

type MenuHandler struct {
	repo *database.MenuRepository
}

func NewMenuHandler(repo *database.MenuRepository) *MenuHandler {
	return &MenuHandler{repo: repo}
}

func (h *MenuHandler) RegisterRoutes(r *mux.Router) {
	menu := r.PathPrefix("/menu").Subrouter()
	menu.HandleFunc("/items", h.GetMenuItems).Methods("GET")
	menu.HandleFunc("/items/{id:[0-9]+}", h.GetMenuItem).Methods("GET")
	menu.HandleFunc("/items", h.CreateMenuItem).Methods("POST")
	menu.HandleFunc("/items/{id:[0-9]+}", h.UpdateMenuItem).Methods("PUT")
	menu.HandleFunc("/items/{id:[0-9]+}", h.DeleteMenuItem).Methods("DELETE")
	menu.HandleFunc("/categories", h.GetCategories).Methods("GET")
	menu.HandleFunc("/categories/{id:[0-9]+}", h.GetCategory).Methods("GET")
	menu.HandleFunc("/categories", h.CreateCategory).Methods("POST")
	menu.HandleFunc("/categories/{id:[0-9]+}", h.UpdateCategory).Methods("PUT")
	menu.HandleFunc("/categories/{id:[0-9]+}", h.DeleteCategory).Methods("DELETE")
	menu.HandleFunc("", h.GetMenuSummary).Methods("GET")
}

func (h *MenuHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	businessIDStr := r.URL.Query().Get("business_id")
	if businessIDStr == "" {
		http.Error(w, "Missing business_id query parameter", http.StatusBadRequest)
		return
	}
	businessID, err := strconv.Atoi(businessIDStr)
	if err != nil || businessID <= 0 {
		http.Error(w, "Invalid business_id query parameter", http.StatusBadRequest)
		return
	}

	categoryIDStr := r.URL.Query().Get("category_id")
	var categoryID *int = nil
	if categoryIDStr != "" {
		id, err := strconv.Atoi(categoryIDStr)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid category_id query parameter", http.StatusBadRequest)
			return
		}
		categoryID = &id
	}

	items, err := h.repo.GetMenuItems(r.Context(), categoryID, businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(items)
}

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	item, err := h.repo.GetMenuItemByID(r.Context(), id, businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var item models.MenuItemCreate
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	item.BusinessID = businessID
	created, err := h.repo.CreateMenuItem(r.Context(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	var item models.MenuItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	item.BusinessID = businessID
	updated, err := h.repo.UpdateMenuItem(r.Context(), id, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteMenuItem(r.Context(), id, businessID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	businessIDStr := r.URL.Query().Get("business_id")
	if businessIDStr == "" {
		http.Error(w, "Missing business_id query parameter", http.StatusBadRequest)
		return
	}
	businessID, err := strconv.Atoi(businessIDStr)
	if err != nil || businessID <= 0 {
		http.Error(w, "Invalid business_id query parameter", http.StatusBadRequest)
		return
	}

	categories, err := h.repo.GetCategories(r.Context(), businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(categories)
}

func (h *MenuHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	category, err := h.repo.GetCategoryByID(r.Context(), id, businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(category)
}

func (h *MenuHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.CategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	category.BusinessID = businessID
	created, err := h.repo.CreateCategory(r.Context(), category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *MenuHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	var category models.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	category.BusinessID = businessID
	updated, err := h.repo.UpdateCategory(r.Context(), id, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updated)
}

func (h *MenuHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}
	if err := h.repo.DeleteCategory(r.Context(), id, businessID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MenuHandler) GetMenuSummary(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	categories, err := h.repo.GetCategories(r.Context(), businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	items, err := h.repo.GetMenuItems(r.Context(), nil, businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := struct {
		Categories []models.Category `json:"categories"`
		Items      []models.MenuItem `json:"items"`
	}{
		Categories: categories,
		Items:      items,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
