package ziptax

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIError_Error(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		underlyingErr := errors.New("connection failed")
		err := &APIError{
			StatusCode: 500,
			Code:       500,
			Name:       "INTERNAL_SERVER_ERROR",
			Message:    "Server error occurred",
			Err:        underlyingErr,
		}
		expected := "ZipTax API error (status 500, code 500): INTERNAL_SERVER_ERROR - Server error occurred: connection failed"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("without underlying error", func(t *testing.T) {
		err := &APIError{
			StatusCode: 400,
			Code:       400,
			Name:       "BAD_REQUEST",
			Message:    "Invalid parameters",
		}
		expected := "ZipTax API error (status 400, code 400): BAD_REQUEST - Invalid parameters"
		assert.Equal(t, expected, err.Error())
	})
}

func TestAPIError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("network error")
	err := &APIError{
		StatusCode: 500,
		Code:       500,
		Name:       "ERROR",
		Message:    "Error occurred",
		Err:        underlyingErr,
	}
	assert.Equal(t, underlyingErr, err.Unwrap())
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "address",
		Value:   "",
		Message: "address cannot be empty",
	}
	expected := "validation error for field 'address': address cannot be empty (value: '')"
	assert.Equal(t, expected, err.Error())
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are defined
	assert.NotNil(t, ErrInvalidAPIKey)
	assert.NotNil(t, ErrRateLimitExceeded)
	assert.NotNil(t, ErrInvalidAddress)
	assert.NotNil(t, ErrInvalidCoordinates)
	assert.NotNil(t, ErrAPIError)
	assert.NotNil(t, ErrNetworkError)
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrContextCanceled)

	// Test error messages
	assert.Contains(t, ErrInvalidAPIKey.Error(), "invalid API key")
	assert.Contains(t, ErrRateLimitExceeded.Error(), "rate limit exceeded")
	assert.Contains(t, ErrInvalidAddress.Error(), "invalid address")
	assert.Contains(t, ErrInvalidCoordinates.Error(), "invalid coordinates")
}
