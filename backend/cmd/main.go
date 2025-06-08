package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"restaurant-management/configs"
	"restaurant-management/internal/infrastructure/storage/postgres"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/service"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем базу данных
	db, err := postgres.NewDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize infrastructure layer (repositories)
	postgresDB := &postgres.DB{DB: db.DB}
	businessRepo := postgres.NewBusinessRepository(postgresDB)
	userRepo := postgres.NewUserRepository(postgresDB)
	menuRepo := postgres.NewMenuRepository(postgresDB)
	orderRepo := postgres.NewOrderRepository(postgresDB)
	tableRepo := postgres.NewTableRepository(postgresDB)
	inventoryRepo := postgres.NewInventoryRepository(postgresDB)
	shiftRepo := postgres.NewShiftRepository(postgresDB)
	supplierRepo := postgres.NewSupplierRepository(postgresDB)
	requestRepo := postgres.NewRequestRepository(postgresDB)
	waiterRepo := postgres.NewWaiterRepository(postgresDB)

	// Initialize services
	services := service.NewServices(
		businessRepo,
		userRepo,
		menuRepo,
		orderRepo,
		tableRepo,
		inventoryRepo,
		shiftRepo,
		supplierRepo,
		requestRepo,
		waiterRepo,
		config.Server.JWTKey,
	)

	// Initialize controllers
	controllers := handler.NewControllers(
		services.Business,
		services.User,
		services.Menu,
		services.Order,
		services.Table,
		services.Inventory,
		services.Shift,
		services.Supplier,
		services.Request,
		services.Waiter,
	)

	r := mux.NewRouter()

	// Apply logging middleware to all routes
	r.Use(middleware.LoggingMiddleware())

	// Создаем FileServer с middleware для установки правильных MIME типов
	fileServer := http.FileServer(http.Dir(config.Paths.Static))
	wrappedFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		fileServer.ServeHTTP(w, r)
	})

	// Обслуживаем статические файлы
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", wrappedFileServer))

	// Создаем роутер для HTML страниц с middleware аутентификации
	htmlRouter := r.PathPrefix("").Subrouter()
	htmlRouter.Use(middleware.HTMLAuthMiddleware(config.Server.JWTKey))

	// All handlers have been migrated to controllers

	// Публичные API
	r.HandleFunc("/api/login", controllers.Auth.Login).Methods("POST")

	// Business selection API (public for login purposes)
	r.HandleFunc("/api/businesses", controllers.Business.GetAllBusinesses).Methods("GET")
	r.HandleFunc("/api/businesses/{id}", controllers.Business.GetBusinessByID).Methods("GET")
	r.HandleFunc("/api/businesses/{id}/select", controllers.Business.SetBusinessCookie).Methods("POST")

	// Add business middleware to all protected APIs
	// Защищенные API маршруты
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(config.Server.JWTKey))
	api.Use(middleware.BusinessMiddleware())

	// Business management routes (admin only)
	businessAdmin := api.PathPrefix("/admin/businesses").Subrouter()
	businessAdmin.HandleFunc("", controllers.Business.CreateBusiness).Methods("POST")
	businessAdmin.HandleFunc("/{id}", controllers.Business.UpdateBusiness).Methods("PUT")
	businessAdmin.HandleFunc("/{id}", controllers.Business.DeleteBusiness).Methods("DELETE")

	// API маршруты для админа (using controllers)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/stats", controllers.Admin.GetStats).Methods("GET")

	// API маршруты для менеджера (using controllers)
	manager := api.PathPrefix("/manager").Subrouter()
	manager.HandleFunc("/history", controllers.Manager.GetOrderHistory).Methods("GET")
	// User management routes
	manager.HandleFunc("/users", controllers.Admin.GetUsers).Methods("GET")
	manager.HandleFunc("/users", controllers.Admin.CreateUser).Methods("POST")
	manager.HandleFunc("/users/{id}", controllers.Admin.UpdateUser).Methods("PUT")
	manager.HandleFunc("/users/{id}", controllers.Admin.DeleteUser).Methods("DELETE")

	// Shift management routes (using controllers)
	manager.HandleFunc("/shifts", controllers.Shift.GetAllShifts).Methods("GET")
	manager.HandleFunc("/shifts", controllers.Shift.CreateShift).Methods("POST")
	manager.HandleFunc("/shifts/{id:[0-9]+}", controllers.Shift.GetShiftByID).Methods("GET")
	manager.HandleFunc("/shifts/{id:[0-9]+}", controllers.Shift.UpdateShift).Methods("PUT")
	manager.HandleFunc("/shifts/{id:[0-9]+}", controllers.Shift.DeleteShift).Methods("DELETE")

	// Inventory management routes (using controllers)
	manager.HandleFunc("/inventory", controllers.Inventory.GetAll).Methods("GET")
	manager.HandleFunc("/inventory", controllers.Inventory.Create).Methods("POST")
	manager.HandleFunc("/inventory/{id}", controllers.Inventory.GetByID).Methods("GET")
	manager.HandleFunc("/inventory/{id}", controllers.Inventory.Update).Methods("PUT")
	manager.HandleFunc("/inventory/{id}", controllers.Inventory.Delete).Methods("DELETE")

	// API маршруты для поставщиков (using controllers)
	manager.HandleFunc("/suppliers", controllers.Supplier.GetAll).Methods("GET")
	manager.HandleFunc("/suppliers", controllers.Supplier.Create).Methods("POST")
	manager.HandleFunc("/suppliers/{id}", controllers.Supplier.GetByID).Methods("GET")
	manager.HandleFunc("/suppliers/{id}", controllers.Supplier.Update).Methods("PUT")
	manager.HandleFunc("/suppliers/{id}", controllers.Supplier.Delete).Methods("DELETE")

	// API маршруты для заявок (using controllers)
	manager.HandleFunc("/requests", controllers.Request.GetAll).Methods("GET")
	manager.HandleFunc("/requests", controllers.Request.Create).Methods("POST")
	manager.HandleFunc("/requests/{id}", controllers.Request.GetByID).Methods("GET")
	manager.HandleFunc("/requests/{id}", controllers.Request.Update).Methods("PUT")
	manager.HandleFunc("/requests/{id}", controllers.Request.Delete).Methods("DELETE")

	// API маршруты для официанта (using controllers)
	waiter := api.PathPrefix("/waiter").Subrouter()
	waiter.HandleFunc("/tables", controllers.Waiter.GetTables).Methods("GET")
	waiter.HandleFunc("/tables/{id}/status", controllers.Waiter.UpdateTableStatus).Methods("PUT")
	waiter.HandleFunc("/orders", controllers.Waiter.GetActiveOrders).Methods("GET")
	waiter.HandleFunc("/history", controllers.Waiter.GetOrderHistory).Methods("GET")
	waiter.HandleFunc("/orders", controllers.Waiter.CreateOrder).Methods("POST")
	waiter.HandleFunc("/orders/{id}/status", controllers.Waiter.UpdateOrderStatus).Methods("PUT")
	waiter.HandleFunc("/profile", controllers.Waiter.GetProfile).Methods("GET")

	// Kitchen routes (using controllers)
	kitchen := api.PathPrefix("/kitchen").Subrouter()
	kitchen.HandleFunc("/orders", controllers.Kitchen.GetKitchenOrders).Methods("GET")
	kitchen.HandleFunc("/orders/{id}/status", controllers.Kitchen.UpdateOrderStatusByCook).Methods("PUT")
	kitchen.HandleFunc("/history", controllers.Kitchen.GetKitchenHistory).Methods("GET")
	kitchen.HandleFunc("/inventory", controllers.Kitchen.GetInventory).Methods("GET")
	kitchen.HandleFunc("/inventory/{id}", controllers.Kitchen.UpdateInventory).Methods("PUT")

	// Menu routes (using controllers)
	controllers.Menu.RegisterRoutes(api)

	// Employee shifts API routes (using controllers)
	apiRouter := api.PathPrefix("/shifts").Subrouter()
	apiRouter.HandleFunc("", controllers.Shift.GetEmployeeShifts).Methods("GET")

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
