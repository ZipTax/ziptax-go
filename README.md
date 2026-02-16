# ZipTax Go SDK

Official Go SDK for the [ZipTax API](https://zip.tax/) - get accurate sales and use tax rates for any US address or geolocation.

[![Go Reference](https://pkg.go.dev/badge/github.com/ziptax/ziptax-go.svg)](https://pkg.go.dev/github.com/ziptax/ziptax-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ziptax/ziptax-go)](https://goreportcard.com/report/github.com/ziptax/ziptax-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- 🚀 Simple, idiomatic Go API
- 🔄 Automatic retry with exponential backoff
- ⏱️ Context support for timeouts and cancellation
- 🔍 Input validation
- 🧪 Comprehensive test coverage (>88%)
- 📝 Full type safety with Go structs
- 🔐 Secure API key authentication
- 🌐 Support for US and Canadian addresses
- 📍 Geolocation-based lookups
- 📦 **TaxCloud Order Management** - Create, retrieve, update, and refund orders

## Installation

```bash
go get github.com/ziptax/ziptax-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ziptax/ziptax-go"
)

func main() {
    // Create a new client with your API key
    client, err := ziptax.NewClient("your-api-key-here")
    if err != nil {
        log.Fatal(err)
    }

    // Get sales tax by address
    ctx := context.Background()
    response, err := client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
    if err != nil {
        log.Fatal(err)
    }

    // Print the results
    fmt.Printf("Address: %s\n", response.AddressDetail.NormalizedAddress)
    if len(response.TaxSummaries) > 0 {
        fmt.Printf("Total Tax Rate: %.4f%%\n", response.TaxSummaries[0].Rate*100)
    }
}
```

## Usage

### Client Initialization

Create a client with default settings (ZipTax API only):

```go
client, err := ziptax.NewClient("your-api-key")
```

Or customize the client with options:

```go
client, err := ziptax.NewClient(
    "your-api-key",
    ziptax.WithTimeout(60*time.Second),
    ziptax.WithMaxRetries(5),
    ziptax.WithRetryWait(2*time.Second, 60*time.Second),
)
```

**Enable TaxCloud Order Management (Optional):**

To use TaxCloud order features, provide both TaxCloud credentials during initialization:

```go
client, err := ziptax.NewClient(
    "your-ziptax-api-key",
    ziptax.WithTaxCloudConnectionID("your-taxcloud-connection-id"),
    ziptax.WithTaxCloudAPIKey("your-taxcloud-api-key"),
)
```

### Get Sales Tax by Address

```go
response, err := client.GetSalesTaxByAddress(
    ctx,
    "200 Spectrum Center Drive, Irvine, CA 92618",
)
```

With optional parameters:

```go
response, err := client.GetSalesTaxByAddress(
    ctx,
    "200 Spectrum Center Drive, Irvine, CA 92618",
    ziptax.WithHistorical("202401"),
    ziptax.WithCountryCode("USA"),
    ziptax.WithFormat("json"),
)
```

### Get Sales Tax by Geolocation

```go
response, err := client.GetSalesTaxByGeoLocation(
    ctx,
    "33.65253",  // latitude
    "-117.74794", // longitude
)
```

### Get Rates by Postal Code

```go
response, err := client.GetRatesByPostalCode(ctx, "92694")
for _, result := range response.Results {
    fmt.Printf("City: %s, Total Tax Rate: %.4f%%\n", result.GeoCity, result.TaxSales*100)
}
```

### Get Account Metrics

```go
metrics, err := client.GetAccountMetrics(ctx)
fmt.Printf("Core Usage: %.2f%%\n", metrics.CoreUsagePercent)
fmt.Printf("Geo Usage: %.2f%%\n", metrics.GeoUsagePercent)
```

## TaxCloud Order Management

The SDK supports comprehensive order lifecycle management through the TaxCloud API. To use these features, you must configure TaxCloud credentials during client initialization.

### Create Order

Create a new order for tax filing:

```go
orderReq := &models.CreateOrderRequest{
    OrderID:         "order-123",
    CustomerID:      "customer-456",
    TransactionDate: "2024-01-15T09:30:00Z",
    CompletedDate:   "2024-01-15T09:30:00Z",
    Origin: models.TaxCloudAddress{
        Line1: "200 Spectrum Center Drive Suite 300",
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
            Tax: models.Tax{
                Amount: 1.31,
                Rate:   0.0813,
            },
        },
    },
    Currency: &models.Currency{},
}

response, err := client.CreateOrder(ctx, orderReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order created: %s\n", response.OrderID)
```

### Get Order

Retrieve an existing order by ID:

```go
order, err := client.GetOrder(ctx, "order-123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order %s has %d line items\n", order.OrderID, len(order.LineItems))
```

### Update Order

Update an order's completed date (when it was shipped/completed):

```go
updateReq := &models.UpdateOrderRequest{
    CompletedDate: "2024-01-16T10:00:00Z",
}

order, err := client.UpdateOrder(ctx, "order-123", updateReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order updated with new completed date: %s\n", order.CompletedDate)
```

### Refund Order

Create a full or partial refund against an order:

**Partial Refund:**

```go
refundReq := &models.RefundTransactionRequest{
    Items: []models.CartItemRefundWithTaxRequest{
        {
            ItemID:   "item-1",
            Quantity: 1.0,
        },
    },
}

refunds, err := client.RefundOrder(ctx, "order-123", refundReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Refunded %d items\n", len(refunds[0].Items))
```

**Full Refund:**

```go
// Empty request for full order refund
refundReq := &models.RefundTransactionRequest{}

refunds, err := client.RefundOrder(ctx, "order-123", refundReq)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Full order refund created\n")
```

**Note:** An order can only be refunded once, regardless of whether it's a partial or full refund.

## Error Handling

The SDK provides structured error types for proper error handling:

```go
response, err := client.GetSalesTaxByAddress(ctx, address)
if err != nil {
    // Check for validation errors
    var validationErr *ziptax.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Validation error: %s\n", validationErr.Message)
        return
    }

    // Check for API errors
    var apiErr *ziptax.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error: %s (code %d)\n", apiErr.Message, apiErr.Code)
        return
    }

    // Check for sentinel errors
    if errors.Is(err, ziptax.ErrInvalidAPIKey) {
        fmt.Println("Invalid API key")
        return
    }

    // Check for TaxCloud errors
    if errors.Is(err, ziptax.ErrTaxCloudNotConfigured) {
        fmt.Println("TaxCloud credentials not configured")
        return
    }

    // Generic error
    fmt.Printf("Error: %v\n", err)
    return
}
```

## Context Support

All API methods accept a `context.Context` for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := client.GetSalesTaxByAddress(ctx, address)
```

```go
// With cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Cancel from another goroutine
go func() {
    time.Sleep(100 * time.Millisecond)
    cancel()
}()

response, err := client.GetSalesTaxByAddress(ctx, address)
```

## Concurrent Requests

The client is safe for concurrent use:

```go
var wg sync.WaitGroup
addresses := []string{"address1", "address2", "address3"}

for _, addr := range addresses {
    wg.Add(1)
    go func(address string) {
        defer wg.Done()
        resp, err := client.GetSalesTaxByAddress(ctx, address)
        // Handle response...
    }(addr)
}

wg.Wait()
```

## Configuration Options

### Client Options

**Core Options:**
- `WithBaseURL(url string)` - Set a custom API base URL
- `WithHTTPClient(client *http.Client)` - Use a custom HTTP client
- `WithTimeout(timeout time.Duration)` - Set request timeout
- `WithMaxRetries(max int)` - Set maximum retry attempts (0 to disable)
- `WithRetryWait(min, max time.Duration)` - Set retry backoff times
- `WithLogger(logger Logger)` - Enable request/response logging
- `WithUserAgent(ua string)` - Set a custom User-Agent header

**TaxCloud Options:**
- `WithTaxCloudConnectionID(id string)` - Set TaxCloud Connection ID (required for order management)
- `WithTaxCloudAPIKey(key string)` - Set TaxCloud API Key (required for order management)
- `WithTaxCloudBaseURL(url string)` - Set custom TaxCloud API base URL

### Request Options

- `WithHistorical(date string)` - Get historical rates (YYYYMM format, e.g., "202401")
- `WithCountryCode(code string)` - Specify country code (USA or CAN)
- `WithFormat(format string)` - Set response format (json or xml)

## Examples

See the [examples](./examples) directory for complete examples:

- [Basic Usage](./examples/basic_usage) - Simple API calls
- [Concurrent Usage](./examples/concurrent_usage) - Parallel requests with goroutines
- [Error Handling](./examples/error_handling) - Proper error handling patterns
- [Context Timeout](./examples/context_timeout) - Using context for timeouts and cancellation
- [TaxCloud Order](./examples/taxcloud_order) - TaxCloud order management operations

## API Reference

For detailed API documentation, see [pkg.go.dev](https://pkg.go.dev/github.com/ziptax/ziptax-go).

## Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile commands)

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with verbose output
make test-verbose
```

### Linting and Formatting

```bash
# Format code
make fmt

# Organize imports
make imports

# Run linter
make lint

# Run all checks
make check
```

### Building

```bash
# Build the SDK
make build
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md) for a detailed list of changes in each release.

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The version is defined in [`version.go`](./version.go).

### Version Format: `MAJOR.MINOR.PATCH`

- **MAJOR**: Breaking changes (incompatible API changes)
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### For Contributors

When submitting a PR, you must bump the version in `version.go`. See [`.github/VERSION_BUMP_GUIDE.md`](./.github/VERSION_BUMP_GUIDE.md) for detailed instructions.

The GitHub Actions workflow will automatically verify that the semantic version has been properly bumped.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

**Before submitting:**
1. Update the version in `version.go` following semantic versioning
2. Update [`CHANGELOG.md`](./CHANGELOG.md) with your changes
3. Add tests for new functionality
4. Update documentation (README, GoDoc comments)
5. Ensure all tests pass (`make test`)
6. Run linter (`make lint`)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For API-related questions and support, please contact [support@zip.tax](mailto:support@zip.tax).

For SDK issues, please [open an issue](https://github.com/ziptax/ziptax-go/issues) on GitHub.

## Links

- [ZipTax Website](https://zip.tax/)
- [API Documentation](https://api.zip-tax.com/)
- [Go Package Documentation](https://pkg.go.dev/github.com/ziptax/ziptax-go)