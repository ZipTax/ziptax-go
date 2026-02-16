package ziptax

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	internalhttp "github.com/ziptax/ziptax-go/internal/http"
	"github.com/ziptax/ziptax-go/internal/validation"
	"github.com/ziptax/ziptax-go/models"
)

// Client is the main ZipTax API client.
type Client struct {
	httpClient *internalhttp.Client
	config     *Config
}

// NewClient creates a new ZipTax API client with the given API key and options.
//
// Example:
//
//	client := ziptax.NewClient("your-api-key")
//
// Or with options:
//
//	client := ziptax.NewClient(
//		"your-api-key",
//		ziptax.WithTimeout(60*time.Second),
//		ziptax.WithMaxRetries(5),
//	)
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	config := &Config{
		APIKey: apiKey,
	}

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Apply defaults
	config.applyDefaults()

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create retry policy
	retryPolicy := internalhttp.DefaultRetryPolicy(
		config.MaxRetries,
		config.RetryWaitMin,
		config.RetryWaitMax,
	)

	// Create HTTP client wrapper
	httpClient := internalhttp.NewClient(
		config.HTTPClient,
		config.BaseURL,
		config.APIKey,
		config.UserAgent,
		retryPolicy,
		config.Logger,
	)

	return &Client{
		httpClient: httpClient,
		config:     config,
	}, nil
}

