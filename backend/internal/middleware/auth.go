package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Middleware для защиты HTML страниц
func HTMLAuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем публичные страницы
			if r.URL.Path == "/" || r.URL.Path == "/login.html" {
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем токен
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Если нет токена, проверяем куки
				cookie, err := r.Cookie("auth_token")
				if err != nil || cookie.Value == "" {
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
				authHeader = "Bearer " + cookie.Value
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtKey), nil
			})

			if err != nil || !token.Valid {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			role := claims["role"].(string)

			// Проверяем доступ к страницам в зависимости от роли
			switch {
			case strings.HasPrefix(r.URL.Path, "/admin") && role != "admin":
				http.Redirect(w, r, "/", http.StatusFound)
				return
			case strings.HasPrefix(r.URL.Path, "/manager") && role != "manager":
				// Разрешаем доступ к статическим файлам
				if strings.HasPrefix(r.URL.Path, "/static/") {
					next.ServeHTTP(w, r)
					return
				}
				// Разрешаем доступ к файлам менеджера
				if strings.HasPrefix(r.URL.Path, "/manager/") {
					next.ServeHTTP(w, r)
					return
				}
				http.Redirect(w, r, "/", http.StatusFound)
				return
			case strings.HasPrefix(r.URL.Path, "/waiter") && role != "waiter":
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// Добавляем данные пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			ctx = context.WithValue(ctx, "role", claims["role"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Существующий AuthMiddleware для API
func AuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Включаем CORS заголовки
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

			// Обработка preflight запросов
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Проверка токена
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Проверка роли
			if strings.HasPrefix(r.URL.Path, "/api/admin") && claims["role"] != "admin" {
				http.Error(w, "Unauthorized: admin access required", http.StatusForbidden)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/waiter") && claims["role"] != "waiter" {
				http.Error(w, "Unauthorized: waiter access required", http.StatusForbidden)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/api/manager") && claims["role"] != "manager" {
				http.Error(w, "Unauthorized: manager access required", http.StatusForbidden)
				return
			}

			// Добавляем данные пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			ctx = context.WithValue(ctx, "role", claims["role"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
