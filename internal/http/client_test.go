package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	httpClient := &http.Client{}
	retryPolicy := DefaultRetryPolicy(3, 1*time.Second, 30*time.Second)

	client := NewClient(httpClient, "https://api.example.com", "test-key", "test-agent/1.0", retryPolicy, nil)

	assert.NotNil(t, client)
	assert.Equal(t, httpClient, client.HTTPClient)
	assert.Equal(t, "https://api.example.com", client.BaseURL)
	assert.Equal(t, "test-key", client.APIKey)
	assert.Equal(t, "test-agent/1.0", client.UserAgent)
	assert.Equal(t, retryPolicy, client.RetryPolicy)
}

func TestClient_Get(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify headers
			assert.Equal(t, "test-key", r.Header.Get("X-API-Key"))
			assert.Equal(t, "test-agent/1.0", r.Header.Get("User-Agent"))
			assert.Equal(t, "application/json", r.Header.Get("Accept"))

			// Verify query params
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"success"}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, server.URL, "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]string
		err := client.Get(context.Background(), "/test", map[string]string{
			"param1": "value1",
			"param2": "value2",
		}, &result)

		require.NoError(t, err)
		assert.Equal(t, "success", result["message"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{
				"metadata": {
					"response": {
						"code": 400,
						"name": "BAD_REQUEST",
						"message": "Invalid parameters"
					}
				}
			}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, server.URL, "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]string
		err := client.Get(context.Background(), "/test", nil, &result)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 400, apiErr.StatusCode)
		assert.Equal(t, 400, apiErr.Code)
		assert.Equal(t, "BAD_REQUEST", apiErr.Name)
		assert.Equal(t, "Invalid parameters", apiErr.Message)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, server.URL, "test-key", "test-agent/1.0", retryPolicy, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var result map[string]string
		err := client.Get(ctx, "/test", nil, &result)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestAPIError_Error(t *testing.T) {
	t.Run("with code", func(t *testing.T) {
		err := &APIError{
			StatusCode: 400,
			Code:       400,
			Name:       "BAD_REQUEST",
			Message:    "Invalid parameters",
		}
		expected := "API error (status 400, code 400): BAD_REQUEST - Invalid parameters"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("without code", func(t *testing.T) {
		err := &APIError{
			StatusCode: 500,
			Message:    "Internal Server Error",
		}
		expected := "API error (status 500): Internal Server Error"
		assert.Equal(t, expected, err.Error())
	})
}