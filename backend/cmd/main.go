package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"user-management/configs"
	"user-management/internal/database"
	"user-management/internal/handlers"
	"user-management/internal/middleware"

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

	// API маршруты
	authHandler := handlers.NewAuthHandler(db, config.Server.JWTKey)
	adminHandler := handlers.NewAdminHandler(db)
	// tableHandler := handlers.NewTableHandler(db)
	// dishHandler := handlers.NewDishHandler(db)

	// Публичные API
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Защищенные API маршруты
	api := r.PathPrefix("/api/admin").Subrouter()
	api.Use(middleware.AuthMiddleware(config.Server.JWTKey))

	api.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	api.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	api.HandleFunc("/stats", adminHandler.GetStats).Methods("GET")

	// api.HandleFunc("/tables", tableHandler.GetTables).Methods("GET")
	// api.HandleFunc("/tables/{id}", tableHandler.GetTable).Methods("GET")
	// api.HandleFunc("/tables/{id}/status", tableHandler.UpdateTableStatus).Methods("PUT")

	// api.HandleFunc("/menu", dishHandler.GetMenu).Methods("GET")
	// api.HandleFunc("/menu/{id}", dishHandler.GetDish).Methods("GET")
	// HTML страницы
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "login.html"))
	}).Methods("GET")

	r.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(config.Paths.Templates, "admin.html"))
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
