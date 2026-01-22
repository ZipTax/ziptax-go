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

Create a client with default settings:

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
    ziptax.WithHistorical("2024-01"),
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

### Get Account Metrics

```go
metrics, err := client.GetAccountMetrics(ctx)
fmt.Printf("Core Usage: %.2f%%\n", metrics.CoreUsagePercent)
fmt.Printf("Geo Usage: %.2f%%\n", metrics.GeoUsagePercent)
```

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

- `WithBaseURL(url string)` - Set a custom API base URL
- `WithHTTPClient(client *http.Client)` - Use a custom HTTP client
- `WithTimeout(timeout time.Duration)` - Set request timeout
- `WithMaxRetries(max int)` - Set maximum retry attempts (0 to disable)
- `WithRetryWait(min, max time.Duration)` - Set retry backoff times
- `WithLogger(logger Logger)` - Enable request/response logging
- `WithUserAgent(ua string)` - Set a custom User-Agent header

### Request Options

- `WithHistorical(date string)` - Get historical rates (YYYY-MM format)
- `WithCountryCode(code string)` - Specify country code (USA or CAN)
- `WithFormat(format string)` - Set response format (json or xml)

## Examples

See the [examples](./examples) directory for complete examples:

- [Basic Usage](./examples/basic_usage) - Simple API calls
- [Concurrent Usage](./examples/concurrent_usage) - Parallel requests with goroutines
- [Error Handling](./examples/error_handling) - Proper error handling patterns
- [Context Timeout](./examples/context_timeout) - Using context for timeouts and cancellation

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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For API-related questions and support, please contact [support@zip.tax](mailto:support@zip.tax).

For SDK issues, please [open an issue](https://github.com/ziptax/ziptax-go/issues) on GitHub.

## Links

- [ZipTax Website](https://zip.tax/)
- [API Documentation](https://api.zip-tax.com/)
- [Go Package Documentation](https://pkg.go.dev/github.com/ziptax/ziptax-go)