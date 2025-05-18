package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"restaurant-management/configs"
	"restaurant-management/internal/database"
	"restaurant-management/internal/handlers"
	"restaurant-management/internal/middleware"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	// Загружаем конфигурацию
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем базу данных
	db, err := database.NewDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()

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

	// Инициализируем обработчики
	authHandler := handlers.NewAuthHandler(db, config.Server.JWTKey)
	adminHandler := handlers.NewAdminHandler(db)
	inventoryHandler := handlers.NewInventoryHandler(db)
	supplierHandler := handlers.NewSupplierHandler(db)
	requestHandler := handlers.NewRequestHandler(db)
	menuRepo := database.NewMenuRepository(db.DB)
	menuHandler := handlers.NewMenuHandler(menuRepo)
	managerHandler := handlers.NewManagerHandler(db)
	waiterHandler := handlers.NewWaiterHandler(db)
	kitchenHandler := handlers.NewKitchenHandler(db)
	shiftHandler := handlers.NewShiftHandler(db)

	// Публичные API
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Защищенные API маршруты
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(config.Server.JWTKey))

	// API маршруты для админа
	admin := api.PathPrefix("/admin").Subrouter()
	// Оставляем только статистику, остальное будет через эндпоинты менеджера
	admin.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	// API маршруты для менеджера
	manager := api.PathPrefix("/manager").Subrouter()
	manager.HandleFunc("/dashboard", managerHandler.GetDashboard).Methods("GET")
	manager.HandleFunc("/orders/history", waiterHandler.GetOrderHistory).Methods("GET")

	// Добавляем маршруты для пользователей, используя обработчики админа
	manager.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	manager.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	manager.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT")
	manager.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")

	// Добавляем маршруты для смен, используя обработчики смен
	manager.HandleFunc("/shifts", shiftHandler.GetAllShifts).Methods("GET")
	manager.HandleFunc("/shifts", shiftHandler.CreateShift).Methods("POST")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.GetShiftByID).Methods("GET")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.UpdateShift).Methods("PUT")
	manager.HandleFunc("/shifts/{id:[0-9]+}", shiftHandler.DeleteShift).Methods("DELETE")

	// API маршруты для инвентаря (перенесены к менеджеру)
	manager.HandleFunc("/inventory", inventoryHandler.GetAll).Methods("GET")
	manager.HandleFunc("/inventory", inventoryHandler.Create).Methods("POST")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.GetByID).Methods("GET")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.Update).Methods("PUT")
	manager.HandleFunc("/inventory/{id}", inventoryHandler.Delete).Methods("DELETE")

	// API маршруты для поставщиков (перенесены к менеджеру)
	manager.HandleFunc("/suppliers", supplierHandler.GetAll).Methods("GET")
	manager.HandleFunc("/suppliers", supplierHandler.Create).Methods("POST")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.GetByID).Methods("GET")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.Update).Methods("PUT")
	manager.HandleFunc("/suppliers/{id}", supplierHandler.Delete).Methods("DELETE")

	// API маршруты для заявок (перенесены к менеджеру)
	manager.HandleFunc("/requests", requestHandler.GetAll).Methods("GET")
	manager.HandleFunc("/requests", requestHandler.Create).Methods("POST")
	manager.HandleFunc("/requests/{id}", requestHandler.GetByID).Methods("GET")
	manager.HandleFunc("/requests/{id}", requestHandler.Update).Methods("PUT")
	manager.HandleFunc("/requests/{id}", requestHandler.Delete).Methods("DELETE")

	// API маршруты для официанта
	waiter := api.PathPrefix("/waiter").Subrouter()
	waiter.HandleFunc("/tables", waiterHandler.GetTables).Methods("GET")
	waiter.HandleFunc("/tables/{id}/status", waiterHandler.UpdateTableStatus).Methods("PUT")
	waiter.HandleFunc("/orders", waiterHandler.GetOrders).Methods("GET")
	waiter.HandleFunc("/history", waiterHandler.GetOrderHistory).Methods("GET")
	waiter.HandleFunc("/orders", waiterHandler.CreateOrder).Methods("POST")
	waiter.HandleFunc("/orders/{id}/status", waiterHandler.UpdateOrderStatus).Methods("PUT")
	waiter.HandleFunc("/profile", waiterHandler.GetProfile).Methods("GET")

	// API маршруты для кухни
	kitchen := api.PathPrefix("/kitchen").Subrouter()
	kitchen.HandleFunc("/orders", kitchenHandler.GetKitchenOrders).Methods("GET")
	kitchen.HandleFunc("/orders/{id}/status", kitchenHandler.UpdateOrderStatusByCook).Methods("PUT")
	kitchen.HandleFunc("/history", kitchenHandler.GetKitchenHistory).Methods("GET")
	kitchen.HandleFunc("/inventory", kitchenHandler.GetInventory).Methods("GET")
	kitchen.HandleFunc("/inventory/{id}", kitchenHandler.UpdateInventory).Methods("PUT")

	// API маршруты для меню
	menuHandler.RegisterRoutes(r)

	// Регистрация обработчиков API для смен
	apiRouter := api.PathPrefix("/shifts").Subrouter()
	apiRouter.HandleFunc("", shiftHandler.GetEmployeeShifts).Methods("GET")

	// HTML страницы (теперь защищены middleware)
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

	log.Printf("Server starting on http://127.0.0.1:%s", config.Server.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%s", config.Server.Port), r))
}
