package ziptax

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &Config{
			APIKey: "test-api-key",
		}
		err := config.validate()
		assert.NoError(t, err)
	})

	t.Run("empty API key", func(t *testing.T) {
		config := &Config{
			APIKey: "",
		}
		err := config.validate()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidAPIKey)
	})
}

func TestConfig_applyDefaults(t *testing.T) {
	t.Run("apply all defaults", func(t *testing.T) {
		config := &Config{
			APIKey:     "test-api-key",
			MaxRetries: -1, // Set to -1 to trigger default
		}
		config.applyDefaults()

		assert.Equal(t, DefaultBaseURL, config.BaseURL)
		assert.Equal(t, DefaultTimeout, config.Timeout)
		assert.Equal(t, DefaultMaxRetries, config.MaxRetries)
		assert.Equal(t, DefaultRetryWaitMin, config.RetryWaitMin)
		assert.Equal(t, DefaultRetryWaitMax, config.RetryWaitMax)
		assert.NotNil(t, config.HTTPClient)
		assert.Equal(t, "ziptax-go/1.0.0", config.UserAgent)
	})

	t.Run("zero max retries remains zero", func(t *testing.T) {
		config := &Config{
			APIKey:     "test-api-key",
			MaxRetries: 0, // Explicitly set to 0 to disable retries
		}
		config.applyDefaults()

		assert.Equal(t, 0, config.MaxRetries)
	})

	t.Run("preserve custom values", func(t *testing.T) {
		customClient := &http.Client{Timeout: 60 * time.Second}
		config := &Config{
			APIKey:       "test-api-key",
			BaseURL:      "https://custom.example.com",
			HTTPClient:   customClient,
			Timeout:      60 * time.Second,
			MaxRetries:   5,
			RetryWaitMin: 2 * time.Second,
			RetryWaitMax: 60 * time.Second,
			UserAgent:    "custom-agent/1.0",
		}
		config.applyDefaults()

		assert.Equal(t, "https://custom.example.com", config.BaseURL)
		assert.Equal(t, customClient, config.HTTPClient)
		assert.Equal(t, 60*time.Second, config.Timeout)
		assert.Equal(t, 5, config.MaxRetries)
		assert.Equal(t, 2*time.Second, config.RetryWaitMin)
		assert.Equal(t, 60*time.Second, config.RetryWaitMax)
		assert.Equal(t, "custom-agent/1.0", config.UserAgent)
	})

	t.Run("negative max retries defaults to zero", func(t *testing.T) {
		config := &Config{
			APIKey:     "test-api-key",
			MaxRetries: -1,
		}
		config.applyDefaults()

		assert.Equal(t, DefaultMaxRetries, config.MaxRetries)
	})
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "https://api.zip-tax.com", DefaultBaseURL)
	assert.Equal(t, 30*time.Second, DefaultTimeout)
	assert.Equal(t, 3, DefaultMaxRetries)
	assert.Equal(t, 1*time.Second, DefaultRetryWaitMin)
	assert.Equal(t, 30*time.Second, DefaultRetryWaitMax)
	assert.Equal(t, "X-API-Key", APIKeyHeader)
}
