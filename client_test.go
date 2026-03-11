package ziptax

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
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
			assert.Equal(t, "202401", r.URL.Query().Get("historical"))
			assert.Equal(t, "USA", r.URL.Query().Get("countryCode"))
			assert.Equal(t, "json", r.URL.Query().Get("format"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
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
			WithHistorical("202401"),
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
			_, _ = w.Write([]byte(`{
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
			_, _ = w.Write([]byte(`{
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
			_, _ = w.Write([]byte(`{
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

func TestClient_GetRatesByPostalCode(t *testing.T) {
	t.Run("successful request with 5-digit postal code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			assert.Equal(t, "92694", r.URL.Query().Get("postalcode"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"version": "v60",
				"rCode": 100,
				"results": [
					{
						"geoPostalCode": "92694",
						"geoCity": "LADERA RANCH",
						"geoCounty": "ORANGE",
						"geoState": "CA",
						"taxSales": 0.0775,
						"taxUse": 0.0775,
						"txbService": "N",
						"txbFreight": "N",
						"stateSalesTax": 0.06,
						"stateUseTax": 0.06,
						"citySalesTax": 0,
						"cityUseTax": 0,
						"cityTaxCode": "",
						"countySalesTax": 0.0025,
						"countyUseTax": 0.0025,
						"countyTaxCode": "",
						"districtSalesTax": 0.015,
						"districtUseTax": 0.015,
						"district1Code": "37",
						"district1SalesTax": 0,
						"district1UseTax": 0,
						"district2Code": "37",
						"district2SalesTax": 0.005,
						"district2UseTax": 0.005,
						"district3Code": "",
						"district3SalesTax": 0,
						"district3UseTax": 0,
						"district4Code": "30",
						"district4SalesTax": 0.01,
						"district4UseTax": 0.01,
						"district5Code": "",
						"district5SalesTax": 0,
						"district5UseTax": 0,
						"originDestination": "D"
					},
					{
						"geoPostalCode": "92694",
						"geoCity": "SAN JUAN CAPISTRANO",
						"geoCounty": "ORANGE",
						"geoState": "CA",
						"taxSales": 0.0775,
						"taxUse": 0.0775,
						"txbService": "N",
						"txbFreight": "N",
						"stateSalesTax": 0.06,
						"stateUseTax": 0.06,
						"citySalesTax": 0,
						"cityUseTax": 0,
						"cityTaxCode": "",
						"countySalesTax": 0.0025,
						"countyUseTax": 0.0025,
						"countyTaxCode": "",
						"districtSalesTax": 0.015,
						"districtUseTax": 0.015,
						"district1Code": "37",
						"district1SalesTax": 0,
						"district1UseTax": 0,
						"district2Code": "37",
						"district2SalesTax": 0.005,
						"district2UseTax": 0.005,
						"district3Code": "",
						"district3SalesTax": 0,
						"district3UseTax": 0,
						"district4Code": "30",
						"district4SalesTax": 0.01,
						"district4UseTax": 0.01,
						"district5Code": "",
						"district5SalesTax": 0,
						"district5UseTax": 0,
						"originDestination": "D"
					}
				],
				"addressDetail": {
					"normalizedAddress": "feature available for geo address lookups only",
					"incorporated": "feature available for geo address lookups only",
					"geoLat": 0,
					"geoLng": 0
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetRatesByPostalCode(context.Background(), "92694")
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "v60", response.Version)
		assert.Equal(t, 100, response.RCode)
		assert.Len(t, response.Results, 2)
		assert.Equal(t, "LADERA RANCH", response.Results[0].GeoCity)
		assert.Equal(t, "SAN JUAN CAPISTRANO", response.Results[1].GeoCity)
		assert.Equal(t, 0.0775, response.Results[0].TaxSales)
		assert.Equal(t, "D", response.Results[0].OriginDestination)
	})

	t.Run("successful request with 9-digit postal code (stripped to 5-digit)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			// 9-digit postal codes are normalized to 5-digit before sending to the API
			assert.Equal(t, "92694", r.URL.Query().Get("postalcode"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"version": "v60",
				"rCode": 100,
				"results": [
					{
						"geoPostalCode": "92694",
						"geoCity": "LADERA RANCH",
						"geoCounty": "ORANGE",
						"geoState": "CA",
						"taxSales": 0.0775,
						"taxUse": 0.0775,
						"txbService": "N",
						"txbFreight": "N",
						"stateSalesTax": 0.06,
						"stateUseTax": 0.06,
						"citySalesTax": 0,
						"cityUseTax": 0,
						"cityTaxCode": "",
						"countySalesTax": 0.0025,
						"countyUseTax": 0.0025,
						"countyTaxCode": "",
						"districtSalesTax": 0.015,
						"districtUseTax": 0.015,
						"district1Code": "37",
						"district1SalesTax": 0,
						"district1UseTax": 0,
						"district2Code": "37",
						"district2SalesTax": 0.005,
						"district2UseTax": 0.005,
						"district3Code": "",
						"district3SalesTax": 0,
						"district3UseTax": 0,
						"district4Code": "30",
						"district4SalesTax": 0.01,
						"district4UseTax": 0.01,
						"district5Code": "",
						"district5SalesTax": 0,
						"district5UseTax": 0,
						"originDestination": "D"
					}
				],
				"addressDetail": {
					"normalizedAddress": "feature available for geo address lookups only",
					"incorporated": "feature available for geo address lookups only",
					"geoLat": 0,
					"geoLng": 0
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetRatesByPostalCode(context.Background(), "92694-1234")
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "v60", response.Version)
		assert.Len(t, response.Results, 1)
	})

	t.Run("successful request with format option", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/request/v60", r.URL.Path)
			assert.Equal(t, "92694", r.URL.Query().Get("postalcode"))
			assert.Equal(t, "json", r.URL.Query().Get("format"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"version": "v60",
				"rCode": 100,
				"results": [],
				"addressDetail": {
					"normalizedAddress": "",
					"incorporated": "",
					"geoLat": 0,
					"geoLng": 0
				}
			}`))
		}))
		defer server.Close()

		client, err := NewClient("test-api-key", WithBaseURL(server.URL))
		require.NoError(t, err)

		response, err := client.GetRatesByPostalCode(
			context.Background(),
			"92694",
			WithFormat("json"),
		)
		require.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("empty postal code", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		_, err = client.GetRatesByPostalCode(context.Background(), "")
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "postalcode", validationErr.Field)
	})

	t.Run("invalid postal code format", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		_, err = client.GetRatesByPostalCode(context.Background(), "1234")
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "postalcode", validationErr.Field)
	})

	t.Run("postal code with letters", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		require.NoError(t, err)

		_, err = client.GetRatesByPostalCode(context.Background(), "9269A")
		require.Error(t, err)
		var validationErr *ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "postalcode", validationErr.Field)
	})
}

