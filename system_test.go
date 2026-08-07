package ziptax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
)

// newGetTestServer returns a server asserting a GET on the expected path.
func newGetTestServer(t *testing.T, wantPath, responseJSON string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, wantPath, r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
}

func TestClient_GetTICCodes(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := newGetTestServer(t, "/data/tic", `{
			"tic_list": [
				{"tic": {
					"id": "00000", "parent": "", "title": "Uncategorized",
					"label": "Uncategorized items", "nl_title": "Uncategorized",
					"nl_label": "Uncategorized items"
				}},
				{"tic": {
					"id": "11010", "parent": "11000", "title": "Shipping",
					"label": "Shipping charges", "nl_title": "Shipping",
					"nl_label": "Shipping charges"
				}}
			]
		}`)
		defer server.Close()

		catalog, err := newTestClient(t, server.URL).GetTICCodes(context.Background())
		require.NoError(t, err)
		require.Len(t, catalog.TICList, 2)
		assert.Equal(t, "00000", catalog.TICList[0].TIC.ID)
		assert.Equal(t, "Shipping", catalog.TICList[1].TIC.Title)
		assert.Equal(t, "11000", catalog.TICList[1].TIC.Parent)
		assert.Equal(t, "Shipping charges", catalog.TICList[1].TIC.NLLabel)
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"title":"Internal Server Error","status":500}`))
		}))
		defer server.Close()

		_, err := newTestClient(t, server.URL).GetTICCodes(context.Background())
		assert.Error(t, err)
	})
}

func TestClient_GetTICSearchSchema(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := newGetTestServer(t, "/schemas/ticsearch",
			`{"$id":"https://api.zip-tax.com/schemas/ticsearch","type":"object"}`)
		defer server.Close()

		schema, err := newTestClient(t, server.URL).GetTICSearchSchema(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "https://api.zip-tax.com/schemas/ticsearch", schema["$id"])
		assert.Equal(t, "object", schema["type"])
	})
}

func TestClient_GetHealth(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		server := newGetTestServer(t, "/system/health", `{
			"status": "ok",
			"components": {"dynamo": "ok", "taxdata": "ok", "taxdata_count": 1234567}
		}`)
		defer server.Close()

		health, err := newTestClient(t, server.URL).GetHealth(context.Background())
		require.NoError(t, err)
		assert.True(t, health.IsHealthy())
		assert.Equal(t, int64(1234567), health.Components.TaxDataCount)
	})

	t.Run("degraded component makes it unhealthy", func(t *testing.T) {
		server := newGetTestServer(t, "/system/health", `{
			"status": "ok",
			"components": {"dynamo": "ok", "taxdata": "partial", "taxdata_count": 10}
		}`)
		defer server.Close()

		health, err := newTestClient(t, server.URL).GetHealth(context.Background())
		require.NoError(t, err)
		assert.False(t, health.IsHealthy())
		assert.Equal(t, models.HealthStatusPartial, health.Components.TaxData)
	})

	t.Run("dynamo connection error makes it unhealthy", func(t *testing.T) {
		server := newGetTestServer(t, "/system/health", `{
			"status": "ok",
			"components": {"dynamo": "connection_error", "taxdata": "ok", "taxdata_count": 10}
		}`)
		defer server.Close()

		health, err := newTestClient(t, server.URL).GetHealth(context.Background())
		require.NoError(t, err)
		assert.False(t, health.IsHealthy())
	})
}

func TestClient_GetSystemMetadata(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := newGetTestServer(t, "/system/metadata",
			`{"go_version":"go1.23.4","hostname":"api-7c9d5f8b6-xk2ql"}`)
		defer server.Close()

		meta, err := newTestClient(t, server.URL).GetSystemMetadata(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "go1.23.4", meta.GoVersion)
		assert.Equal(t, "api-7c9d5f8b6-xk2ql", meta.Hostname)
	})
}

func TestClient_GetDetailedAccountMetrics(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := newGetTestServer(t, "/account/metrics", `{
			"core_request_count": 15595,
			"core_request_limit": 1000000,
			"core_usage_percent": 1.5595,
			"geo_enabled": true,
			"geo_request_count": 43891,
			"geo_request_limit": 1000000,
			"geo_usage_percent": 4.3891,
			"merchant_request_count": 120,
			"merchant_request_limit": 5000,
			"merchant_usage_percent": 2.4,
			"is_active": true,
			"message": "Contact support@zip.tax to modify your account"
		}`)
		defer server.Close()

		metrics, err := newTestClient(t, server.URL).GetDetailedAccountMetrics(context.Background())
		require.NoError(t, err)
		assert.Equal(t, int64(15595), metrics.CoreRequestCount)
		assert.Equal(t, 4.3891, metrics.GeoUsagePercent)
		assert.Equal(t, int64(120), metrics.MerchantRequestCount)
		assert.Equal(t, int64(5000), metrics.MerchantRequestLimit)
		assert.Equal(t, 2.4, metrics.MerchantUsagePercent)
		assert.True(t, metrics.GeoEnabled)
		assert.True(t, metrics.IsActive)
	})

	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"title":"Unauthorized","status":401}`))
		}))
		defer server.Close()

		_, err := newTestClient(t, server.URL).GetDetailedAccountMetrics(context.Background())
		assert.Error(t, err)
		var apiErr *APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	})
}
