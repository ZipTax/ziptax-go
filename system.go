package ziptax

import (
	"context"

	"github.com/ziptax/ziptax-go/models"
)

// Reference data and service health endpoints.
//
// GetTICCodes, GetTICSearchSchema, GetHealth, and GetSystemMetadata are public
// endpoints that do not require an API key. The SDK still sends the configured
// key with every request, which the API ignores on these paths.

// GetTICCodes returns the full catalog of Taxability Information Codes.
//
// A TIC classifies a product or service for product-specific tax rules. Use the
// returned ID as the taxabilityCode parameter on rate requests, or as the TIC on
// merchant cart and order line items.
//
// To find a TIC from a product description instead of paging the catalog, use
// SearchProductCodes or RecommendProductCode.
//
// The API can also serve this catalog as XML; the SDK requests JSON only.
//
// Example:
//
//	ctx := context.Background()
//	catalog, err := client.GetTICCodes(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get TIC codes: %w", err)
//	}
//	for _, entry := range catalog.TICList {
//		fmt.Printf("%s: %s\n", entry.TIC.ID, entry.TIC.Title)
//	}
func (c *Client) GetTICCodes(ctx context.Context) (*models.TICDataResponse, error) {
	var response models.TICDataResponse
	if err := c.httpClient.Get(ctx, "/data/tic", nil, &response); err != nil {
		return nil, wrapError("failed to get TIC codes", err)
	}

	return &response, nil
}

// GetTICSearchSchema returns the JSON Schema describing the product code search
// response, as served at /schemas/ticsearch.
//
// The schema is returned as a decoded map rather than a typed value: it is a JSON
// Schema document, and its contents are meant to be consumed by validators and
// code generators rather than read field by field.
//
// Example:
//
//	ctx := context.Background()
//	schema, err := client.GetTICSearchSchema(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get TIC search schema: %w", err)
//	}
//	fmt.Println(schema["$id"])
func (c *Client) GetTICSearchSchema(ctx context.Context) (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := c.httpClient.Get(ctx, "/schemas/ticsearch", nil, &response); err != nil {
		return nil, wrapError("failed to get TIC search schema", err)
	}

	return response, nil
}

// GetHealth returns the health of the ZipTax API and its components.
//
// Use HealthResponse.IsHealthy for a single overall verdict; the Components field
// carries the per-component detail behind it.
//
// Example:
//
//	ctx := context.Background()
//	health, err := client.GetHealth(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get health: %w", err)
//	}
//	fmt.Printf("healthy: %t (%d tax data records)\n",
//		health.IsHealthy(), health.Components.TaxDataCount)
func (c *Client) GetHealth(ctx context.Context) (*models.HealthResponse, error) {
	var response models.HealthResponse
	if err := c.httpClient.Get(ctx, "/system/health", nil, &response); err != nil {
		return nil, wrapError("failed to get health", err)
	}

	return &response, nil
}

// GetSystemMetadata returns metadata about the instance serving the request.
//
// Example:
//
//	ctx := context.Background()
//	meta, err := client.GetSystemMetadata(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get system metadata: %w", err)
//	}
//	fmt.Printf("%s running %s\n", meta.Hostname, meta.GoVersion)
func (c *Client) GetSystemMetadata(ctx context.Context) (*models.SystemMetadata, error) {
	var response models.SystemMetadata
	if err := c.httpClient.Get(ctx, "/system/metadata", nil, &response); err != nil {
		return nil, wrapError("failed to get system metadata", err)
	}

	return &response, nil
}

// GetDetailedAccountMetrics returns account usage broken down by entitlement:
// core tax lookups, geocoding, and merchant requests.
//
// GetAccountMetrics reports the simpler combined counter from the v6.0 endpoint.
// Use this function when you need the per-entitlement quotas, including the
// merchant counters that Merchant Management consumes.
//
// Example:
//
//	ctx := context.Background()
//	metrics, err := client.GetDetailedAccountMetrics(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to get account metrics: %w", err)
//	}
//	fmt.Printf("Core usage: %.2f%%, merchant usage: %.2f%%\n",
//		metrics.CoreUsagePercent, metrics.MerchantUsagePercent)
func (c *Client) GetDetailedAccountMetrics(ctx context.Context) (*models.AccountMetrics, error) {
	var response models.AccountMetrics
	if err := c.httpClient.Get(ctx, "/account/metrics", nil, &response); err != nil {
		return nil, wrapError("failed to get detailed account metrics", err)
	}

	return &response, nil
}
