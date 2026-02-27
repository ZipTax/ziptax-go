package ziptax

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
)

func TestCalculateCart_ZipTax_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/calculate/cart", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req models.CalculateCartRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Len(t, req.Items, 1)
		assert.Equal(t, "customer-453", req.Items[0].CustomerID)
		assert.Len(t, req.Items[0].LineItems, 1)
		assert.Equal(t, "item-1", req.Items[0].LineItems[0].ItemID)
		assert.Equal(t, 10.75, req.Items[0].LineItems[0].Price)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"items": [
				{
					"cartId": "ce4a-test-uuid",
					"customerId": "customer-453",
					"destination": {
						"address": "200 Spectrum Center Dr, Irvine, CA 92618-1905"
					},
					"origin": {
						"address": "323 Washington Ave N, Minneapolis, MN 55401-2427"
					},
					"lineItems": [
						{
							"itemId": "item-1",
							"price": 10.75,
							"quantity": 1.5,
							"tax": {
								"rate": 0.09025,
								"amount": 1.46
							}
						}
					]
				}
			]
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID: "customer-453",
				Currency:   models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{
					Address: "200 Spectrum Center Dr, Irvine, CA 92618-1905",
				},
				Origin: models.CartAddress{
					Address: "323 Washington Ave N, Minneapolis, MN 55401-2427",
				},
				LineItems: []models.CartLineItem{
					{
						ItemID:   "item-1",
						Price:    10.75,
						Quantity: 1.5,
					},
				},
			},
		},
	}

	result, err := client.CalculateCart(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should be a ZipTax response (no TaxCloud credentials configured)
	resp, ok := result.(*models.CalculateCartResponse)
	require.True(t, ok, "expected *models.CalculateCartResponse, got %T", result)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "ce4a-test-uuid", resp.Items[0].CartID)
	assert.Equal(t, "customer-453", resp.Items[0].CustomerID)
	require.Len(t, resp.Items[0].LineItems, 1)
	assert.Equal(t, "item-1", resp.Items[0].LineItems[0].ItemID)
	assert.Equal(t, 10.75, resp.Items[0].LineItems[0].Price)
	assert.Equal(t, 1.5, resp.Items[0].LineItems[0].Quantity)
	assert.Equal(t, 0.09025, resp.Items[0].LineItems[0].Tax.Rate)
	assert.Equal(t, 1.46, resp.Items[0].LineItems[0].Tax.Amount)
}

func TestCalculateCart_ZipTax_WithTaxabilityCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req models.CalculateCartRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.NotNil(t, req.Items[0].LineItems[0].TaxabilityCode)
		assert.Equal(t, int64(0), *req.Items[0].LineItems[0].TaxabilityCode)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items": [{"cartId": "uuid", "customerId": "c", "destination": {"address": "a"}, "origin": {"address": "b"}, "lineItems": [{"itemId": "i", "price": 1, "quantity": 1, "tax": {"rate": 0.05, "amount": 0.05}}]}]}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	tic := int64(0)
	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5, TaxabilityCode: &tic},
				},
			},
		},
	}

	result, err := client.CalculateCart(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	_, ok := result.(*models.CalculateCartResponse)
	require.True(t, ok)
}

