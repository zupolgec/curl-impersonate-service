package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate request ID
		requestID := uuid.New().String()
		r.Header.Set("X-Request-ID", requestID)

		// Wrap response writer to capture status code
		rw := &responseWriter{w, http.StatusOK}

		// Add request ID to response headers
		rw.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d - %v", requestID, r.Method, r.URL.Path, rw.statusCode, duration)
	})
}
