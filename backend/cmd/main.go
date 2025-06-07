package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"restaurant-management/configs"
	"restaurant-management/internal/infrastructure/persistence/postgres"
	"restaurant-management/internal/infrastructure/persistence/postgres/repository"
	"restaurant-management/internal/presentation/http/handlers"
	"restaurant-management/internal/presentation/http/middleware"
	"restaurant-management/internal/services"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	postgresConfig := &postgres.Config{
		Host:     config.Database.Host,
		Port:     config.Database.Port,
		User:     config.Database.User,
		Password: config.Database.Password,
		DBName:   config.Database.DBName,
		SSLMode:  config.Database.SSLMode,
	}

	db, err := postgres.NewConnection(postgresConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// REPOS

	repoFactory := repository.NewFactory(db)
	userRepo := repoFactory.NewUserRepository()
	businessRepo := repoFactory.NewBusinessRepository()
	tableRepo := repoFactory.NewTableRepository()
	shiftRepo := repoFactory.NewShiftRepository()
	orderRepo := repoFactory.NewOrderRepository()
	menuRepo := repoFactory.NewMenuRepository()
	supplierRepo := repoFactory.NewSupplierRepository()
	requestRepo := repoFactory.NewRequestRepository()
	inventoryRepo := repoFactory.NewInventoryRepository()

	// SERVICES

	serviceFactory := services.NewFactory(
		userRepo,
		businessRepo,
		tableRepo,
		shiftRepo,
		orderRepo,
		requestRepo,
		supplierRepo,
		menuRepo,
		inventoryRepo,
		config.Server.JWTKey,
	)
	authService := serviceFactory.NewAuthService()
	userService := serviceFactory.NewUserService()
	businessService := serviceFactory.NewBusinessService()
	tableService := serviceFactory.NewTableService()
	shiftService := serviceFactory.NewShiftService()
	orderService := serviceFactory.NewOrderService()
	menuService := serviceFactory.NewMenuService()
	supplierService := serviceFactory.NewSupplierService()
	requestService := serviceFactory.NewRequestService()
	inventoryService := serviceFactory.NewInventoryService()
	waiterService := serviceFactory.NewWaiterService()
	kitchenService := serviceFactory.NewKitchenService()
	managerService := serviceFactory.NewManagerService()

	// ROUTES

	r := mux.NewRouter()

	// static file handling
	fileServer := http.FileServer(http.Dir(config.Paths.Static))
	wrappedFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		fileServer.ServeHTTP(w, r)
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", wrappedFileServer))

	// templates file handling
	templateServer := http.FileServer(http.Dir(config.Paths.Templates))
	wrappedTemplateServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".html") {
			w.Header().Set("Content-Type", "text/html")
		}
		templateServer.ServeHTTP(w, r)
	})
	r.PathPrefix("/templates/").Handler(http.StripPrefix("/templates/", wrappedTemplateServer))

	// Create middleware instances
	authMiddleware := middleware.NewAuthMiddleware(config.Server.JWTKey)
	businessMiddleware := middleware.NewBusinessMiddleware()

	// router for html files handling
	htmlRouter := r.PathPrefix("").Subrouter()
	htmlRouter.Use(authMiddleware.HTMLAuth)

	authHandler := handlers.NewAuthHandler(authService, userService, config.Server.JWTKey)
	businessHandler := handlers.NewBusinessHandler(businessService, userService)
	adminHandler := handlers.NewAdminHandler(userService, businessService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)
	requestHandler := handlers.NewRequestHandler(requestService)
	menuHandler := handlers.NewMenuHandler(menuService)
	managerHandler := handlers.NewManagerHandler(managerService)
	waiterHandler := handlers.NewWaiterHandler(waiterService, tableService, orderService, userService)
	kitchenHandler := handlers.NewKitchenHandler(kitchenService, orderService, inventoryService)
	shiftHandler := handlers.NewShiftHandler(shiftService)

	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Business selection API
	r.HandleFunc("/api/businesses", businessHandler.ListBusinesses).Methods("GET")
	r.HandleFunc("/api/businesses/{id}", businessHandler.GetBusiness).Methods("GET")
	r.HandleFunc("/api/businesses/{id}/select", businessHandler.SelectBusiness).Methods("POST")
	// Add business middleware to all protected APIs
	// Защищенные API маршруты
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware.APIAuth)
	api.Use(businessMiddleware.RequireBusiness)

	// Business management routes (admin only)
	businessAdmin := api.PathPrefix("/admin/businesses").Subrouter()
	businessAdmin.HandleFunc("", businessHandler.CreateBusiness).Methods("POST")
	businessAdmin.HandleFunc("/{id}", businessHandler.UpdateBusiness).Methods("PUT")
	businessAdmin.HandleFunc("/{id}", businessHandler.DeleteBusiness).Methods("DELETE")

	// API маршруты для админа
	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	// API маршруты для менеджера
	manager := api.PathPrefix("/manager").Subrouter()

	// Manager operations
	manager.HandleFunc("/staff", managerHandler.GetStaffList).Methods("GET")
	manager.HandleFunc("/staff", managerHandler.CreateStaffMember).Methods("POST")
	manager.HandleFunc("/staff/{id}", managerHandler.UpdateStaffMember).Methods("PUT")
	manager.HandleFunc("/staff/{id}", managerHandler.DeactivateStaffMember).Methods("DELETE")

	// Reports
	manager.HandleFunc("/reports/daily", managerHandler.GetDailyReport).Methods("GET")
	manager.HandleFunc("/reports/revenue", managerHandler.GetRevenueReport).Methods("GET")
	manager.HandleFunc("/reports/staff-performance", managerHandler.GetStaffPerformanceReport).Methods("GET")
	manager.HandleFunc("/business/hours", managerHandler.UpdateBusinessHours).Methods("PUT")
	manager.HandleFunc("/business/statistics", managerHandler.GetBusinessStatistics).Methods("GET")
	manager.HandleFunc("/history", managerHandler.GetOrderHistory).Methods("GET")

	// Users management (using admin handler)
	manager.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	manager.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	manager.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT")
	manager.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")

	// Shifts management
	manager.HandleFunc("/shifts", shiftHandler.GetShifts).Methods("GET")
	manager.HandleFunc("/shifts", shiftHandler.CreateShift).Methods("POST")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.GetShiftByID).Methods("GET")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.UpdateShift).Methods("PUT")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.DeleteShift).Methods("DELETE")
	manager.HandleFunc("/shifts/{id:[0-9]+}/assign", shiftHandler.AssignEmployee).Methods("POST")
	manager.HandleFunc("/shifts/{id:[0-9]+}/remove", shiftHandler.RemoveEmployee).Methods("POST")

	// API маршруты для инвентаря (перенесены к менеджеру)
	manager.HandleFunc("/inventory", inventoryHandler.GetAll).Methods("GET")
	manager.HandleFunc("/inventory", inventoryHandler.Create).Methods("POST")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.GetByID).Methods("GET")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.Update).Methods("PUT")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.Delete).Methods("DELETE")

	// API маршруты для поставщиков (перенесены к менеджеру)
	manager.HandleFunc("/suppliers", supplierHandler.GetSuppliers).Methods("GET")
	manager.HandleFunc("/suppliers", supplierHandler.CreateSupplier).Methods("POST")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.GetSupplierByID).Methods("GET")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.UpdateSupplier).Methods("PUT")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.DeleteSupplier).Methods("DELETE")
	manager.HandleFunc("/suppliers/{id}/orders", supplierHandler.GetPurchaseOrders).Methods("GET")
	manager.HandleFunc("/purchase-orders", supplierHandler.CreatePurchaseOrder).Methods("POST")

	// API маршруты для заявок (перенесены к менеджеру)
	manager.HandleFunc("/requests", requestHandler.GetRequests).Methods("GET")
	manager.HandleFunc("/requests", requestHandler.CreateRequest).Methods("POST")
	manager.HandleFunc("/requests/{id}", requestHandler.GetRequestByID).Methods("GET")
	manager.HandleFunc("/requests/{id}", requestHandler.UpdateRequest).Methods("PUT")
	manager.HandleFunc("/requests/{id}/status", requestHandler.UpdateRequestStatus).Methods("PUT")
	manager.HandleFunc("/requests/{id}/assign", requestHandler.AssignRequest).Methods("POST")
	manager.HandleFunc("/requests/{id}/acknowledge", requestHandler.AcknowledgeRequest).Methods("POST")
	manager.HandleFunc("/requests/{id}/complete", requestHandler.CompleteRequest).Methods("POST")
	manager.HandleFunc("/request-types", requestHandler.GetRequestTypes).Methods("GET")

	// API маршруты для официанта
	waiter := api.PathPrefix("/waiter").Subrouter()
	waiter.HandleFunc("/tables", waiterHandler.GetTables).Methods("GET")
	waiter.HandleFunc("/tables/{id}/status", waiterHandler.UpdateTableStatus).Methods("PUT")
	waiter.HandleFunc("/orders", waiterHandler.GetOrders).Methods("GET")
	waiter.HandleFunc("/orders", waiterHandler.CreateOrder).Methods("POST")
	waiter.HandleFunc("/orders/{id}/status", waiterHandler.UpdateOrderStatus).Methods("PUT")
	waiter.HandleFunc("/history", waiterHandler.GetOrderHistory).Methods("GET")
	waiter.HandleFunc("/profile", waiterHandler.GetProfile).Methods("GET")

	// API маршруты для кухни
	kitchen := api.PathPrefix("/kitchen").Subrouter()
	kitchen.HandleFunc("/orders", kitchenHandler.GetKitchenOrders).Methods("GET")
	kitchen.HandleFunc("/orders/{id}/status", kitchenHandler.UpdateOrderStatusByCook).Methods("PUT")
	kitchen.HandleFunc("/history", kitchenHandler.GetKitchenHistory).Methods("GET")
	kitchen.HandleFunc("/inventory", kitchenHandler.GetInventory).Methods("GET")
	kitchen.HandleFunc("/inventory/{id}", kitchenHandler.UpdateInventory).Methods("PUT")

	// API маршруты для меню
	menu := api.PathPrefix("/menu").Subrouter()
	menu.HandleFunc("", menuHandler.GetMenu).Methods("GET")
	menu.HandleFunc("/items", menuHandler.GetMenuItems).Methods("GET")
	menu.HandleFunc("/items", menuHandler.CreateMenuItem).Methods("POST")
	menu.HandleFunc("/items/{id}", menuHandler.GetMenuItem).Methods("GET")
	menu.HandleFunc("/items/{id}", menuHandler.UpdateMenuItem).Methods("PUT")
	menu.HandleFunc("/items/{id}", menuHandler.DeleteMenuItem).Methods("DELETE")
	menu.HandleFunc("/categories", menuHandler.GetCategories).Methods("GET")
	menu.HandleFunc("/categories", menuHandler.CreateCategory).Methods("POST")
	menu.HandleFunc("/summary", menuHandler.GetMenuSummary).Methods("GET")

	// Регистрация обработчиков API для смен
	apiRouter := api.PathPrefix("/shifts").Subrouter()
	apiRouter.HandleFunc("", shiftHandler.GetEmployeeShifts).Methods("GET")

	// Authentication routes
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/change-password", authHandler.ChangePassword).Methods("POST")
	auth.HandleFunc("/reset-password", authHandler.ResetPassword).Methods("POST")

	// HTML страницы
	// Add business selection page
	r.HandleFunc("/select-business", func(w http.ResponseWriter, r *http.Request) {
		// Check for auth token cookie
		authCookie, err := r.Cookie("auth_token")
		if err == nil && authCookie != nil {
			// Parse the token to get user role
			token, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Server.JWTKey), nil
			})

			if err == nil && token.Valid {
				// Get claims and user role
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if role, ok := claims["role"].(string); ok {
						// Only allow admins to access this page
						if role != "admin" {
							// Redirect non-admin users to their appropriate pages
							var redirectPath string
							switch role {
							case "manager":
								redirectPath = "/manager"
							case "waiter":
								redirectPath = "/waiter"
							case "cook":
								redirectPath = "/kitchen"
							default:
								redirectPath = "/"
							}

							log.Printf("Non-admin user (role: %s) attempted to access business selection page. Redirecting to %s", role, redirectPath)
							http.Redirect(w, r, redirectPath, http.StatusFound)
							return
						}
					}
				}
			}
		}

		// For admins or if token parsing failed, show the business selection page
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "select-business.html"))
	}).Methods("GET")

	// Other HTML routes with auth middleware
	htmlRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "login.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "admin.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/inventory", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/menu", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/staff", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	// Страницы для официантов
	htmlRouter.HandleFunc("/waiter", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "waiter.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/orders", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "waiter.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/history", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "waiter.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/profile", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "waiter.html"))
	}).Methods("GET")

	// Страница для кухни
	htmlRouter.HandleFunc("/kitchen", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "kitchen.html"))
	}).Methods("GET")

	log.Printf("Server starting on port %s", config.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Server.Port), r))
}