func TestCalculateCart_TaxCloud_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/tax/connections/conn-123/carts", r.URL.Path)
		assert.Equal(t, "tc-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify request body has been transformed to TaxCloud format
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req models.TaxCloudCalculateCartRequest
		require.NoError(t, json.Unmarshal(body, &req))
		require.Len(t, req.Items, 1)

		// Verify address was parsed into structured format
		assert.Equal(t, "200 Spectrum Center Dr", req.Items[0].Destination.Line1)
		assert.Equal(t, "Irvine", req.Items[0].Destination.City)
		assert.Equal(t, "CA", req.Items[0].Destination.State)
		assert.Equal(t, "92618-1905", req.Items[0].Destination.Zip)
		require.NotNil(t, req.Items[0].Destination.CountryCode)
		assert.Equal(t, "US", *req.Items[0].Destination.CountryCode)

		assert.Equal(t, "323 Washington Ave N", req.Items[0].Origin.Line1)
		assert.Equal(t, "Minneapolis", req.Items[0].Origin.City)
		assert.Equal(t, "MN", req.Items[0].Origin.State)
		assert.Equal(t, "55401-2427", req.Items[0].Origin.Zip)

		// Verify line item transformation
		require.Len(t, req.Items[0].LineItems, 1)
		assert.Equal(t, int64(0), req.Items[0].LineItems[0].Index)
		assert.Equal(t, "item-1", req.Items[0].LineItems[0].ItemID)
		assert.Equal(t, 10.75, req.Items[0].LineItems[0].Price)
		assert.Equal(t, 1.5, req.Items[0].LineItems[0].Quantity)
		assert.Equal(t, int64(0), req.Items[0].LineItems[0].TIC)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"connectionId": "conn-123",
			"items": [
				{
					"cartId": "ce4a1234-5678-90ab-cdef-1234567890ab",
					"customerId": "customer-453",
					"currency": {"currencyCode": "USD"},
					"deliveredBySeller": false,
					"destination": {
						"line1": "200 Spectrum Center Dr",
						"city": "Irvine",
						"state": "CA",
						"zip": "92618-1905",
						"countryCode": "US"
					},
					"origin": {
						"line1": "323 Washington Ave N",
						"city": "Minneapolis",
						"state": "MN",
						"zip": "55401-2427",
						"countryCode": "US"
					},
					"exemption": {
						"exemptionId": null,
						"isExempt": null
					},
					"lineItems": [
						{
							"index": 0,
							"itemId": "item-1",
							"price": 10.75,
							"quantity": 1.5,
							"tax": {
								"amount": 1.46,
								"rate": 0.0903
							},
							"tic": 0
						}
					]
				}
			],
			"transactionDate": "2024-01-15T09:30:00Z"
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

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID: "customer-453",
				Currency:   models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{
					Address: "200 Spectrum Center Dr, Irvine, CA 92618-1905",
				},
				Origin: models.CartAddress{
					Address: "323 Washington Ave N, Minneapolis, MN 55401-2427",
				},
				LineItems: []models.CartLineItem{
					{
						ItemID:   "item-1",
						Price:    10.75,
						Quantity: 1.5,
					},
				},
			},
		},
	}

	result, err := client.CalculateCart(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should be a TaxCloud response (TaxCloud credentials are configured)
	resp, ok := result.(*models.TaxCloudCalculateCartResponse)
	require.True(t, ok, "expected *models.TaxCloudCalculateCartResponse, got %T", result)
	assert.Equal(t, "conn-123", resp.ConnectionID)
	assert.Equal(t, "2024-01-15T09:30:00Z", resp.TransactionDate)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "ce4a1234-5678-90ab-cdef-1234567890ab", resp.Items[0].CartID)
	assert.Equal(t, "customer-453", resp.Items[0].CustomerID)
	assert.Equal(t, "USD", resp.Items[0].Currency.CurrencyCode)
	assert.False(t, resp.Items[0].DeliveredBySeller)
	assert.Equal(t, "200 Spectrum Center Dr", resp.Items[0].Destination.Line1)
	assert.Equal(t, "Irvine", resp.Items[0].Destination.City)
	require.Len(t, resp.Items[0].LineItems, 1)
	assert.Equal(t, int64(0), resp.Items[0].LineItems[0].Index)
	assert.Equal(t, "item-1", resp.Items[0].LineItems[0].ItemID)
	assert.Equal(t, 0.0903, resp.Items[0].LineItems[0].Tax.Rate)
	assert.Equal(t, 1.46, resp.Items[0].LineItems[0].Tax.Amount)
}