func TestCreateOrder_NoTaxCloudCredentials(t *testing.T) {
	// Create client without TaxCloud credentials
	client, err := NewClient("test-api-key")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Try to create an order
	ctx := context.Background()
	orderReq := &models.CreateOrderRequest{
		OrderID:         "test-order-123",
		CustomerID:      "customer-456",
		TransactionDate: "2024-01-15T09:30:00Z",
		CompletedDate:   "2024-01-15T09:30:00Z",
		Origin: models.TaxCloudAddress{
			Line1: "323 Washington Ave N",
			City:  "Minneapolis",
			State: "MN",
			Zip:   "55401-2427",
		},
		Destination: models.TaxCloudAddress{
			Line1: "323 Washington Ave N",
			City:  "Minneapolis",
			State: "MN",
			Zip:   "55401-2427",
		},
		LineItems: []models.CartItemWithTax{
			{
				Index:    0,
				ItemID:   "item-1",
				Price:    10.8,
				Quantity: 1.5,
				Tax: models.Tax{
					Amount: 1.31,
					Rate:   0.0813,
				},
			},
		},
		Currency: &models.Currency{},
	}

	_, err = client.CreateOrder(ctx, orderReq)

	// Should return ErrTaxCloudNotConfigured
	if err == nil {
		t.Fatal("Expected error when TaxCloud credentials not configured, got nil")
	}

	if !errors.Is(err, ErrTaxCloudNotConfigured) {
		t.Errorf("Expected ErrTaxCloudNotConfigured, got: %v", err)
	}
}

