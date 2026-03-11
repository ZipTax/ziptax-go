# CLAUDE.md - AI-Assisted Development Documentation

This document provides insights into how this SDK was developed with AI assistance, architectural decisions, and guidelines for future AI-assisted contributions.

## Project Overview

**ZipTax Go SDK** is an official Go client library for the ZipTax and TaxCloud APIs, providing:
- Sales and use tax rate lookups for US addresses and geolocations
- Complete TaxCloud order lifecycle management (create, retrieve, update, refund)
- Comprehensive error handling and retry logic
- Full type safety with Go structs

## AI-Assisted Development Approach

### Development Philosophy

This SDK was developed using **AI-pair programming** with Claude (Anthropic's AI assistant), following these principles:

1. **Specification-Driven Development**: All features were first documented in `docs/spec.yaml` before implementation
2. **Test-First Mindset**: Tests were written alongside or immediately after implementation
3. **Idiomatic Go**: All code follows Go best practices and conventions
4. **Documentation-First**: Every function includes comprehensive GoDoc comments with examples

### Development Workflow

```
1. API Research → 2. Spec Definition → 3. Model Creation → 4. Implementation → 5. Testing → 6. Documentation
```

#### Example: TaxCloud Order Management Feature

**Phase 1: API Documentation Gathering**
- Fetched official TaxCloud API documentation from https://docs.taxcloud.com
- Analyzed endpoint specifications, request/response formats, authentication methods

**Phase 2: Specification**
- Documented all endpoints in `docs/spec.yaml`:
  - CreateOrder (POST)
  - GetOrder (GET)
  - UpdateOrder (PATCH)
  - RefundOrder (POST)
- Defined data models with complete type information
- Included example requests/responses

**Phase 3: Implementation**
- Created data models in `models/taxcloud_orders.go`
- Added HTTP client methods (Post, Patch)
- Implemented client functions with error handling
- Added configuration options for TaxCloud credentials

**Phase 4: Testing & Validation**
- Unit tests for credential validation
- Integration test structure
- Build verification

**Phase 5: Documentation**
- Updated README.md with usage examples
- Added GoDoc comments to all public APIs
- Created example code in `examples/taxcloud_order/`

## Architecture Decisions

### Dual API Design

**Challenge**: Support two separate APIs (ZipTax and TaxCloud) in a single SDK.

**Solution**:
```go
// Optional TaxCloud credentials
client := ziptax.NewClient(
    "ziptax-api-key",
    ziptax.WithTaxCloudConnectionID("connection-id"),
    ziptax.WithTaxCloudAPIKey("taxcloud-key"),
)
```

**Rationale**:
- TaxCloud features are optional - only available when credentials provided
- Single client instance manages both APIs
- Clear error messages (`ErrTaxCloudNotConfigured`) when features unavailable
- No breaking changes to existing ZipTax-only users

### Configuration Pattern

**Functional Options Pattern**:
```go
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}
```

**Benefits**:
- Backward compatible - new options don't break existing code
- Self-documenting - option names clearly indicate purpose
- Flexible - options can be composed and reused
- Type-safe - compiler catches incorrect usage

### Error Handling Strategy

**Three-Tier Error System**:

1. **Sentinel Errors**: Pre-defined errors for common cases
   ```go
   var ErrInvalidAPIKey = errors.New("invalid API key")
   var ErrTaxCloudNotConfigured = errors.New("TaxCloud credentials not configured")
   ```

2. **Structured Errors**: Rich error types with context
   ```go
   type ValidationError struct {
       Field   string
       Value   string
       Message string
   }
   ```

3. **API Errors**: HTTP error responses with status codes
   ```go
   type APIError struct {
       StatusCode int
       Code       int
       Name       string
       Message    string
   }
   ```

**Rationale**: Enables precise error handling with `errors.Is()` and `errors.As()`

### HTTP Client Design

**Internal Package Pattern**:
```
internal/http/
  ├── client.go    # HTTP operations
  ├── retry.go     # Retry logic
  └── errors.go    # HTTP errors
```

**Key Features**:
- Automatic retry with exponential backoff
- Context support for timeouts/cancellation
- Request/response logging capability
- Method-specific functions (Get, Post, Patch)

**Rationale**:
- Separation of concerns - HTTP logic isolated from business logic
- Reusable across all API endpoints
- Easy to mock for testing
- `internal/` package prevents external import

## Code Organization

### Project Structure

```
ziptax-go/
├── client.go              # Main client and ZipTax API methods
├── config.go              # Configuration and validation
├── options.go             # Functional options
├── errors.go              # Error types and sentinels
├── models/
│   ├── v60.go            # ZipTax API models
│   └── taxcloud_orders.go # TaxCloud API models
├── internal/
│   ├── http/             # HTTP client implementation
│   └── validation/       # Input validation
├── examples/             # Usage examples
├── docs/
│   └── spec.yaml        # API specification
└── README.md            # User documentation
```

### Naming Conventions

**Consistent Patterns**:
- Models use exact API field names with JSON tags
- Request types: `*Request` suffix (e.g., `CreateOrderRequest`)
- Response types: `*Response` suffix (e.g., `OrderResponse`)
- Options: `With*` prefix (e.g., `WithTimeout`)
- Validation functions: `Validate*` prefix

## Testing Strategy

### Test Organization

Tests live alongside source files:
- `client_test.go` - Client functionality tests
- `config_test.go` - Configuration tests
- `options_test.go` - Options tests
- `errors_test.go` - Error handling tests

**Rationale**: Go convention - keeps tests close to implementation

### Test Coverage

Current coverage: **>88%**

**Focus Areas**:
1. Configuration validation
2. Error handling paths
3. Input validation
4. Credential checks
5. HTTP client behavior

### Testing Best Practices

```go
// Table-driven tests for multiple scenarios
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "12345", false},
        {"invalid", "abc", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

## Documentation Standards

### GoDoc Guidelines

Every exported function includes:
1. **Purpose**: One-line summary
2. **Details**: Additional context if needed
3. **Example**: Practical usage code
4. **Error Cases**: What errors might be returned

**Example**:
```go
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
func (c *Client) GetOrder(ctx context.Context, orderID string) (*models.OrderResponse, error)
```

### Specification Format

`docs/spec.yaml` contains:
- API endpoint specifications
- Request/response schemas
- Authentication details
- Example payloads
- Implementation notes

**Purpose**: Single source of truth for API contracts

## AI Assistance Best Practices

### Effective Prompting

**Successful Pattern**:
1. Provide context: "We're building a Go SDK for tax APIs"
2. Share specifications: Link to API documentation
3. Request structure: "First update spec.yaml, then implement"
4. Verify iteratively: "Build and test after each change"

**Example Interaction**:
```
Human: "Add TaxCloud UpdateOrder endpoint from [documentation URL]"

Claude:
1. Fetches and analyzes API docs
2. Updates spec.yaml with endpoint definition
3. Adds data models
4. Implements client method
5. Adds tests
6. Updates documentation
```

### Code Review Checklist

When reviewing AI-generated code:
- ✅ Follows Go conventions (gofmt, golint)
- ✅ Includes error handling
- ✅ Has comprehensive documentation
- ✅ Includes usage examples
- ✅ Adds tests
- ✅ Updates relevant documentation
- ✅ Maintains backward compatibility

## Future Development

### Adding New Features

**Process**:
1. Research API endpoint/feature
2. Update `docs/spec.yaml` with specification
3. Add/update models in `models/`
4. Implement in `client.go` or create new file
5. Add tests
6. Update README.md
7. Add example if appropriate

### Maintaining Consistency

**Guidelines**:
- All TaxCloud features check credentials with `HasTaxCloudCredentials()`
- All errors use structured types or sentinels
- All public APIs include GoDoc with examples
- All inputs are validated
- All operations support context for cancellation

### Breaking Changes

**Avoid**:
- Changing public API signatures
- Renaming exported types
- Removing features

**Safe Changes**:
- Adding new optional parameters
- Adding new methods
- Adding new error types
- Enhancing documentation

## Development Environment

### Required Tools

- Go 1.21 or higher
- Make (optional, for Makefile commands)
- golangci-lint (for linting)

### Common Commands

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Build
make build

# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

## Lessons Learned

### What Worked Well

1. **Specification-First**: Having `spec.yaml` as source of truth prevented confusion
2. **Incremental Development**: Building one feature at a time with tests
3. **Clear Communication**: Specific prompts with context produced better results
4. **Documentation Focus**: Writing docs alongside code improved clarity

### Challenges Overcome

1. **Dual API Design**: Solved with optional configuration and clear error messages
2. **Test Organization**: Initially in subdirectory, moved to root per Go conventions
3. **HTTP Client Abstraction**: Internal package provides clean separation
4. **Type Safety**: Go's type system caught errors early in development

## Contributing

When extending this SDK with AI assistance:

1. **Start with spec.yaml**: Document the feature first
2. **Follow patterns**: Look at existing code for consistency
3. **Test thoroughly**: Add tests for new functionality
4. **Document everything**: Update README.md and add GoDoc
5. **Verify builds**: Run `make check` before committing

## Resources

### Project Documentation
- [README.md](./README.md) - User-facing documentation
- [CHANGELOG.md](./CHANGELOG.md) - Release history and changes
- [docs/spec.yaml](./docs/spec.yaml) - API specifications
- [examples/](./examples/) - Usage examples

### External References
- [ZipTax API Documentation](https://api.zip-tax.com/)
- [TaxCloud API Documentation](https://docs.taxcloud.com/)
- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)

## Version History

See [CHANGELOG.md](./CHANGELOG.md) for detailed release notes.

### v0.2.2-beta (Create Order from Cart)
- CreateOrderFromCart for converting TaxCloud carts into finalized orders
- CreateOrderFromCartRequest model with cartId and orderId
- Cart-to-order validation (cartId and orderId required)
- Use UpdateOrder after creation to set a completed date
- Comprehensive test coverage (92.3%)

### v0.2.1-beta (Cart Tax Calculation)
- CalculateCart with dual API routing (ZipTax default, TaxCloud when configured)
- Cart tax calculation models: request, response, line items, tax details
- TaxCloud cart models with structured addresses and TIC mapping
- CalculateCartResult interface for polymorphic return types
- Address parsing utility for TaxCloud request transformation
- Cart request validation (items, line items, currency, addresses)
- Comprehensive test coverage (92.7%)

### v0.2.0-beta (TaxCloud Integration)
- TaxCloud order management: CreateOrder, GetOrder, UpdateOrder, RefundOrder
- Dual API support: ZipTax (required) + TaxCloud (optional)
- Unified APIError handling across both APIs
- Historical date format corrected to YYYYMM
- 9-digit postal codes normalized to 5-digit
- GitHub Actions for version enforcement and automated releases
- Comprehensive test coverage (88.5%)

### v0.1.0 (Initial Release)
- ZipTax API support
- Tax rate lookups by address, geolocation, postal code
- Account metrics
- Retry logic with exponential backoff
- Context support, input validation, structured error handling

---

**Maintained with**: Claude (Anthropic AI Assistant)
**Development Approach**: AI-Pair Programming
**Last Updated**: February 2026
