package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Middleware для защиты HTML страниц
func HTMLAuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Skip public pages (like "/" and "/login.html")
			if r.URL.Path == "/" || r.URL.Path == "/login.html" {
				next.ServeHTTP(w, r)
				return
			}

			// 2. Get token from Authorization header (Bearer only)
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			tokenString := bearerToken[1]

			// 3. Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtKey), nil
			})
			if err != nil || !token.Valid {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// 4. Get claims and validate roleClaim
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			roleClaim, roleOk := claims["role"].(string)
			if !roleOk || roleClaim == "" {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// 5. Role-based access control (hierarchical)
			switch {
			case roleClaim == "admin":
				// Allow
			case roleClaim == "manager":
				if strings.HasPrefix(r.URL.Path, "/admin") {
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			case roleClaim == "waiter":
				if strings.HasPrefix(r.URL.Path, "/admin") || strings.HasPrefix(r.URL.Path, "/manager") {
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			case roleClaim == "cook":
				if !strings.HasPrefix(r.URL.Path, "/kitchen") {
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			default:
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// If no redirect happened, access is granted. Set context and serve.
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			ctx = context.WithValue(ctx, "role", roleClaim)
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

			// Проверка роли (hierarchical)
			requestedPath := r.URL.Path
			userRole := claims["role"].(string)

			switch {
			// Admin can access all API paths
			case userRole == "admin":
				// Allow
			// Manager can access Manager, Waiter, Kitchen API paths
			case userRole == "manager":
				if strings.HasPrefix(requestedPath, "/api/admin") {
					http.Error(w, "Unauthorized: admin access required", http.StatusForbidden)
					return
				}
			// Waiter can access Waiter and Kitchen API paths
			case userRole == "waiter":
				if strings.HasPrefix(requestedPath, "/api/admin") || strings.HasPrefix(requestedPath, "/api/manager") {
					http.Error(w, "Unauthorized: manager or admin access required", http.StatusForbidden)
					return
				}
			// Cook can access Kitchen API paths
			case userRole == "cook":
				if !strings.HasPrefix(requestedPath, "/api/kitchen") {
					http.Error(w, "Unauthorized: kitchen access required", http.StatusForbidden)
					return
				}
			// Deny access for any other case
			default:
				http.Error(w, "Unauthorized: Insufficient role", http.StatusForbidden)
				return
			}

			// Добавляем данные пользователя в контекст запроса
			ctx := context.WithValue(r.Context(), "user_id", claims["user_id"])
			ctx = context.WithValue(ctx, "role", claims["role"])

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext извлекает user_ID из контекста запроса.
// Возвращает ID пользователя и true, если ID найден и имеет корректный тип.
// В противном случае возвращает 0 и false.
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userIDVal := ctx.Value("user_id")
	if userIDVal == nil {
		return 0, false
	}

	// claims["user_id"] обычно float64 после парсинга JWT, поэтому нужна проверка типа
	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		// Попробуем как int, если вдруг уже преобразовано
		userIDInt, okInt := userIDVal.(int)
		if okInt {
			return userIDInt, true
		}
		// Попробуем как string и сконвертируем (менее вероятно для JWT claims, но для полноты)
		userIDStr, okStr := userIDVal.(string)
		if okStr {
			id, err := strconv.Atoi(userIDStr) // Потребуется импорт "strconv"
			if err == nil {
				return id, true
			}
		}
		return 0, false // Не удалось привести к float64, int или string->int
	}
	return int(userIDFloat), true
}

// GetUserRoleFromContext извлекает role из контекста запроса.
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	roleVal := ctx.Value("role")
	if roleVal == nil {
		return "", false
	}
	roleStr, ok := roleVal.(string)
	return roleStr, ok
}
