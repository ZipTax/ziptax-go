package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCheckRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		expected   bool
	}{
		{
			name:     "network error",
			err:      errors.New("network error"),
			expected: true,
		},
		{
			name:       "status 500",
			statusCode: 500,
			expected:   true,
		},
		{
			name:       "status 502",
			statusCode: 502,
			expected:   true,
		},
		{
			name:       "status 503",
			statusCode: 503,
			expected:   true,
		},
		{
			name:       "status 429",
			statusCode: 429,
			expected:   true,
		},
		{
			name:       "status 200",
			statusCode: 200,
			expected:   false,
		},
		{
			name:       "status 404",
			statusCode: 404,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode != 0 {
				resp = &http.Response{StatusCode: tt.statusCode}
			}
			result := DefaultCheckRetry(resp, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExponentialBackoff(t *testing.T) {
	min := 1 * time.Second
	max := 30 * time.Second

	tests := []struct {
		name    string
		attempt int
	}{
		{"attempt 0", 0},
		{"attempt 1", 1},
		{"attempt 2", 2},
		{"attempt 3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait := ExponentialBackoff(tt.attempt, min, max)
			assert.GreaterOrEqual(t, wait, min)
			assert.LessOrEqual(t, wait, max)
		})
	}
}

func TestDoWithRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		policy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		ctx := context.Background()

		resp, err := DoWithRetry(ctx, client, req, policy)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("retry on 500 error", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		policy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		ctx := context.Background()

		resp, err := DoWithRetry(ctx, client, req, policy)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 3, attempts)
		resp.Body.Close()
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		require.NoError(t, err)

		policy := DefaultRetryPolicy(3, 10*time.Millisecond, 100*time.Millisecond)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = DoWithRetry(ctx, client, req, policy)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}