package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// APIError represents a structured error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// ErrorResponse represents a standard error response structure
type ErrorResponse struct {
	Status  string  `json:"status"`
	Error   string  `json:"error"`
	Code    int     `json:"code,omitempty"`
	Details *string `json:"details,omitempty"`
}

// ErrorMiddleware provides centralized error handling for the API
func ErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set default headers
		w.Header().Set("Content-Type", "application/json")
		
		// Create a custom response writer to capture errors
		ew := &errorResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		
		// Handle panics
		defer func() {
			if recover := recover(); recover != nil {
				log.Printf("Panic recovered in handler: %v", recover)
				ew.WriteError(http.StatusInternalServerError, "Internal server error", "A panic occurred while processing the request")
			}
		}()
		
		// Call the next handler
		next.ServeHTTP(ew, r)
	})
}

// errorResponseWriter wraps http.ResponseWriter to provide error handling
type errorResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (ew *errorResponseWriter) WriteError(statusCode int, message, detail string) {
	ew.statusCode = statusCode
	ew.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Status:  "error",
		Error:   message,
		Code:    statusCode,
	}
	
	if detail != "" {
		response.Details = &detail
	}
	
	if err := json.NewEncoder(ew).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
		// Fallback to plain text error
		ew.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(ew, "%d %s: %s", statusCode, message, detail)
	}
}

func (ew *errorResponseWriter) Write(p []byte) (int, error) {
	return ew.ResponseWriter.Write(p)
}

func (ew *errorResponseWriter) WriteHeader(statusCode int) {
	ew.statusCode = statusCode
	ew.ResponseWriter.WriteHeader(statusCode)
}

// ValidationError creates a structured validation error response
func ValidationError(w http.ResponseWriter, field, message string) {
	err := &errorResponseWriter{ResponseWriter: w, statusCode: http.StatusBadRequest}
	err.WriteError(http.StatusBadRequest, 
		fmt.Sprintf("Validation error: %s", field), 
		message)
}

// DatabaseError handles database-related errors
func DatabaseError(w http.ResponseWriter, operation string, err error) {
	log.Printf("Database error during %s: %v", operation, err)
	
	errWriter := &errorResponseWriter{ResponseWriter: w, statusCode: http.StatusInternalServerError}
	errWriter.WriteError(http.StatusInternalServerError, 
		"Database error", 
		fmt.Sprintf("Failed to %s: %v", operation, err))
}

// NotFoundError handles not found errors
func NotFoundError(w http.ResponseWriter, resource string) {
	errWriter := &errorResponseWriter{ResponseWriter: w, statusCode: http.StatusNotFound}
	errWriter.WriteError(http.StatusNotFound, 
		fmt.Sprintf("%s not found", resource), 
		"The requested resource could not be found")
}

// UnauthorizedError handles unauthorized errors
func UnauthorizedError(w http.ResponseWriter, message string) {
	errWriter := &errorResponseWriter{ResponseWriter: w, statusCode: http.StatusUnauthorized}
	errWriter.WriteError(http.StatusUnauthorized, 
		"Unauthorized", 
		message)
}

// ConflictError handles conflict errors (e.g., duplicate resources)
func ConflictError(w http.ResponseWriter, message string) {
	errWriter := &errorResponseWriter{ResponseWriter: w, statusCode: http.StatusConflict}
	errWriter.WriteError(http.StatusConflict, 
		"Conflict", 
		message)
}