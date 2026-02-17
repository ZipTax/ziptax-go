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
			_, _ = w.Write([]byte(`{"message":"success"}`))
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
			_, _ = w.Write([]byte(`{
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

func TestClient_Post(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/orders", r.URL.Path)
			assert.Equal(t, "test-tc-key", r.Header.Get("X-API-Key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-agent/1.0", r.Header.Get("User-Agent"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"orderId":"order-123","status":"created"}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		body := map[string]string{"orderId": "order-123"}
		headers := map[string]string{"X-API-Key": "test-tc-key"}
		var result map[string]string
		err := client.Post(context.Background(), server.URL, "/api/orders", headers, body, &result)

		require.NoError(t, err)
		assert.Equal(t, "order-123", result["orderId"])
		assert.Equal(t, "created", result["status"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{
				"title": "Bad Request",
				"status": 400,
				"detail": "invalid order data"
			}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]string
		err := client.Post(context.Background(), server.URL, "/api/orders", nil, nil, &result)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 400, apiErr.StatusCode)
		assert.Equal(t, "Bad Request", apiErr.Name)
		assert.Equal(t, "invalid order data", apiErr.Message)
	})

	t.Run("nil body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]interface{}
		err := client.Post(context.Background(), server.URL, "/test", nil, nil, &result)

		require.NoError(t, err)
	})
}

func TestClient_Patch(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/api/orders/order-123", r.URL.Path)
			assert.Equal(t, "test-tc-key", r.Header.Get("X-API-Key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":"order-123","completedDate":"2024-01-16T10:00:00Z"}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		body := map[string]string{"completedDate": "2024-01-16T10:00:00Z"}
		headers := map[string]string{"X-API-Key": "test-tc-key"}
		var result map[string]string
		err := client.Patch(context.Background(), server.URL, "/api/orders/order-123", headers, body, &result)

		require.NoError(t, err)
		assert.Equal(t, "order-123", result["orderId"])
		assert.Equal(t, "2024-01-16T10:00:00Z", result["completedDate"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{
				"title": "Bad Request",
				"status": 400,
				"detail": "order is already completed"
			}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]string
		err := client.Patch(context.Background(), server.URL, "/api/orders/order-123", nil, nil, &result)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 400, apiErr.StatusCode)
	})
}

func TestClient_GetWithOptions(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/api/orders/order-123", r.URL.Path)
			assert.Equal(t, "test-tc-key", r.Header.Get("X-API-Key"))
			assert.Equal(t, "test-agent/1.0", r.Header.Get("User-Agent"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"orderId":"order-123"}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		headers := map[string]string{"X-API-Key": "test-tc-key"}
		var result map[string]string
		err := client.GetWithOptions(context.Background(), server.URL, "/api/orders/order-123", headers, &result)

		require.NoError(t, err)
		assert.Equal(t, "order-123", result["orderId"])
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{
				"title": "Not Found",
				"status": 404,
				"detail": "order not found"
			}`))
		}))
		defer server.Close()

		httpClient := &http.Client{}
		retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
		client := NewClient(httpClient, "unused", "test-key", "test-agent/1.0", retryPolicy, nil)

		var result map[string]string
		err := client.GetWithOptions(context.Background(), server.URL, "/api/orders/order-123", nil, &result)

		require.Error(t, err)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 404, apiErr.StatusCode)
		assert.Equal(t, "Not Found", apiErr.Name)
		assert.Equal(t, "order not found", apiErr.Message)
	})
}

func TestHandleErrorResponse_TaxCloudFormat(t *testing.T) {
	httpClient := &http.Client{}
	retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
	client := NewClient(httpClient, "https://example.com", "test-key", "test-agent/1.0", retryPolicy, nil)

	t.Run("TaxCloud error format", func(t *testing.T) {
		body := []byte(`{
			"$schema": "https://api.v3.taxcloud.com/schemas/error",
			"title": "Bad Request",
			"status": 400,
			"detail": "order is already completed"
		}`)

		err := client.handleErrorResponse(400, body)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 400, apiErr.StatusCode)
		assert.Equal(t, 400, apiErr.Code)
		assert.Equal(t, "Bad Request", apiErr.Name)
		assert.Equal(t, "order is already completed", apiErr.Message)
	})

	t.Run("ZipTax error format", func(t *testing.T) {
		body := []byte(`{
			"metadata": {
				"response": {
					"code": 111,
					"name": "RESPONSE_CODE_INVALID_HISTORICAL",
					"message": "The provided historical parameter is not in a valid format."
				}
			}
		}`)

		err := client.handleErrorResponse(422, body)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 422, apiErr.StatusCode)
		assert.Equal(t, 111, apiErr.Code)
		assert.Equal(t, "RESPONSE_CODE_INVALID_HISTORICAL", apiErr.Name)
	})

	t.Run("fallback generic error", func(t *testing.T) {
		body := []byte(`something went wrong`)

		err := client.handleErrorResponse(500, body)
		apiErr, ok := err.(*APIError)
		require.True(t, ok)
		assert.Equal(t, 500, apiErr.StatusCode)
		assert.Contains(t, apiErr.Message, "HTTP 500")
	})
}

func TestClient_WithLogger(t *testing.T) {
	var logged []string
	logger := &testLogger{logs: &logged}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	httpClient := &http.Client{}
	retryPolicy := DefaultRetryPolicy(0, 10*time.Millisecond, 100*time.Millisecond)
	client := NewClient(httpClient, server.URL, "test-key", "test-agent/1.0", retryPolicy, logger)

	var result map[string]interface{}
	err := client.Get(context.Background(), "/test", nil, &result)
	require.NoError(t, err)
	assert.Len(t, logged, 2) // Request + Response log lines
}

type testLogger struct {
	logs *[]string
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	*l.logs = append(*l.logs, format)
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
