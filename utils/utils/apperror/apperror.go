package apperror

import (
	"net/http"
	constant "vestra-ecommerce/utils/constants"
)

// AppError defines a standard application error structure
type AppError struct {
	Status  int    // HTTP status code
	Message string // User-facing error message
	Code    string // Application-specific error code
}

// Error returns the error message
func (e *AppError) Error() string {
	return e.Message
}

// New creates a new AppError instance
func New(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Message: message,
		Code:    code,
	}
}

// Common reusable application errors
var (
	ErrInvalidRequest = New(http.StatusBadRequest, constant.INVALID_REQUEST, "Invalid request body") // Invalid request payload
	ErrUnauthorized   = New(http.StatusUnauthorized, constant.UN_AUTHORIZED, "Unauthorized")        // Unauthorized access
	ErrInternal       = New(http.StatusInternalServerError, constant.INTERNAL_ERROR, "Internal server error") // Internal server error
)