func TestCreateOrder_WithCredentials(t *testing.T) {
	// Create client with TaxCloud credentials
	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("test-connection-id"),
		WithTaxCloudAPIKey("test-taxcloud-key"),
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Verify credentials are configured
	if !client.config.HasTaxCloudCredentials() {
		t.Fatal("TaxCloud credentials should be configured")
	}

	// Verify the config values are set correctly
	if client.config.TaxCloudConnectionID != "test-connection-id" {
		t.Errorf("Expected connection ID 'test-connection-id', got '%s'", client.config.TaxCloudConnectionID)
	}

	if client.config.TaxCloudAPIKey != "test-taxcloud-key" {
		t.Errorf("Expected API key 'test-taxcloud-key', got '%s'", client.config.TaxCloudAPIKey)
	}

	if client.config.TaxCloudBaseURL != DefaultTaxCloudBaseURL {
		t.Errorf("Expected base URL '%s', got '%s'", DefaultTaxCloudBaseURL, client.config.TaxCloudBaseURL)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/orders", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"orderId": "order-123",
			"customerId": "customer-456",
			"connectionId": "conn-123",
			"transactionDate": "2024-01-15T09:30:00Z",
			"completedDate": "2024-01-15T09:30:00Z",
			"origin": {
				"line1": "200 Spectrum Center Drive",
				"city": "Irvine",
				"state": "CA",
				"zip": "92618",
				"countryCode": "US"
			},
			"destination": {
				"line1": "323 Washington Ave N",
				"city": "Minneapolis",
				"state": "MN",
				"zip": "55401-2427",
				"countryCode": "US"
			},
			"lineItems": [
				{
					"index": 0,
					"itemId": "item-1",
					"price": 10.8,
					"quantity": 1.5,
					"tax": {"amount": 1.31, "rate": 0.0813},
					"tic": 0
				}
			],
			"currency": {"currencyCode": "USD"},
			"deliveredBySeller": false,
			"excludeFromFiling": false
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	orderReq := &models.CreateOrderRequest{
		OrderID:         "order-123",
		CustomerID:      "customer-456",
		TransactionDate: "2024-01-15T09:30:00Z",
		CompletedDate:   "2024-01-15T09:30:00Z",
		Origin: models.TaxCloudAddress{
			Line1: "200 Spectrum Center Drive",
			City:  "Irvine",
			State: "CA",
			Zip:   "92618",
		},
		Destination: models.TaxCloudAddress{
			Line1: "323 Washington Ave N",
			City:  "Minneapolis",
			State: "MN",
			Zip:   "55401-2427",
		},
		LineItems: []models.CartItemWithTax{
			{
				Index:    0,
				ItemID:   "item-1",
				Price:    10.8,
				Quantity: 1.5,
				Tax:      models.Tax{Amount: 1.31, Rate: 0.0813},
			},
		},
		Currency: &models.Currency{},
	}

	response, err := client.CreateOrder(context.Background(), orderReq)
	require.NoError(t, err)
	assert.Equal(t, "order-123", response.OrderID)
	assert.Equal(t, "customer-456", response.CustomerID)
	assert.Equal(t, "conn-123", response.ConnectionID)
	assert.Len(t, response.LineItems, 1)
	assert.Equal(t, 1.31, response.LineItems[0].Tax.Amount)
}

func TestCreateOrder_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"$schema": "https://api.v3.taxcloud.com/schemas/error",
			"title": "Bad Request",
			"status": 400,
			"detail": "order is already completed"
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	orderReq := &models.CreateOrderRequest{
		OrderID:    "order-123",
		CustomerID: "customer-456",
	}

	_, err = client.CreateOrder(context.Background(), orderReq)
	require.Error(t, err)

	// Verify the error can be extracted as a public APIError
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "Bad Request", apiErr.Name)
	assert.Equal(t, "order is already completed", apiErr.Message)
}

func TestGetOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/orders/order-456", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"orderId": "order-456",
			"customerId": "customer-789",
			"connectionId": "conn-123",
			"transactionDate": "2024-01-15T09:30:00Z",
			"completedDate": "2024-01-15T09:30:00Z",
			"origin": {
				"line1": "200 Spectrum Center Drive",
				"city": "Irvine",
				"state": "CA",
				"zip": "92618",
				"countryCode": "US"
			},
			"destination": {
				"line1": "323 Washington Ave N",
				"city": "Minneapolis",
				"state": "MN",
				"zip": "55401-2427",
				"countryCode": "US"
			},
			"lineItems": [
				{
					"index": 0,
					"itemId": "item-1",
					"price": 10.8,
					"quantity": 1.5,
					"tax": {"amount": 1.31, "rate": 0.0813},
					"tic": 0
				}
			],
			"currency": {"currencyCode": "USD"},
			"deliveredBySeller": false,
			"excludeFromFiling": false
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	response, err := client.GetOrder(context.Background(), "order-456")
	require.NoError(t, err)
	assert.Equal(t, "order-456", response.OrderID)
	assert.Equal(t, "customer-789", response.CustomerID)
	assert.Len(t, response.LineItems, 1)
}

