package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Middleware для защиты HTML страниц
func HTMLAuthMiddleware(jwtKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Skip public pages (like "/")
			if r.URL.Path == "/" { // Or any other public paths like "/login.html"
				next.ServeHTTP(w, r)
				return
			}

			// 2. Get token string (primarily from cookie for HTML navigation)
			var tokenString string
			cookie, err := r.Cookie("auth_token")
			if err != nil || cookie.Value == "" {
				// Optional: Fallback to Authorization header if you want, or just redirect
				log.Println("HTMLAuthMiddleware: auth_token cookie not found or empty.")
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			tokenString = cookie.Value

			// 3. Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtKey), nil
			})
			if err != nil || !token.Valid {
				log.Printf("HTMLAuthMiddleware: Invalid token from cookie: %v", err)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// 4. Get claims and validated roleClaim
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Println("HTMLAuthMiddleware: Invalid token claims from cookie")
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			roleClaim, roleOk := claims["role"].(string)
			if !roleOk || roleClaim == "" {
				log.Printf("HTMLAuthMiddleware: Role claim missing/invalid in JWT (cookie) for path: %s", r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			log.Printf("HTMLAuth: Path='%s', Role='%s' (from cookie)", r.URL.Path, roleClaim)

			// 5. Role-based access control (fall-through style)
			switch {
			case strings.HasPrefix(r.URL.Path, "/admin") && roleClaim != "admin":
				log.Printf("HTMLAuth: Admin access denied for role '%s' to path '%s'", roleClaim, r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return // Only return if redirecting
			case strings.HasPrefix(r.URL.Path, "/manager") && roleClaim != "manager":
				// Special handling for static files if a non-manager somehow requests them via /manager path
				if strings.HasPrefix(r.URL.Path, "/static/") {
					log.Printf("HTMLAuth: Allowing non-manager '%s' direct access to static file: %s", roleClaim, r.URL.Path)
					next.ServeHTTP(w, r) // Let static files pass
					return
				}
				log.Printf("HTMLAuth: Manager access denied for role '%s' to path '%s'", roleClaim, r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return // Only return if redirecting
			case strings.HasPrefix(r.URL.Path, "/waiter") && roleClaim != "waiter":
				log.Printf("HTMLAuth: Waiter access denied for role '%s' to path '%s'", roleClaim, r.URL.Path)
				http.Redirect(w, r, "/", http.StatusFound)
				return // Only return if redirecting
			}

			// 6. If no redirect happened, access is granted. Set context and serve.
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
