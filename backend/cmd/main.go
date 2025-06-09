package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"restaurant-management/configs"
	"restaurant-management/internal/handler"
	"restaurant-management/internal/infrastructure/email"
	"restaurant-management/internal/infrastructure/storage/postgres"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/service"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// startNotificationWorker starts a background goroutine that periodically checks for low inventory
// and processes pending notifications automatically
func startNotificationWorker(services *service.Services) {
	log.Println("Starting notification background worker...")

	// Check for low inventory every 30 minutes
	inventoryTicker := time.NewTicker(1000 * time.Hour)

	// Process pending notifications every 5 minutes
	processingTicker := time.NewTicker(1000 * time.Hour)

	go func() {
		defer inventoryTicker.Stop()
		defer processingTicker.Stop()

		for {
			select {
			case <-inventoryTicker.C:
				log.Println("Checking for low inventory levels across all businesses...")
				ctx := context.Background()

				// Get all businesses
				businesses, _, err := services.Business.GetAllBusinesses(ctx)
				if err != nil {
					log.Printf("Error getting businesses: %v", err)
					continue
				}

				// Check each business for low inventory
				for _, business := range businesses {
					log.Printf("Checking inventory for business: %s (ID: %d)", business.Name, business.ID)

					// Get low stock items for this business
					lowStockItems, err := services.Inventory.CheckLowStockLevels(ctx, business.ID)
					if err != nil {
						log.Printf("Error checking inventory levels for business %d: %v", business.ID, err)
						continue
					}

					if len(lowStockItems) == 0 {
						log.Printf("No low stock items found for business %s", business.Name)
						continue
					}

					// Get all managers for this business to send notifications
					users, err := services.User.GetUsers(ctx, business.ID)
					if err != nil {
						log.Printf("Error getting users for business %d: %v", business.ID, err)
						continue
					}

					// Filter managers and get their emails
					var managerEmails []string
					for _, user := range users {
						if user.Role == "manager" && user.Email != "" {
							managerEmails = append(managerEmails, user.Email)
						}
					}

					if len(managerEmails) == 0 {
						log.Printf("No managers found for business %s", business.Name)
						continue
					}

					log.Printf("Found %d managers for business %s: %v", len(managerEmails), business.Name, managerEmails)

					// Send notifications for each low stock item
					for _, item := range lowStockItems {
						err := services.Notification.SendLowInventoryAlert(
							ctx, business.ID, item.Name, item.Quantity, item.MinQuantity, item.Unit,
						)
						if err != nil {
							log.Printf("Error sending low inventory alert for %s in business %s: %v", item.Name, business.Name, err)
						} else {
							log.Printf("Successfully sent low inventory alert for %s to managers of %s", item.Name, business.Name)
						}
					}
				}

			case <-processingTicker.C:
				log.Println("Processing pending notifications...")
				ctx := context.Background()
				err := services.Notification.ProcessPendingNotifications(ctx)
				if err != nil {
					log.Printf("Error processing pending notifications: %v", err)
				}
			}
		}
	}()
}