func TestGetOrder_NoCredentials(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	_, err = client.GetOrder(context.Background(), "order-123")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTaxCloudNotConfigured)
}

func TestGetOrder_APIError(t *testing.T) {
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

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	_, err = client.GetOrder(context.Background(), "nonexistent-order")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 404, apiErr.StatusCode)
}

func TestUpdateOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/orders/order-456", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"orderId": "order-456",
			"customerId": "customer-789",
			"connectionId": "conn-123",
			"transactionDate": "2024-01-15T09:30:00Z",
			"completedDate": "2024-01-16T10:00:00Z",
			"origin": {
				"line1": "200 Spectrum Center Drive",
				"city": "Irvine",
				"state": "CA",
				"zip": "92618",
				"countryCode": "US"
			},
			"destination": {
				"line1": "323 Washington Ave N",
				"city": "Minneapolis",
				"state": "MN",
				"zip": "55401-2427",
				"countryCode": "US"
			},
			"lineItems": [],
			"currency": {"currencyCode": "USD"},
			"deliveredBySeller": false,
			"excludeFromFiling": false
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	updateReq := &models.UpdateOrderRequest{
		CompletedDate: "2024-01-16T10:00:00Z",
	}

	response, err := client.UpdateOrder(context.Background(), "order-456", updateReq)
	require.NoError(t, err)
	assert.Equal(t, "order-456", response.OrderID)
	assert.Equal(t, "2024-01-16T10:00:00Z", response.CompletedDate)
}

func TestUpdateOrder_NoCredentials(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	updateReq := &models.UpdateOrderRequest{CompletedDate: "2024-01-16T10:00:00Z"}
	_, err = client.UpdateOrder(context.Background(), "order-123", updateReq)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTaxCloudNotConfigured)
}

func TestUpdateOrder_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"title": "Bad Request",
			"status": 400,
			"detail": "order is already completed and cannot be updated"
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	updateReq := &models.UpdateOrderRequest{CompletedDate: "2024-01-16T10:00:00Z"}
	_, err = client.UpdateOrder(context.Background(), "order-456", updateReq)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "order is already completed and cannot be updated", apiErr.Message)
}

func TestRefundOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/orders/refunds/order-456", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`[
			{
				"connectionId": "conn-123",
				"createdDate": "2024-01-17T12:00:00Z",
				"items": [
					{
						"index": 0,
						"itemId": "item-1",
						"price": 10.8,
						"quantity": 1.0,
						"tax": {"amount": 0.87}
					}
				]
			}
		]`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	refundReq := &models.RefundTransactionRequest{
		Items: []models.CartItemRefundWithTaxRequest{
			{
				ItemID:   "item-1",
				Quantity: 1.0,
			},
		},
	}

	refunds, err := client.RefundOrder(context.Background(), "order-456", refundReq)
	require.NoError(t, err)
	require.Len(t, refunds, 1)
	assert.Equal(t, "conn-123", refunds[0].ConnectionID)
	require.Len(t, refunds[0].Items, 1)
	assert.Equal(t, "item-1", refunds[0].Items[0].ItemID)
	assert.Equal(t, 0.87, refunds[0].Items[0].Tax.Amount)
}

func TestRefundOrder_FullRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`[
			{
				"connectionId": "conn-123",
				"createdDate": "2024-01-17T12:00:00Z",
				"items": [
					{
						"index": 0,
						"itemId": "item-1",
						"price": 10.8,
						"quantity": 1.5,
						"tax": {"amount": 1.31}
					}
				]
			}
		]`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	// Empty request for full refund
	refundReq := &models.RefundTransactionRequest{}

	refunds, err := client.RefundOrder(context.Background(), "order-456", refundReq)
	require.NoError(t, err)
	require.Len(t, refunds, 1)
}

func TestRefundOrder_NoCredentials(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	refundReq := &models.RefundTransactionRequest{}
	_, err = client.RefundOrder(context.Background(), "order-123", refundReq)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTaxCloudNotConfigured)
}

func TestRefundOrder_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"title": "Bad Request",
			"status": 400,
			"detail": "order has already been refunded"
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	refundReq := &models.RefundTransactionRequest{}
	_, err = client.RefundOrder(context.Background(), "order-456", refundReq)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
}

func TestWrapError_WithAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{
			"metadata": {
				"response": {
					"code": 401,
					"name": "UNAUTHORIZED",
					"message": "Invalid API key"
				}
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("bad-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	_, err = client.GetSalesTaxByAddress(context.Background(), "200 Spectrum Center Dr")
	require.Error(t, err)

	// Verify the public APIError can be extracted through the wrap
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.StatusCode)
	assert.Equal(t, 401, apiErr.Code)
	assert.Equal(t, "UNAUTHORIZED", apiErr.Name)
	assert.Equal(t, "Invalid API key", apiErr.Message)
}

func TestWrapError_WithNonAPIError(t *testing.T) {
	// Test wrapError with a non-API error (e.g., context canceled)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	_, err = client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Dr")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get sales tax by address")
	assert.Contains(t, err.Error(), "context canceled")

	// Should NOT be extractable as an APIError
	var apiErr *APIError
	assert.False(t, errors.As(err, &apiErr))
}

func TestCreateOrderFromCart_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/carts/orders", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"orderId": "my-order-1",
			"customerId": "customer-456",
			"connectionId": "conn-123",
			"transactionDate": "2024-01-15T09:30:00Z",
			"completedDate": "2024-01-15T09:30:00Z",
			"origin": {
				"line1": "200 Spectrum Center Drive",
				"city": "Irvine",
				"state": "CA",
				"zip": "92618",
				"countryCode": "US"
			},
			"destination": {
				"line1": "323 Washington Ave N",
				"city": "Minneapolis",
				"state": "MN",
				"zip": "55401-2427",
				"countryCode": "US"
			},
			"lineItems": [
				{
					"index": 0,
					"itemId": "item-1",
					"price": 10.8,
					"quantity": 1.5,
					"tax": {"amount": 1.31, "rate": 0.0813},
					"tic": 0
				}
			],
			"currency": {"currencyCode": "USD"},
			"deliveredBySeller": false,
			"excludeFromFiling": false
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	req := &models.CreateOrderFromCartRequest{
		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
		OrderID: "my-order-1",
	}

	response, err := client.CreateOrderFromCart(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, "my-order-1", response.OrderID)
	assert.Equal(t, "customer-456", response.CustomerID)
	assert.Equal(t, "conn-123", response.ConnectionID)
	assert.Len(t, response.LineItems, 1)
	assert.Equal(t, 1.31, response.LineItems[0].Tax.Amount)
}

func TestCreateOrderFromCart_NoCredentials(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	req := &models.CreateOrderFromCartRequest{
		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
		OrderID: "my-order-1",
	}

	_, err = client.CreateOrderFromCart(context.Background(), req)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTaxCloudNotConfigured)
}

func TestCreateOrderFromCart_EmptyCartID(t *testing.T) {
	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
	)
	require.NoError(t, err)

	req := &models.CreateOrderFromCartRequest{
		CartID:  "",
		OrderID: "my-order-1",
	}

	_, err = client.CreateOrderFromCart(context.Background(), req)
	require.Error(t, err)

	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Equal(t, "CreateOrderFromCartRequest", valErr.Field)
	assert.Contains(t, valErr.Message, "cartId is required")
}

func TestCreateOrderFromCart_EmptyOrderID(t *testing.T) {
	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
	)
	require.NoError(t, err)

	req := &models.CreateOrderFromCartRequest{
		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
		OrderID: "",
	}

	_, err = client.CreateOrderFromCart(context.Background(), req)
	require.Error(t, err)

	var valErr *ValidationError
	require.ErrorAs(t, err, &valErr)
	assert.Equal(t, "CreateOrderFromCartRequest", valErr.Field)
	assert.Contains(t, valErr.Message, "orderId is required")
}

func TestCreateOrderFromCart_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"$schema": "https://api.v3.taxcloud.com/schemas/error",
			"title": "Bad Request",
			"status": 400,
			"detail": "cart has already been converted to an order"
		}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	req := &models.CreateOrderFromCartRequest{
		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
		OrderID: "my-order-1",
	}

	_, err = client.CreateOrderFromCart(context.Background(), req)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "Bad Request", apiErr.Name)
	assert.Equal(t, "cart has already been converted to an order", apiErr.Message)
}

func TestWithTaxCloudBaseURL(t *testing.T) {
	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL("https://custom-taxcloud.example.com"),
	)
	require.NoError(t, err)
	assert.Equal(t, "https://custom-taxcloud.example.com", client.config.TaxCloudBaseURL)
}
