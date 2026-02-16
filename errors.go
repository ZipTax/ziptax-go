package ziptax

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrInvalidAPIKey indicates that the provided API key is invalid or missing.
	ErrInvalidAPIKey = errors.New("invalid API key")

	// ErrRateLimitExceeded indicates that the API rate limit has been exceeded.
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrInvalidAddress indicates that the provided address is invalid.
	ErrInvalidAddress = errors.New("invalid address")

	// ErrInvalidCoordinates indicates that the provided lat/lng coordinates are invalid.
	ErrInvalidCoordinates = errors.New("invalid coordinates")

	// ErrAPIError indicates a general API error occurred.
	ErrAPIError = errors.New("API error")

	// ErrNetworkError indicates a network-related error occurred.
	ErrNetworkError = errors.New("network error")

	// ErrTimeout indicates that a request timed out.
	ErrTimeout = errors.New("request timeout")

	// ErrContextCanceled indicates that the context was canceled.
	ErrContextCanceled = errors.New("context canceled")

	// ErrTaxCloudNotConfigured indicates that TaxCloud credentials are not configured.
	ErrTaxCloudNotConfigured = errors.New("TaxCloud credentials not configured: both Connection ID and API Key are required for order operations")
)

// APIError represents an error response from the ZipTax API.
type APIError struct {
	StatusCode int    // HTTP status code
	Code       int    // API response code
	Name       string // Response code name
	Message    string // Error message
	Err        error  // Underlying error, if any
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ZipTax API error (status %d, code %d): %s - %s: %v",
			e.StatusCode, e.Code, e.Name, e.Message, e.Err)
	}
	return fmt.Sprintf("ZipTax API error (status %d, code %d): %s - %s",
		e.StatusCode, e.Code, e.Name, e.Message)
}

// Unwrap returns the underlying error, if any.
func (e *APIError) Unwrap() error {
	return e.Err
}

// ValidationError represents an input validation error.
type ValidationError struct {
	Field   string // Field name that failed validation
	Value   string // Field value that failed validation
	Message string // Validation error message
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: '%s')",
		e.Field, e.Message, e.Value)
}
