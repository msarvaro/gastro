package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Context keys
type ContextKey string

const (
	UserIDKey     ContextKey = "user_id"
	RoleKey       ContextKey = "role"
	BusinessIDKey ContextKey = "business_id"
)

// AuthMiddleware protects API endpoints by validating JWT tokens
type AuthMiddleware struct {
	jwtKey string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtKey string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtKey: jwtKey,
	}
}

// APIAuth middleware for API endpoints
func (m *AuthMiddleware) APIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Try to get token from cookie first
		tokenString := ""
		cookie, err := r.Cookie("auth_token")
		if err == nil && cookie.Value != "" {
			tokenString = cookie.Value
		} else {
			// If no cookie, try Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}
			tokenString = bearerToken[1]
		}

		// No valid token found
		if tokenString == "" {
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtKey), nil
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

		// Check role (hierarchical)
		requestedPath := r.URL.Path
		userRole, _ := claims["role"].(string)

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

		// Add user data to request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		ctx = context.WithValue(ctx, RoleKey, claims["role"])
		if businessID, exists := claims["business_id"]; exists {
			ctx = context.WithValue(ctx, BusinessIDKey, businessID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HTMLAuth middleware for web pages
func (m *AuthMiddleware) HTMLAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			next.ServeHTTP(w, r)
			return
		}

		// Try to get token from cookie first
		tokenString := ""
		cookie, err := r.Cookie("auth_token")
		if err == nil && cookie.Value != "" {
			tokenString = cookie.Value
		} else {
			// If no cookie, try Authorization header
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
			tokenString = bearerToken[1]
		}

		// No valid token found
		if tokenString == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtKey), nil
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
		roleClaim, roleOk := claims["role"].(string)
		if !roleOk || roleClaim == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Role-based access control (hierarchical)
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
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		ctx = context.WithValue(ctx, RoleKey, roleClaim)
		if businessID, exists := claims["business_id"]; exists {
			ctx = context.WithValue(ctx, BusinessIDKey, businessID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userIDVal := ctx.Value(UserIDKey)
	if userIDVal == nil {
		return 0, false
	}

	// JWT claims["user_id"] is usually a float64 after parsing
	switch v := userIDVal.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		id, err := strconv.Atoi(v)
		if err == nil {
			return id, true
		}
	}

	return 0, false
}

// GetBusinessIDFromContext extracts the business ID from the request context
func GetBusinessIDFromContext(ctx context.Context) (int, bool) {
	businessIDVal := ctx.Value(BusinessIDKey)
	if businessIDVal == nil {
		return 0, false
	}

	// JWT claims["business_id"] is usually a float64 after parsing
	switch v := businessIDVal.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		id, err := strconv.Atoi(v)
		if err == nil {
			return id, true
		}
	}

	return 0, false
}

// GetUserRoleFromContext extracts the user role from the request context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	roleVal := ctx.Value(RoleKey)
	if roleVal == nil {
		return "", false
	}

	role, ok := roleVal.(string)
	return role, ok
}
