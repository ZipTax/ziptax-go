package ziptax

import (
	"context"
	"fmt"

	"github.com/ziptax/ziptax-go/internal/http"
	"github.com/ziptax/ziptax-go/internal/validation"
	"github.com/ziptax/ziptax-go/models"
)

// Client is the main ZipTax API client.
type Client struct {
	httpClient *http.Client
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
	retryPolicy := http.DefaultRetryPolicy(
		config.MaxRetries,
		config.RetryWaitMin,
		config.RetryWaitMax,
	)

	// Create HTTP client wrapper
	httpClient := http.NewClient(
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
