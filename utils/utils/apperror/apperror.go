package apperror

import (
	"net/http"
	constant "vestra-ecommerce/utils/constants"
)

type AppError struct {
	Status  int
	Message string
	Code    string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(status int, code, message string) *AppError {
	return &AppError{
		Status:  status,
		Message: message,
		Code:    code,
	}
}

// Common errors
var (
	ErrInvalidRequest = New(http.StatusBadRequest, constant.INVALID_REQUEST, "Invalid request body")
	ErrUnauthorized   = New(http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	ErrInternal       = New(http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
)