func TestCalculateCart_TaxCloud_WithTaxabilityCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req models.TaxCloudCalculateCartRequest
		require.NoError(t, json.Unmarshal(body, &req))
		// taxabilityCode=42 should be mapped to tic=42
		assert.Equal(t, int64(42), req.Items[0].LineItems[0].TIC)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"connectionId": "c", "items": [{"cartId": "id", "customerId": "c", "currency": {"currencyCode": "USD"}, "deliveredBySeller": false, "destination": {"line1": "a", "city": "b", "state": "CA", "zip": "92618", "countryCode": "US"}, "origin": {"line1": "c", "city": "d", "state": "MN", "zip": "55401", "countryCode": "US"}, "exemption": {}, "lineItems": [{"index": 0, "itemId": "i", "price": 10, "quantity": 1, "tax": {"amount": 0.5, "rate": 0.05}, "tic": 42}]}], "transactionDate": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	tic := int64(42)
	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.00, Quantity: 1.0, TaxabilityCode: &tic},
				},
			},
		},
	}

	result, err := client.CalculateCart(context.Background(), req)
	require.NoError(t, err)
	resp, ok := result.(*models.TaxCloudCalculateCartResponse)
	require.True(t, ok)
	require.NotNil(t, resp.Items[0].LineItems[0].TIC)
	assert.Equal(t, int64(42), *resp.Items[0].LineItems[0].TIC)
}

func TestCalculateCart_TaxCloud_MultipleLineItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req models.TaxCloudCalculateCartRequest
		require.NoError(t, json.Unmarshal(body, &req))
		require.Len(t, req.Items[0].LineItems, 2)
		// Verify auto-generated indices
		assert.Equal(t, int64(0), req.Items[0].LineItems[0].Index)
		assert.Equal(t, int64(1), req.Items[0].LineItems[1].Index)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"connectionId": "c", "items": [{"cartId": "id", "customerId": "c", "currency": {"currencyCode": "USD"}, "deliveredBySeller": false, "destination": {"line1": "a", "city": "b", "state": "CA", "zip": "92618", "countryCode": "US"}, "origin": {"line1": "c", "city": "d", "state": "MN", "zip": "55401", "countryCode": "US"}, "exemption": {}, "lineItems": [{"index": 0, "itemId": "item-1", "price": 10, "quantity": 1, "tax": {"amount": 0.5, "rate": 0.05}}, {"index": 1, "itemId": "item-2", "price": 20, "quantity": 2, "tax": {"amount": 2.0, "rate": 0.05}}]}], "transactionDate": "2024-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.00, Quantity: 1.0},
					{ItemID: "item-2", Price: 20.00, Quantity: 2.0},
				},
			},
		},
	}

	result, err := client.CalculateCart(context.Background(), req)
	require.NoError(t, err)
	resp, ok := result.(*models.TaxCloudCalculateCartResponse)
	require.True(t, ok)
	require.Len(t, resp.Items[0].LineItems, 2)
}

func TestCalculateCart_NilRequest(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	_, err = client.CalculateCart(context.Background(), nil)
	require.Error(t, err)
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "request", validationErr.Field)
	assert.Contains(t, validationErr.Message, "cannot be nil")
}

func TestCalculateCart_EmptyItems(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Message, "exactly 1 cart element")
}

func TestCalculateCart_InvalidCurrency(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "EUR"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Message, "currency.currencyCode must be 'USD'")
}

func TestCalculateCart_InvalidLineItem(t *testing.T) {
	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Message, "itemId is required")
}

func TestCalculateCart_TaxCloud_BadAddress(t *testing.T) {
	// When TaxCloud is configured but the address can't be parsed
	client, err := NewClient(
		"test-api-key",
		WithTaxCloudConnectionID("conn-123"),
		WithTaxCloudAPIKey("tc-api-key"),
	)
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "this is not a valid address format"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)
	var validationErr *ValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Contains(t, validationErr.Message, "failed to transform request for TaxCloud")
}

func TestCalculateCart_ZipTax_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{
			"metadata": {
				"response": {
					"code": 400,
					"name": "BAD_REQUEST",
					"message": "Invalid request format"
				}
			}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Equal(t, "BAD_REQUEST", apiErr.Name)
}

func TestCalculateCart_TaxCloud_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{
			"title": "Forbidden",
			"status": 403,
			"detail": "Invalid API key"
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

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(context.Background(), req)
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 403, apiErr.StatusCode)
	assert.Equal(t, "Forbidden", apiErr.Name)
}

