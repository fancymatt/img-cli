// Package errors provides custom error types and utilities for consistent error handling
// throughout the image generation application. It includes error wrapping, categorization,
// and context addition capabilities.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// ValidationError indicates invalid input or parameters
	ValidationError ErrorType = "VALIDATION_ERROR"

	// FileError indicates file system related errors
	FileError ErrorType = "FILE_ERROR"

	// APIError indicates external API related errors
	APIError ErrorType = "API_ERROR"

	// CacheError indicates cache operation errors
	CacheError ErrorType = "CACHE_ERROR"

	// ConfigError indicates configuration related errors
	ConfigError ErrorType = "CONFIG_ERROR"

	// GenerationError indicates image generation errors
	GenerationError ErrorType = "GENERATION_ERROR"

	// AnalysisError indicates image analysis errors
	AnalysisError ErrorType = "ANALYSIS_ERROR"

	// WorkflowError indicates workflow execution errors
	WorkflowError ErrorType = "WORKFLOW_ERROR"

	// InternalError indicates unexpected internal errors
	InternalError ErrorType = "INTERNAL_ERROR"
)

// AppError represents a custom application error with context
type AppError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errType,
		Message: message,
	}
}

// Newf creates a new AppError with formatted message
func Newf(errType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errType ErrorType, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, preserve the original type if it's more specific
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Type:    appErr.Type,
			Message: fmt.Sprintf("%s: %s", message, appErr.Message),
			Cause:   appErr.Cause,
			Context: appErr.Context,
		}
	}

	return &AppError{
		Type:    errType,
		Message: message,
		Cause:   err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *AppError {
	if err == nil {
		return nil
	}
	return Wrap(err, errType, fmt.Sprintf(format, args...))
}

// Is checks if the error is of a specific type
func Is(err error, errType ErrorType) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == errType
	}
	return false
}

// GetType returns the error type if it's an AppError
func GetType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return InternalError
}

// Validation errors

// ErrInvalidInput creates a validation error for invalid input
func ErrInvalidInput(field, reason string) *AppError {
	return Newf(ValidationError, "invalid %s: %s", field, reason).
		WithContext("field", field).
		WithContext("reason", reason)
}

// ErrMissingRequired creates a validation error for missing required field
func ErrMissingRequired(field string) *AppError {
	return Newf(ValidationError, "missing required field: %s", field).
		WithContext("field", field)
}

// File errors

// ErrFileNotFound creates a file not found error
func ErrFileNotFound(path string) *AppError {
	return Newf(FileError, "file not found: %s", path).
		WithContext("path", path)
}

// ErrFileAccess creates a file access error
func ErrFileAccess(path string, err error) *AppError {
	return Wrapf(err, FileError, "cannot access file: %s", path).
		WithContext("path", path)
}

// API errors

// ErrAPIRequest creates an API request error
func ErrAPIRequest(service string, err error) *AppError {
	return Wrapf(err, APIError, "API request to %s failed", service).
		WithContext("service", service)
}

// ErrAPIResponse creates an API response error
func ErrAPIResponse(service string, status int, message string) *AppError {
	return Newf(APIError, "API error from %s (status %d): %s", service, status, message).
		WithContext("service", service).
		WithContext("status", status)
}

// ErrRateLimit creates a rate limit error
func ErrRateLimit(service string) *AppError {
	return Newf(APIError, "rate limit exceeded for %s", service).
		WithContext("service", service)
}