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
	tableHandler := handlers.NewTableHandler(db)
	orderHandler := handlers.NewOrderHandler(db)
	menuRepo := database.NewMenuRepository(db.DB)
	menuHandler := handlers.NewMenuHandler(menuRepo)
	managerHandler := handlers.NewManagerHandler(db)

	// Публичные API
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Защищенные API маршруты
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(config.Server.JWTKey))

	// API маршруты для менеджера
	manager := api.PathPrefix("/manager").Subrouter()
	manager.HandleFunc("/dashboard", managerHandler.GetDashboard).Methods("GET")
	manager.HandleFunc("/orders/history", managerHandler.GetOrderHistory).Methods("GET")

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

	// API маршруты для пользователей (остаются у админа)
	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	// API маршруты для столов (доступны для официантов)
	waiter := api.PathPrefix("/waiter").Subrouter()
	waiter.HandleFunc("/tables", tableHandler.GetAll).Methods("GET")
	waiter.HandleFunc("/tables/{id}", tableHandler.GetByID).Methods("GET")
	waiter.HandleFunc("/tables/status", tableHandler.GetStatus).Methods("GET")
	waiter.HandleFunc("/tables/{id}/status", tableHandler.UpdateStatus).Methods("PUT")

	// API маршруты для заказов (доступны для официантов)
	waiter.HandleFunc("/orders", orderHandler.GetAll).Methods("GET")
	waiter.HandleFunc("/orders/{id}", orderHandler.GetByID).Methods("GET")
	waiter.HandleFunc("/orders", orderHandler.Create).Methods("POST")
	waiter.HandleFunc("/orders/{id}", orderHandler.Update).Methods("PUT")
	waiter.HandleFunc("/orders/status", orderHandler.GetStatus).Methods("GET")
	waiter.HandleFunc("/orders/history", orderHandler.GetHistory).Methods("GET")

	// API маршруты для меню
	menuHandler.RegisterRoutes(r)

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

	htmlRouter.HandleFunc("/manager/finances", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/staff", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/settings", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/manager/analytics", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "manager.html"))
	}).Methods("GET")

	// Страницы для официантов
	htmlRouter.HandleFunc("/waiter", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "index.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/orders", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "orders.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/history", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "history.html"))
	}).Methods("GET")

	htmlRouter.HandleFunc("/waiter/create-order", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "create-order.html"))
	}).Methods("GET")

	// Логируем пути для отладки
	log.Printf("Using configuration:")
	log.Printf("Project root: %s", config.Paths.ProjectRoot)
	log.Printf("Frontend dir: %s", config.Paths.Frontend)
	log.Printf("Static dir: %s", config.Paths.Static)
	log.Printf("Templates dir: %s", config.Paths.Templates)

	serverAddr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Server starting on http://localhost%s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
