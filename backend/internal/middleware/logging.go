package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs all incoming HTTP requests
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // default status code
			}

			// Log the incoming request
			log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Log the response with duration
			duration := time.Since(start)
			log.Printf("[RESPONSE] %s %s - Status: %d - Duration: %v",
				r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}
