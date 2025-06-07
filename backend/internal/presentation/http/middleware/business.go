package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
)

// BusinessMiddleware handles business context in requests
type BusinessMiddleware struct {
	// Could include dependencies if needed
}

// NewBusinessMiddleware creates a new business middleware
func NewBusinessMiddleware() *BusinessMiddleware {
	return &BusinessMiddleware{}
}

// RequireBusiness extracts business_id from cookies and headers and adds it to context
func (m *BusinessMiddleware) RequireBusiness(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var businessID int

		// First, check if business_id already exists in the context (from JWT)
		if existingBusinessID, exists := GetBusinessIDFromContext(r.Context()); exists && existingBusinessID != 0 {
			businessID = existingBusinessID
		} else {
			// Extract business_id from cookie
			businessCookie, err := r.Cookie("business_id")
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
			if businessID == 0 {
				businessIDHeader := r.Header.Get("X-Business-ID")
				if businessIDHeader != "" {
					id, err := strconv.Atoi(businessIDHeader)
					if err == nil {
						businessID = id
					}
				}
			}
		}

		// Add business_id to context (update or set)
		ctx := setBusinessIDInContext(r.Context(), businessID)

		// Log business ID for debugging purposes
		log.Printf("Request for business ID: %d", businessID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to set business ID in context
func setBusinessIDInContext(ctx context.Context, businessID int) context.Context {
	return context.WithValue(ctx, BusinessIDKey, businessID)
}
