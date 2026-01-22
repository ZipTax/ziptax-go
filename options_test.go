package ziptax

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBaseURL(t *testing.T) {
	client, err := NewClient("test-api-key", WithBaseURL("https://custom.example.com"))
	require.NoError(t, err)
	assert.Equal(t, "https://custom.example.com", client.config.BaseURL)
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	client, err := NewClient("test-api-key", WithHTTPClient(customClient))
	require.NoError(t, err)
	assert.Equal(t, customClient, client.config.HTTPClient)
}

func TestWithTimeout(t *testing.T) {
	client, err := NewClient("test-api-key", WithTimeout(60*time.Second))
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, client.config.Timeout)
}

func TestWithMaxRetries(t *testing.T) {
	client, err := NewClient("test-api-key", WithMaxRetries(5))
	require.NoError(t, err)
	assert.Equal(t, 5, client.config.MaxRetries)
}

func TestWithRetryWait(t *testing.T) {
	client, err := NewClient("test-api-key", WithRetryWait(2*time.Second, 60*time.Second))
	require.NoError(t, err)
	assert.Equal(t, 2*time.Second, client.config.RetryWaitMin)
	assert.Equal(t, 60*time.Second, client.config.RetryWaitMax)
}

func TestWithLogger(t *testing.T) {
	logger := log.New(os.Stdout, "[test] ", log.LstdFlags)
	client, err := NewClient("test-api-key", WithLogger(logger))
	require.NoError(t, err)
	assert.NotNil(t, client.config.Logger)
}

func TestWithUserAgent(t *testing.T) {
	client, err := NewClient("test-api-key", WithUserAgent("custom-agent/1.0"))
	require.NoError(t, err)
	assert.Equal(t, "custom-agent/1.0", client.config.UserAgent)
}

func TestMultipleOptions(t *testing.T) {
	client, err := NewClient(
		"test-api-key",
		WithBaseURL("https://custom.example.com"),
		WithMaxRetries(5),
		WithTimeout(60*time.Second),
		WithUserAgent("custom-agent/1.0"),
	)
	require.NoError(t, err)
	assert.Equal(t, "https://custom.example.com", client.config.BaseURL)
	assert.Equal(t, 5, client.config.MaxRetries)
	assert.Equal(t, 60*time.Second, client.config.Timeout)
	assert.Equal(t, "custom-agent/1.0", client.config.UserAgent)
}