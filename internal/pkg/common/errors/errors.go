package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Application error codes
const (
	// Error codes from 1000-1999 for client errors
	ErrCodeValidation     = 1000
	ErrCodeAuthentication = 1001
	ErrCodeAuthorization  = 1002
	ErrCodeNotFound       = 1003
	ErrCodeBadRequest     = 1004
	ErrCodeConflict       = 1005

	// Error codes from 2000-2999 for server errors
	ErrCodeInternal    = 2000
	ErrCodeDatabase    = 2001
	ErrCodeExternalAPI = 2002
	ErrCodeForbidden   = 2003
)

// AppError represents an application error
type AppError struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	HTTPStatus int         `json:"-"`
	Err        error       `json:"-"`
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details interface{}) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		HTTPStatus: http.StatusForbidden,
		Message:    message,
	}
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeAuthentication,
		HTTPStatus: http.StatusUnauthorized,
		Message:    message,
	}
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeAuthorization,
		HTTPStatus: http.StatusUnauthorized,
		Message:    message,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		HTTPStatus: http.StatusNotFound,
		Message:    message,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		HTTPStatus: http.StatusBadRequest,
		Message:    message,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		HTTPStatus: http.StatusConflict,
		Message:    message,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		HTTPStatus: http.StatusInternalServerError,
		Message:    message,
	}
}

func NewAppError(message string, err error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		HTTPStatus: http.StatusInternalServerError,
		Message:    message,
		Err:        err,
	}
}

// AsAppError converts any error to an AppError
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return NewAppError("Internal server error", err)
}