func TestCalculateCart_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient("test-api-key")
	require.NoError(t, err)

	req := &models.CalculateCartRequest{
		Items: []models.CartItem{
			{
				CustomerID:  "customer-453",
				Currency:    models.CartCurrency{CurrencyCode: "USD"},
				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
				LineItems: []models.CartLineItem{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}

	_, err = client.CalculateCart(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// Should NOT be extractable as an APIError
	var apiErr *APIError
	assert.False(t, errors.As(err, &apiErr))
}

func TestTransformToTaxCloudCartRequest(t *testing.T) {
	t.Run("transforms addresses correctly", func(t *testing.T) {
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID: "customer-453",
					Currency:   models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{
						Address: "200 Spectrum Center Dr, Irvine, CA 92618-1905",
					},
					Origin: models.CartAddress{
						Address: "323 Washington Ave N, Minneapolis, MN 55401-2427",
					},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
					},
				},
			},
		}

		tcReq, err := transformToTaxCloudCartRequest(req)
		require.NoError(t, err)
		require.Len(t, tcReq.Items, 1)

		assert.Equal(t, "200 Spectrum Center Dr", tcReq.Items[0].Destination.Line1)
		assert.Equal(t, "Irvine", tcReq.Items[0].Destination.City)
		assert.Equal(t, "CA", tcReq.Items[0].Destination.State)
		assert.Equal(t, "92618-1905", tcReq.Items[0].Destination.Zip)
		require.NotNil(t, tcReq.Items[0].Destination.CountryCode)
		assert.Equal(t, "US", *tcReq.Items[0].Destination.CountryCode)

		assert.Equal(t, "323 Washington Ave N", tcReq.Items[0].Origin.Line1)
		assert.Equal(t, "Minneapolis", tcReq.Items[0].Origin.City)
		assert.Equal(t, "MN", tcReq.Items[0].Origin.State)
		assert.Equal(t, "55401-2427", tcReq.Items[0].Origin.Zip)
	})

	t.Run("maps taxabilityCode to TIC", func(t *testing.T) {
		tic := int64(42)
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID:  "c",
					Currency:    models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
					Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.0, Quantity: 1.0, TaxabilityCode: &tic},
					},
				},
			},
		}

		tcReq, err := transformToTaxCloudCartRequest(req)
		require.NoError(t, err)
		assert.Equal(t, int64(42), tcReq.Items[0].LineItems[0].TIC)
	})

	t.Run("defaults nil taxabilityCode to 0", func(t *testing.T) {
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID:  "c",
					Currency:    models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
					Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.0, Quantity: 1.0},
					},
				},
			},
		}

		tcReq, err := transformToTaxCloudCartRequest(req)
		require.NoError(t, err)
		assert.Equal(t, int64(0), tcReq.Items[0].LineItems[0].TIC)
	})

	t.Run("auto-generates line item indices", func(t *testing.T) {
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID:  "c",
					Currency:    models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
					Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.0, Quantity: 1.0},
						{ItemID: "item-2", Price: 20.0, Quantity: 2.0},
						{ItemID: "item-3", Price: 30.0, Quantity: 3.0},
					},
				},
			},
		}

		tcReq, err := transformToTaxCloudCartRequest(req)
		require.NoError(t, err)
		require.Len(t, tcReq.Items[0].LineItems, 3)
		assert.Equal(t, int64(0), tcReq.Items[0].LineItems[0].Index)
		assert.Equal(t, int64(1), tcReq.Items[0].LineItems[1].Index)
		assert.Equal(t, int64(2), tcReq.Items[0].LineItems[2].Index)
	})

	t.Run("invalid destination address", func(t *testing.T) {
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID:  "c",
					Currency:    models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{Address: "not a valid address"},
					Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.0, Quantity: 1.0},
					},
				},
			},
		}

		_, err := transformToTaxCloudCartRequest(req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "destination address")
	})

	t.Run("invalid origin address", func(t *testing.T) {
		req := &models.CalculateCartRequest{
			Items: []models.CartItem{
				{
					CustomerID:  "c",
					Currency:    models.CartCurrency{CurrencyCode: "USD"},
					Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
					Origin:      models.CartAddress{Address: "not a valid address"},
					LineItems: []models.CartLineItem{
						{ItemID: "item-1", Price: 10.0, Quantity: 1.0},
					},
				},
			},
		}

		_, err := transformToTaxCloudCartRequest(req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "origin address")
	})
}
