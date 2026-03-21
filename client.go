package ziptax

import (
	"context"
	"errors"
	"fmt"

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
		return nil, wrapError("failed to get sales tax by address", err)
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
		return nil, wrapError("failed to get sales tax by geolocation", err)
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
		return nil, wrapError("failed to get account metrics", err)
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

	// Strip 9-digit suffix if present (API only supports 5-digit postal codes)
	postalCode = validation.NormalizePostalCode(postalCode)

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
		return nil, wrapError("failed to get rates by postal code", err)
	}

	return &response, nil
}

// RequestOptions holds optional parameters for API requests.
type RequestOptions struct {
	Historical  string // Historical date for rates (YYYYMM format, e.g., "202401")
	CountryCode string // Country code (USA or CAN)
	Format      string // Response format (json or xml)
}

// RequestOption is a functional option for API requests.
type RequestOption func(*RequestOptions)

// WithHistorical sets the historical date for tax rates (YYYYMM format, e.g., "202401").
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

// SearchProductCodes searches for product codes (TICs) by natural language description.
// Returns all matching Taxability Information Codes ranked and scored by relevance.
//
// Use the returned TicID as the taxabilityCode parameter in rate requests or
// cart line items. For v60 rate requests (e.g., GetSalesTaxByAddress), pass TicID
// as a string. For cart line items, convert to int64.
//
// Example:
//
//	ctx := context.Background()
//	response, err := client.SearchProductCodes(ctx, "baked goods sold in plastic packaging")
//	if err != nil {
//		return fmt.Errorf("failed to search product codes: %w", err)
//	}
//	for _, result := range response.Results {
//		fmt.Printf("TIC %s: %s (rank %s, score %s)\n",
//			result.TicID, result.Label, result.Rank, result.Score)
//	}
func (c *Client) SearchProductCodes(ctx context.Context, query string) (*models.ProductCodeSearchResponse, error) {
	// Validate query
	if err := validation.ValidateProductQuery(query); err != nil {
		return nil, &ValidationError{
			Field:   "query",
			Value:   query,
			Message: err.Error(),
		}
	}

	// Build request body
	reqBody := &models.ProductCodeSearchRequest{
		Query: query,
	}

	// Make the POST request to ZipTax API.
	// Note: Post() requires explicit baseURL and headers because it was designed
	// for multi-API routing (ZipTax vs TaxCloud). Unlike Get(), which auto-sets
	// X-API-Key from the HTTP client's stored key, Post() relies on the caller
	// to provide headers. This is consistent with calculateCartZipTax and all
	// other Post() call sites.
	var response models.ProductCodeSearchResponse
	if err := c.httpClient.Post(ctx, c.config.BaseURL, "/search/tic", map[string]string{
		"X-API-Key": c.config.APIKey,
	}, reqBody, &response); err != nil {
		return nil, wrapError("failed to search product codes", err)
	}

	return &response, nil
}

