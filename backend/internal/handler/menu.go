package handler

import (
	"encoding/json"
	"net/http"
	"restaurant-management/internal/domain/menu"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type MenuController struct {
	menuService menu.Service
}

func NewMenuController(menuService menu.Service) *MenuController {
	return &MenuController{
		menuService: menuService,
	}
}

func (c *MenuController) RegisterRoutes(r *mux.Router) {
	menuRouter := r.PathPrefix("/menu").Subrouter()
	menuRouter.HandleFunc("/items", c.GetMenuItems).Methods("GET")
	menuRouter.HandleFunc("/items/{id:[0-9]+}", c.GetMenuItem).Methods("GET")
	menuRouter.HandleFunc("/items", c.CreateMenuItem).Methods("POST")
	menuRouter.HandleFunc("/items/{id:[0-9]+}", c.UpdateMenuItem).Methods("PUT")
	menuRouter.HandleFunc("/items/{id:[0-9]+}", c.DeleteMenuItem).Methods("DELETE")
	menuRouter.HandleFunc("/categories", c.GetCategories).Methods("GET")
	menuRouter.HandleFunc("/categories/{id:[0-9]+}", c.GetCategory).Methods("GET")
	menuRouter.HandleFunc("/categories", c.CreateCategory).Methods("POST")
	menuRouter.HandleFunc("/categories/{id:[0-9]+}", c.UpdateCategory).Methods("PUT")
	menuRouter.HandleFunc("/categories/{id:[0-9]+}", c.DeleteCategory).Methods("DELETE")
	menuRouter.HandleFunc("", c.GetMenuSummary).Methods("GET")
}

func (c *MenuController) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
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

	items, err := c.menuService.GetMenuItems(r.Context(), categoryID, businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(items)
}

func (c *MenuController) GetMenuItem(w http.ResponseWriter, r *http.Request) {
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
	item, err := c.menuService.GetMenuItemByID(r.Context(), id, businessID)
	if err != nil {
		switch err {
		case menu.ErrMenuItemNotFound:
			http.Error(w, "Menu item not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(item)
}

func (c *MenuController) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var item menu.MenuItemCreate
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	created, err := c.menuService.CreateMenuItem(r.Context(), item, businessID)
	if err != nil {
		switch err {
		case menu.ErrInvalidMenuData:
			http.Error(w, "Invalid menu item data", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (c *MenuController) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid menu item ID", http.StatusBadRequest)
		return
	}
	var item menu.MenuItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	updated, err := c.menuService.UpdateMenuItem(r.Context(), id, item, businessID)
	if err != nil {
		switch err {
		case menu.ErrMenuItemNotFound:
			http.Error(w, "Menu item not found", http.StatusNotFound)
		case menu.ErrInvalidMenuData:
			http.Error(w, "Invalid menu item data", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (c *MenuController) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
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

	if err := c.menuService.DeleteMenuItem(r.Context(), id, businessID); err != nil {
		switch err {
		case menu.ErrMenuItemNotFound:
			http.Error(w, "Menu item not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *MenuController) GetCategories(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	categories, err := c.menuService.GetCategories(r.Context(), businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(categories)
}

func (c *MenuController) GetCategory(w http.ResponseWriter, r *http.Request) {
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
	category, err := c.menuService.GetCategoryByID(r.Context(), id, businessID)
	if err != nil {
		switch err {
		case menu.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(category)
}

func (c *MenuController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category menu.CategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	created, err := c.menuService.CreateCategory(r.Context(), category, businessID)
	if err != nil {
		switch err {
		case menu.ErrInvalidMenuData:
			http.Error(w, "Invalid category data", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (c *MenuController) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	var category menu.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	updated, err := c.menuService.UpdateCategory(r.Context(), id, category, businessID)
	if err != nil {
		switch err {
		case menu.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		case menu.ErrInvalidMenuData:
			http.Error(w, "Invalid category data", http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(updated)
}

func (c *MenuController) DeleteCategory(w http.ResponseWriter, r *http.Request) {
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

	if err := c.menuService.DeleteCategory(r.Context(), id, businessID); err != nil {
		switch err {
		case menu.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (c *MenuController) GetMenuSummary(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok || businessID == 0 {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	summary, err := c.menuService.GetMenuSummary(r.Context(), businessID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
