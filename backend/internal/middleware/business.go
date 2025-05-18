package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
)

// BusinessContext key type
type businessContextKey string

// BusinessIDKey is the key for the business ID in the request context
const BusinessIDKey businessContextKey = "business_id"

// BusinessMiddleware extracts business_id from cookies and adds it to the request context
func BusinessMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract business_id from cookie
			businessCookie, err := r.Cookie("business_id")

			var businessID int

			if err == nil && businessCookie != nil {
				// Try to parse the business ID from the cookie
				businessIDStr := businessCookie.Value
				if businessIDStr != "" {
					id, err := strconv.Atoi(businessIDStr)
					if err == nil {
						businessID = id
					}
				}
			}

			// Also check header (for API requests)
			businessIDHeader := r.Header.Get("X-Business-ID")
			if businessIDHeader != "" {
				id, err := strconv.Atoi(businessIDHeader)
				if err == nil {
					businessID = id
				}
			}

			// Add business_id to context
			ctx := context.WithValue(r.Context(), BusinessIDKey, businessID)

			// Log business ID for debugging purposes
			log.Printf("Request for business ID: %d", businessID)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetBusinessIDFromContext extracts business ID from request context
func GetBusinessIDFromContext(ctx context.Context) (int, bool) {
	businessID, ok := ctx.Value(BusinessIDKey).(int)
	return businessID, ok
}