func main() {
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.NewDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
	notificationRepo := postgres.NewNotificationRepository(postgresDB)

	// Initialize email service
	emailService := email.NewSMTPService(&config.SMTP)

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
		notificationRepo,
		emailService,
		config.Server.JWTKey,
	)

	handlers := handler.NewControllers(
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
		services.Notification,
	)

	r := mux.NewRouter()

	r.Use(middleware.LoggingMiddleware())

	fileServer := http.FileServer(http.Dir(config.Paths.Static))
	wrappedFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Content-Type", "application/javascript")
		}
		fileServer.ServeHTTP(w, r)
	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", wrappedFileServer))

	htmlRouter := r.PathPrefix("").Subrouter()
	htmlRouter.Use(middleware.HTMLAuthMiddleware(config.Server.JWTKey))

	r.HandleFunc("/api/login", handlers.Auth.Login).Methods("POST")
	r.HandleFunc("/api/login/google", handlers.Auth.GoogleLogin).Methods("POST")

	r.HandleFunc("/api/businesses", handlers.Business.GetAllBusinesses).Methods("GET")
	r.HandleFunc("/api/businesses/{id}", handlers.Business.GetBusinessByID).Methods("GET")
	r.HandleFunc("/api/businesses/{id}/select", handlers.Business.SetBusinessCookie).Methods("POST")

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(config.Server.JWTKey))
	api.Use(middleware.BusinessMiddleware())

	businessAdmin := api.PathPrefix("/admin/businesses").Subrouter()
	businessAdmin.HandleFunc("", handlers.Business.CreateBusiness).Methods("POST")
	businessAdmin.HandleFunc("/{id}", handlers.Business.UpdateBusiness).Methods("PUT")
	businessAdmin.HandleFunc("/{id}", handlers.Business.DeleteBusiness).Methods("DELETE")

	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/stats", handlers.Admin.GetStats).Methods("GET")

	manager := api.PathPrefix("/manager").Subrouter()
	manager.HandleFunc("/history", handlers.Manager.GetOrderHistory).Methods("GET")
	manager.HandleFunc("/users", handlers.Admin.GetUsers).Methods("GET")
	manager.HandleFunc("/users", handlers.Admin.CreateUser).Methods("POST")
	manager.HandleFunc("/users/{id}", handlers.Admin.UpdateUser).Methods("PUT")
	manager.HandleFunc("/users/{id}", handlers.Admin.DeleteUser).Methods("DELETE")

	manager.HandleFunc("/shifts", handlers.Shift.GetAllShifts).Methods("GET")
	manager.HandleFunc("/shifts", handlers.Shift.CreateShift).Methods("POST")
	manager.HandleFunc("/shifts/{id:[0-9]+}", handlers.Shift.GetShiftByID).Methods("GET")
	manager.HandleFunc("/shifts/{id:[0-9]+}", handlers.Shift.UpdateShift).Methods("PUT")
	manager.HandleFunc("/shifts/{id:[0-9]+}", handlers.Shift.DeleteShift).Methods("DELETE")

	manager.HandleFunc("/inventory", handlers.Inventory.GetAll).Methods("GET")
	manager.HandleFunc("/inventory", handlers.Inventory.Create).Methods("POST")
	manager.HandleFunc("/inventory/{id}", handlers.Inventory.GetByID).Methods("GET")
	manager.HandleFunc("/inventory/{id}", handlers.Inventory.Update).Methods("PUT")
	manager.HandleFunc("/inventory/{id}", handlers.Inventory.Delete).Methods("DELETE")

	manager.HandleFunc("/suppliers", handlers.Supplier.GetAll).Methods("GET")
	manager.HandleFunc("/suppliers", handlers.Supplier.Create).Methods("POST")
	manager.HandleFunc("/suppliers/{id}", handlers.Supplier.GetByID).Methods("GET")
	manager.HandleFunc("/suppliers/{id}", handlers.Supplier.Update).Methods("PUT")
	manager.HandleFunc("/suppliers/{id}", handlers.Supplier.Delete).Methods("DELETE")

	manager.HandleFunc("/requests", handlers.Request.GetAll).Methods("GET")
	manager.HandleFunc("/requests", handlers.Request.Create).Methods("POST")
	manager.HandleFunc("/requests/{id}", handlers.Request.GetByID).Methods("GET")
	manager.HandleFunc("/requests/{id}", handlers.Request.Update).Methods("PUT")
	manager.HandleFunc("/requests/{id}", handlers.Request.Delete).Methods("DELETE")

	manager.HandleFunc("/notifications", handlers.Notification.GetRecentNotifications).Methods("GET")
	manager.HandleFunc("/notifications/stats", handlers.Notification.GetNotificationStats).Methods("GET")
	manager.HandleFunc("/notifications", handlers.Notification.CreateNotification).Methods("POST")
	manager.HandleFunc("/notifications/inventory-alert", handlers.Notification.SendLowInventoryAlert).Methods("POST")
	manager.HandleFunc("/notifications/hiring-alert", handlers.Notification.SendNewHiringAlert).Methods("POST")
	manager.HandleFunc("/notifications/process", handlers.Notification.ProcessPendingNotifications).Methods("POST")

	waiter := api.PathPrefix("/waiter").Subrouter()
	waiter.HandleFunc("/tables", handlers.Waiter.GetTables).Methods("GET")
	waiter.HandleFunc("/tables/{id}/status", handlers.Waiter.UpdateTableStatus).Methods("PUT")
	waiter.HandleFunc("/orders", handlers.Waiter.GetActiveOrders).Methods("GET")
	waiter.HandleFunc("/history", handlers.Waiter.GetOrderHistory).Methods("GET")
	waiter.HandleFunc("/orders", handlers.Waiter.CreateOrder).Methods("POST")
	waiter.HandleFunc("/orders/{id}/status", handlers.Waiter.UpdateOrderStatus).Methods("PUT")
	waiter.HandleFunc("/profile", handlers.Waiter.GetProfile).Methods("GET")

	kitchen := api.PathPrefix("/kitchen").Subrouter()
	kitchen.HandleFunc("/orders", handlers.Kitchen.GetKitchenOrders).Methods("GET")
	kitchen.HandleFunc("/orders/{id}/status", handlers.Kitchen.UpdateOrderStatusByCook).Methods("PUT")
	kitchen.HandleFunc("/history", handlers.Kitchen.GetKitchenHistory).Methods("GET")
	kitchen.HandleFunc("/inventory", handlers.Kitchen.GetInventory).Methods("GET")
	kitchen.HandleFunc("/inventory/{id}", handlers.Kitchen.UpdateInventory).Methods("PUT")

	handlers.Menu.RegisterRoutes(api)

	apiRouter := api.PathPrefix("/shifts").Subrouter()
	apiRouter.HandleFunc("", handlers.Shift.GetEmployeeShifts).Methods("GET")

	r.HandleFunc("/select-business", func(w http.ResponseWriter, r *http.Request) {
		authCookie, err := r.Cookie("auth_token")
		if err == nil && authCookie != nil {
			token, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.Server.JWTKey), nil
			})

			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if role, ok := claims["role"].(string); ok {
						if role != "admin" {

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

		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "select-business.html"))
	}).Methods("GET")

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

	htmlRouter.HandleFunc("/kitchen", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "kitchen.html"))
	}).Methods("GET")

	// Start background notification worker
	startNotificationWorker(services)

	log.Printf("Server starting on port %s", config.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Server.Port), r))
}
