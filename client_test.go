package ziptax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("valid API key", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("empty API key", func(t *testing.T) {
		client, err := NewClient("")
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrInvalidAPIKey)
	})

	t.Run("with options", func(t *testing.T) {
		client, err := NewClient("test-api-key",
			WithBaseURL("https://custom.example.com"),
			WithMaxRetries(5),
		)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://custom.example.com", client.config.BaseURL)
		assert.Equal(t, 5, client.config.MaxRetries)
	})
}

func TestClient_GetSalesTaxByAddress(t *testing.T) {
	t.Run("successful request with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			assert.Equal(t, "200 Spectrum Center Dr", r.URL.Query().Get("address"))
			assert.Equal(t, "2024-01", r.URL.Query().Get("historical"))
			assert.Equal(t, "USA", r.URL.Query().Get("countryCode"))
			assert.Equal(t, "json", r.URL.Query().Get("format"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"metadata": {
					"version": "v60",
					"response": {
						"code": 100,
						"name": "RESPONSE_CODE_SUCCESS",
						"message": "Successful API Request.",
						"definition": "http://api.zip-tax.com/request/v60/schema"
					}
				},
				"baseRates": [],
				"service": {
					"adjustmentType": "SERVICE_TAXABLE",
					"taxable": "N",
					"description": "Services non-taxable"
				},
				"shipping": {
					"adjustmentType": "FREIGHT_TAXABLE",
					"taxable": "N",
					"description": "Freight non-taxable"
				},
				"addressDetail": {
					"normalizedAddress": "200 Spectrum Center Dr, Irvine, CA 92618",
					"incorporated": "true",
					"geoLat": 33.65253,
					"geoLng": -117.74794
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetSalesTaxByAddress(
			context.Background(),
			"200 Spectrum Center Dr",
			WithHistorical("2024-01"),
			WithCountryCode("USA"),
			WithFormat("json"),
		)
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			assert.Equal(t, "200 Spectrum Center Dr", r.URL.Query().Get("address"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"metadata": {
					"version": "v60",
					"response": {
						"code": 100,
						"name": "RESPONSE_CODE_SUCCESS",
						"message": "Successful API Request.",
						"definition": "http://api.zip-tax.com/request/v60/schema"
					}
				},
				"baseRates": [
					{
						"rate": 0.06,
						"jurType": "US_STATE_SALES_TAX",
						"jurName": "CA"
					}
				],
				"service": {
					"adjustmentType": "SERVICE_TAXABLE",
					"taxable": "N",
					"description": "Services non-taxable"
				},
				"shipping": {
					"adjustmentType": "FREIGHT_TAXABLE",
					"taxable": "N",
					"description": "Freight non-taxable"
				},
				"taxSummaries": [
					{
						"rate": 0.0775,
						"taxType": "SALES_TAX",
						"summaryName": "Total Base Sales Tax",
						"displayRates": [
							{
								"name": "Total Rate",
								"rate": 0.0775
							}
						]
					}
				],
				"addressDetail": {
					"normalizedAddress": "200 Spectrum Center Dr, Irvine, CA 92618",
					"incorporated": "true",
					"geoLat": 33.65253,
					"geoLng": -117.74794
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetSalesTaxByAddress(context.Background(), "200 Spectrum Center Dr")
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "v60", response.Metadata.Version)
		assert.Equal(t, 100, response.Metadata.Response.Code)
		assert.Len(t, response.BaseRates, 1)
		assert.Equal(t, 0.06, response.BaseRates[0].Rate)
		assert.Len(t, response.TaxSummaries, 1)
		assert.Equal(t, 0.0775, response.TaxSummaries[0].Rate)
	})

	t.Run("empty address", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		_, err = client.GetSalesTaxByAddress(context.Background(), "")
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "address", validationErr.Field)
	})

	t.Run("address too long", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		longAddress := string(make([]byte, 101))
		_, err = client.GetSalesTaxByAddress(context.Background(), longAddress)
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
	})
}

func TestClient_GetSalesTaxByGeoLocation(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			assert.Equal(t, "33.65253", r.URL.Query().Get("lat"))
			assert.Equal(t, "-117.74794", r.URL.Query().Get("lng"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"metadata": {
					"version": "v60",
					"response": {
						"code": 100,
						"name": "RESPONSE_CODE_SUCCESS",
						"message": "Successful API Request.",
						"definition": "http://api.zip-tax.com/request/v60/schema"
					}
				},
				"baseRates": [],
				"service": {
					"adjustmentType": "SERVICE_TAXABLE",
					"taxable": "N",
					"description": "Services non-taxable"
				},
				"shipping": {
					"adjustmentType": "FREIGHT_TAXABLE",
					"taxable": "N",
					"description": "Freight non-taxable"
				},
				"addressDetail": {
					"normalizedAddress": "Irvine, CA",
					"incorporated": "true",
					"geoLat": 33.65253,
					"geoLng": -117.74794
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetSalesTaxByGeoLocation(context.Background(), "33.65253", "-117.74794")
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "v60", response.Metadata.Version)
		assert.Equal(t, 33.65253, response.AddressDetail.GeoLat)
		assert.Equal(t, -117.74794, response.AddressDetail.GeoLng)
	})

	t.Run("empty coordinates", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		_, err = client.GetSalesTaxByGeoLocation(context.Background(), "", "")
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
	})
}

func TestClient_GetAccountMetrics(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/account/v60/metrics", r.URL.Path)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"core_request_count": 15595,
				"core_request_limit": 1000000,
				"core_usage_percent": 1.5595,
				"geo_enabled": true,
				"geo_request_count": 43891,
				"geo_request_limit": 1000000,
				"geo_usage_percent": 4.3891,
				"is_active": true,
				"message": "Contact support@zip.tax to modify your account"
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		metrics, err := client.GetAccountMetrics(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(15595), metrics.CoreRequestCount)
		assert.Equal(t, int64(1000000), metrics.CoreRequestLimit)
		assert.Equal(t, 1.5595, metrics.CoreUsagePercent)
		assert.True(t, metrics.GeoEnabled)
		assert.True(t, metrics.IsActive)
	})
}