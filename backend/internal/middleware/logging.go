package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs all HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Log the request
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		
		// Create a custom response writer to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Call the next handler
		next.ServeHTTP(lrw, r)
		
		// Log the response
		log.Printf("Completed %s %s %d in %v", 
			r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// ErrorHandlingMiddleware provides centralized error handling
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)
				http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}