// RecommendProductCode gets an AI-powered product code (TIC) recommendation.
// Returns a single best-match TIC code with higher accuracy than SearchProductCodes.
// Has slightly higher latency due to the AI processing step.
//
// Use the returned TicID as the taxabilityCode parameter in rate requests or
// cart line items. For v60 rate requests (e.g., GetSalesTaxByAddress), pass TicID
// as a string. For cart line items, convert to int64.
//
// Example:
//
//	ctx := context.Background()
//	response, err := client.RecommendProductCode(ctx, "baked goods sold in plastic packaging")
//	if err != nil {
//		return fmt.Errorf("failed to recommend product code: %w", err)
//	}
//	prediction := response.Predictions[0]
//	if prediction.Status == "success" {
//		fmt.Printf("Recommended TIC: %s (%s)\n", prediction.TicID, prediction.Label)
//	}
func (c *Client) RecommendProductCode(ctx context.Context, query string) (*models.ProductCodeRecommendationResponse, error) {
	// Validate query
	if err := validation.ValidateProductQuery(query); err != nil {
		return nil, &ValidationError{
			Field:   "query",
			Value:   query,
			Message: err.Error(),
		}
	}

	// Build request body
	reqBody := &models.ProductCodeSearchRequest{
		Query: query,
	}

	// Make the POST request to ZipTax API (see SearchProductCodes for Post() pattern notes).
	var response models.ProductCodeRecommendationResponse
	if err := c.httpClient.Post(ctx, c.config.BaseURL, "/search/tic/recommend", map[string]string{
		"X-API-Key": c.config.APIKey,
	}, reqBody, &response); err != nil {
		return nil, wrapError("failed to recommend product code", err)
	}

	return &response, nil
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

	// Validate request
	if request == nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: "request cannot be nil",
		}
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
		return nil, wrapError("failed to create order", err)
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

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the GET request to TaxCloud API
	var response models.OrderResponse
	if err := c.httpClient.GetWithOptions(ctx, c.config.TaxCloudBaseURL, path, headers, &response); err != nil {
		return nil, wrapError("failed to get order", err)
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

	// Validate request
	if request == nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: "request cannot be nil",
		}
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
		return nil, wrapError("failed to update order", err)
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

	// Validate request
	if request == nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: "request cannot be nil",
		}
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
		return nil, wrapError("failed to refund order", err)
	}

	return response, nil
}

// CreateOrderFromCart creates an order from a previously calculated cart in TaxCloud.
// The user must have previously called CalculateCart with TaxCloud credentials and stored
// the returned cartId from the TaxCloudCartItemResponse.
//
// TaxCloud automatically commits the order at the time of creation, finalizing it for tax
// filing. To set a completed date on the order, use UpdateOrder after creation.
//
// This function requires TaxCloud credentials to be configured during client initialization.
//
// Example:
//
//	ctx := context.Background()
//	req := &models.CreateOrderFromCartRequest{
//		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
//		OrderID: "my-order-1",
//	}
//	response, err := client.CreateOrderFromCart(ctx, req)
//	if err != nil {
//		return fmt.Errorf("failed to create order from cart: %w", err)
//	}
//	fmt.Printf("Order created with ID: %s\n", response.OrderID)
func (c *Client) CreateOrderFromCart(ctx context.Context, request *models.CreateOrderFromCartRequest) (*models.OrderResponse, error) {
	// Check if TaxCloud credentials are configured
	if !c.config.HasTaxCloudCredentials() {
		return nil, ErrTaxCloudNotConfigured
	}

	// Validate request
	if request == nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: "request cannot be nil",
		}
	}

	if err := validation.ValidateCreateOrderFromCartRequest(&validation.CreateOrderFromCartValidationInput{
		CartID:  request.CartID,
		OrderID: request.OrderID,
	}); err != nil {
		return nil, &ValidationError{
			Field:   "CreateOrderFromCartRequest",
			Value:   fmt.Sprintf("cartId=%q, orderId=%q", request.CartID, request.OrderID),
			Message: err.Error(),
		}
	}

	// Build the path with connection ID
	path := fmt.Sprintf("/tax/connections/%s/carts/orders", c.config.TaxCloudConnectionID)

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the POST request to TaxCloud API
	var response models.OrderResponse
	if err := c.httpClient.Post(ctx, c.config.TaxCloudBaseURL, path, headers, request, &response); err != nil {
		return nil, wrapError("failed to create order from cart", err)
	}

	return &response, nil
}

// wrapError converts internal HTTP APIError to public APIError for proper errors.As support,
// then wraps it with a descriptive message.
func wrapError(msg string, err error) error {
	var internalErr *internalhttp.APIError
	if errors.As(err, &internalErr) {
		return fmt.Errorf("%s: %w", msg, &APIError{
			StatusCode: internalErr.StatusCode,
			Code:       internalErr.Code,
			Name:       internalErr.Name,
			Message:    internalErr.Message,
		})
	}
	return fmt.Errorf("%s: %w", msg, err)
}