// GetSalesTaxByAddress returns sales and use tax rate details from an address input.
//
// Example:
//
//	ctx := context.Background()
//	response, err := client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
//	if err != nil {
//		return fmt.Errorf("failed to get tax rates: %w", err)
//	}
//	fmt.Printf("Total tax rate: %.4f\n", response.TaxSummaries[0].Rate)
func (c *Client) GetSalesTaxByAddress(ctx context.Context, address string, opts ...RequestOption) (*models.V60Response, error) {
	// Validate address
	if err := validation.ValidateAddress(address); err != nil {
		return nil, &ValidationError{
			Field:   "address",
			Value:   address,
			Message: err.Error(),
		}
	}

	// Build request options
	reqOpts := &RequestOptions{}
	for _, opt := range opts {
		opt(reqOpts)
	}

	// Validate optional parameters
	if err := validation.ValidateHistoricalDate(reqOpts.Historical); err != nil {
		return nil, &ValidationError{
			Field:   "historical",
			Value:   reqOpts.Historical,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateCountryCode(reqOpts.CountryCode); err != nil {
		return nil, &ValidationError{
			Field:   "countryCode",
			Value:   reqOpts.CountryCode,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateFormat(reqOpts.Format); err != nil {
		return nil, &ValidationError{
			Field:   "format",
			Value:   reqOpts.Format,
			Message: err.Error(),
		}
	}

	// Build query parameters
	queryParams := map[string]string{
		"address": address,
	}
	if reqOpts.Historical != "" {
		queryParams["historical"] = reqOpts.Historical
	}
	if reqOpts.CountryCode != "" {
		queryParams["countryCode"] = reqOpts.CountryCode
	}
	if reqOpts.Format != "" {
		queryParams["format"] = reqOpts.Format
	}

	// Make request
	var response models.V60Response
	if err := c.httpClient.Get(ctx, "/request/v60", queryParams, &response); err != nil {
		return nil, fmt.Errorf("failed to get sales tax by address: %w", err)
	}

	return &response, nil
}

// GetSalesTaxByGeoLocation returns sales and use tax rate details from a geolocation input.
//
// Example:
//
//	ctx := context.Background()
//	response, err := client.GetSalesTaxByGeoLocation(ctx, "33.65253", "-117.74794")
//	if err != nil {
//		return fmt.Errorf("failed to get tax rates: %w", err)
//	}
//	fmt.Printf("Total tax rate: %.4f\n", response.TaxSummaries[0].Rate)
func (c *Client) GetSalesTaxByGeoLocation(ctx context.Context, lat, lng string, opts ...RequestOption) (*models.V60Response, error) {
	// Validate coordinates
	if err := validation.ValidateCoordinates(lat, lng); err != nil {
		return nil, &ValidationError{
			Field:   "lat/lng",
			Value:   fmt.Sprintf("%s,%s", lat, lng),
			Message: err.Error(),
		}
	}

	// Build request options
	reqOpts := &RequestOptions{}
	for _, opt := range opts {
		opt(reqOpts)
	}

	// Validate optional parameters
	if err := validation.ValidateHistoricalDate(reqOpts.Historical); err != nil {
		return nil, &ValidationError{
			Field:   "historical",
			Value:   reqOpts.Historical,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateCountryCode(reqOpts.CountryCode); err != nil {
		return nil, &ValidationError{
			Field:   "countryCode",
			Value:   reqOpts.CountryCode,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateFormat(reqOpts.Format); err != nil {
		return nil, &ValidationError{
			Field:   "format",
			Value:   reqOpts.Format,
			Message: err.Error(),
		}
	}

	// Build query parameters
	queryParams := map[string]string{
		"lat": lat,
		"lng": lng,
	}
	if reqOpts.Historical != "" {
		queryParams["historical"] = reqOpts.Historical
	}
	if reqOpts.CountryCode != "" {
		queryParams["countryCode"] = reqOpts.CountryCode
	}
	if reqOpts.Format != "" {
		queryParams["format"] = reqOpts.Format
	}

	// Make request
	var response models.V60Response
	if err := c.httpClient.Get(ctx, "/request/v60", queryParams, &response); err != nil {
		return nil, fmt.Errorf("failed to get sales tax by geolocation: %w", err)
	}

	return &response, nil
}

// GetAccountMetrics returns account metrics related to sales and use tax.
//
// Example:
//
//	ctx := context.Background()
//	metrics, err := client.GetAccountMetrics(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get account metrics: %w", err)
//	}
//	fmt.Printf("Core usage: %.2f%%\n", metrics.CoreUsagePercent)
func (c *Client) GetAccountMetrics(ctx context.Context) (*models.V60AccountMetrics, error) {
	var metrics models.V60AccountMetrics
	if err := c.httpClient.Get(ctx, "/account/v60/metrics", nil, &metrics); err != nil {
		return nil, fmt.Errorf("failed to get account metrics: %w", err)
	}

	return &metrics, nil
}

// GetRatesByPostalCode returns sales and use tax rate details from a US postal code input.
// A single postal code may return multiple results for different cities that share the same postal code.
//
// Example:
//
//	ctx := context.Background()
//	response, err := client.GetRatesByPostalCode(ctx, "92694")
//	if err != nil {
//		return fmt.Errorf("failed to get tax rates: %w", err)
//	}
//	for _, result := range response.Results {
//		fmt.Printf("City: %s, Total tax rate: %.4f\n", result.GeoCity, result.TaxSales)
//	}
func (c *Client) GetRatesByPostalCode(ctx context.Context, postalCode string, opts ...RequestOption) (*models.V60PostalCodeResponse, error) {
	// Validate postal code
	if err := validation.ValidatePostalCode(postalCode); err != nil {
		return nil, &ValidationError{
			Field:   "postalcode",
			Value:   postalCode,
			Message: err.Error(),
		}
	}

	// Build request options
	reqOpts := &RequestOptions{}
	for _, opt := range opts {
		opt(reqOpts)
	}

	// Validate optional parameters
	if err := validation.ValidateFormat(reqOpts.Format); err != nil {
		return nil, &ValidationError{
			Field:   "format",
			Value:   reqOpts.Format,
			Message: err.Error(),
		}
	}

	// Build query parameters
	queryParams := map[string]string{
		"postalcode": postalCode,
	}
	if reqOpts.Format != "" {
		queryParams["format"] = reqOpts.Format
	}

	// Make request
	var response models.V60PostalCodeResponse
	if err := c.httpClient.Get(ctx, "/request/v60", queryParams, &response); err != nil {
		return nil, fmt.Errorf("failed to get rates by postal code: %w", err)
	}

	return &response, nil
}

// RequestOptions holds optional parameters for API requests.
type RequestOptions struct {
	Historical  string // Historical date for rates (YYYY-MM format)
	CountryCode string // Country code (USA or CAN)
	Format      string // Response format (json or xml)
}

// RequestOption is a functional option for API requests.
type RequestOption func(*RequestOptions)

// WithHistorical sets the historical date for tax rates (YYYY-MM format).
func WithHistorical(date string) RequestOption {
	return func(o *RequestOptions) {
		o.Historical = date
	}
}

// WithCountryCode sets the country code (USA or CAN).
func WithCountryCode(code string) RequestOption {
	return func(o *RequestOptions) {
		o.CountryCode = code
	}
}

// WithFormat sets the response format (json or xml).
func WithFormat(format string) RequestOption {
	return func(o *RequestOptions) {
		o.Format = format
	}
}

// CreateOrder creates an order in TaxCloud for order management and tax filing.
// This function requires TaxCloud credentials to be configured during client initialization.
//
// Example:
//
//	ctx := context.Background()
//	orderReq := &models.CreateOrderRequest{
//		OrderID:         "order-123",
//		CustomerID:      "customer-456",
//		TransactionDate: "2024-01-15T09:30:00Z",
//		CompletedDate:   "2024-01-15T09:30:00Z",
//		Origin: models.TaxCloudAddress{
//			Line1: "323 Washington Ave N",
//			City:  "Minneapolis",
//			State: "MN",
//			Zip:   "55401-2427",
//		},
//		Destination: models.TaxCloudAddress{
//			Line1: "323 Washington Ave N",
//			City:  "Minneapolis",
//			State: "MN",
//			Zip:   "55401-2427",
//		},
//		LineItems: []models.CartItemWithTax{
//			{
//				Index:    0,
//				ItemID:   "item-1",
//				Price:    10.8,
//				Quantity: 1.5,
//				Tax: models.Tax{
//					Amount: 1.31,
//					Rate:   0.0813,
//				},
//			},
//		},
//		Currency: &models.Currency{},
//	}
//	response, err := client.CreateOrder(ctx, orderReq)
//	if err != nil {
//		return fmt.Errorf("failed to create order: %w", err)
//	}
//	fmt.Printf("Order created with ID: %s\n", response.OrderID)
func (c *Client) CreateOrder(ctx context.Context, request *models.CreateOrderRequest) (*models.OrderResponse, error) {
	// Check if TaxCloud credentials are configured
	if !c.config.HasTaxCloudCredentials() {
		return nil, ErrTaxCloudNotConfigured
	}

	// Build the path with connection ID
	path := fmt.Sprintf("/tax/connections/%s/orders", c.config.TaxCloudConnectionID)

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the POST request to TaxCloud API
	var response models.OrderResponse
	if err := c.httpClient.Post(ctx, c.config.TaxCloudBaseURL, path, headers, request, &response); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &response, nil
}

// GetOrder retrieves a specific order by its ID from TaxCloud.
// This function requires TaxCloud credentials to be configured during client initialization.
//
// Example:
//
//	ctx := context.Background()
//	order, err := client.GetOrder(ctx, "order-123")
//	if err != nil {
//		return fmt.Errorf("failed to get order: %w", err)
//	}
//	fmt.Printf("Order %s has %d line items\n", order.OrderID, len(order.LineItems))
func (c *Client) GetOrder(ctx context.Context, orderID string) (*models.OrderResponse, error) {
	// Check if TaxCloud credentials are configured
	if !c.config.HasTaxCloudCredentials() {
		return nil, ErrTaxCloudNotConfigured
	}

	// Build the path with connection ID and order ID
	path := fmt.Sprintf("/tax/connections/%s/orders/%s", c.config.TaxCloudConnectionID, orderID)

	// Build URL
	u := c.config.TaxCloudBaseURL + path

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", c.config.TaxCloudAPIKey)
	req.Header.Set("User-Agent", "ziptax-go/1.0.0")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response models.OrderResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &response, nil
}

// UpdateOrder updates an existing order's completedDate in TaxCloud.
// Use this endpoint to change when an order was shipped/completed.
// This function requires TaxCloud credentials to be configured during client initialization.
//
// Example:
//
//	ctx := context.Background()
//	updateReq := &models.UpdateOrderRequest{
//		CompletedDate: "2024-01-16T10:00:00Z",
//	}
//	order, err := client.UpdateOrder(ctx, "order-123", updateReq)
//	if err != nil {
//		return fmt.Errorf("failed to update order: %w", err)
//	}
//	fmt.Printf("Order updated with new completed date: %s\n", order.CompletedDate)
func (c *Client) UpdateOrder(ctx context.Context, orderID string, request *models.UpdateOrderRequest) (*models.OrderResponse, error) {
	// Check if TaxCloud credentials are configured
	if !c.config.HasTaxCloudCredentials() {
		return nil, ErrTaxCloudNotConfigured
	}

	// Build the path with connection ID and order ID
	path := fmt.Sprintf("/tax/connections/%s/orders/%s", c.config.TaxCloudConnectionID, orderID)

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the PATCH request to TaxCloud API
	var response models.OrderResponse
	if err := c.httpClient.Patch(ctx, c.config.TaxCloudBaseURL, path, headers, request, &response); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return &response, nil
}

// RefundOrder creates a refund against an order in TaxCloud.
// An order can only be refunded once, regardless of whether the order is partially or fully refunded.
// This function requires TaxCloud credentials to be configured during client initialization.
//
// Example (partial refund):
//
//	ctx := context.Background()
//	refundReq := &models.RefundTransactionRequest{
//		Items: []models.CartItemRefundWithTaxRequest{
//			{
//				ItemID:   "item-1",
//				Quantity: 1.0,
//			},
//		},
//	}
//	refunds, err := client.RefundOrder(ctx, "order-123", refundReq)
//	if err != nil {
//		return fmt.Errorf("failed to refund order: %w", err)
//	}
//	fmt.Printf("Refunded %d items\n", len(refunds[0].Items))
//
// Example (full refund):
//
//	ctx := context.Background()
//	refundReq := &models.RefundTransactionRequest{} // Empty request for full refund
//	refunds, err := client.RefundOrder(ctx, "order-123", refundReq)
func (c *Client) RefundOrder(ctx context.Context, orderID string, request *models.RefundTransactionRequest) ([]models.RefundTransactionResponse, error) {
	// Check if TaxCloud credentials are configured
	if !c.config.HasTaxCloudCredentials() {
		return nil, ErrTaxCloudNotConfigured
	}

	// Build the path with connection ID and order ID
	path := fmt.Sprintf("/tax/connections/%s/orders/refunds/%s", c.config.TaxCloudConnectionID, orderID)

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the POST request to TaxCloud API
	var response []models.RefundTransactionResponse
	if err := c.httpClient.Post(ctx, c.config.TaxCloudBaseURL, path, headers, request, &response); err != nil {
		return nil, fmt.Errorf("failed to refund order: %w", err)
	}

	return response, nil
}
