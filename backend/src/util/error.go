package util

import "fmt"

// AppError is a custom error type for application-specific errors.
type AppError struct {
	Message string
	Code    int // HTTP status code or custom error code
	Err     error // Original error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError.
func NewAppError(message string, code int, err error) *AppError {
	return &AppError{
		Message: message,
		Code:    code,
		Err:     err,
	}
}
