# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0-beta] - 2026-02-16

### Added
- **TaxCloud Order Management**: Full order lifecycle support via the TaxCloud API
  - `CreateOrder` - Create orders from marketplace transactions or bulk uploads
  - `GetOrder` - Retrieve existing orders by ID
  - `UpdateOrder` - Update an order's completed date
  - `RefundOrder` - Create full or partial refunds against orders
- TaxCloud client configuration options:
  - `WithTaxCloudConnectionID(id string)` - Set TaxCloud Connection ID
  - `WithTaxCloudAPIKey(key string)` - Set TaxCloud API Key
  - `WithTaxCloudBaseURL(url string)` - Set custom TaxCloud API base URL
- TaxCloud data models in `models/taxcloud_orders.go`:
  - `CreateOrderRequest`, `OrderResponse`, `UpdateOrderRequest`
  - `RefundTransactionRequest`, `RefundTransactionResponse`
  - `TaxCloudAddress`, `CartItemWithTax`, `Tax`, `Currency`, `Exemption`
- `ErrTaxCloudNotConfigured` sentinel error for missing TaxCloud credentials
- `HasTaxCloudCredentials()` config method to check credential availability
- Internal HTTP client methods: `Post()`, `Patch()`, `GetWithOptions()`
- TaxCloud error response parsing (alongside existing ZipTax format)
- `wrapError()` helper to convert internal API errors to public `APIError` type
- `NormalizePostalCode()` validation helper to strip 9-digit suffix
- TaxCloud order example in `examples/taxcloud_order/`
- GitHub Actions workflows:
  - `version-bump-check.yml` - PR semantic version validation
  - `release.yml` - Automated release creation on version tags
- `CLAUDE.md` - AI-assisted development documentation
- `.github/VERSION_BUMP_GUIDE.md` - Contributor versioning guide
- `.github/WORKFLOWS.md` - Workflow documentation
- This `CHANGELOG.md`

### Changed
- SDK now supports dual APIs: ZipTax (required) and TaxCloud (optional)
- Default UserAgent now dynamically uses `Version` constant (`"ziptax-go/" + Version`)
- Historical date validation updated from `YYYY-MM` to `YYYYMM` format to match ZipTax API
- `GetOrder` refactored to use internal HTTP client with retry support and structured errors
- All client methods now use `wrapError()` for consistent `APIError` extraction via `errors.As()`
- 9-digit postal codes are now normalized to 5-digit before sending to the API
- Month range validation (01-12) added to historical date validator
- Comprehensive unit test coverage for all TaxCloud methods (coverage: 88.5%)

### Fixed
- `errors.As(err, &ziptax.APIError{})` now works correctly across all API methods
- Historical date validation rejects invalid month values (00, 13, etc.)
- 9-digit postal codes no longer cause 422 errors from the ZipTax API
- `gofmt` formatting issue in `models/taxcloud_orders.go`

## [0.1.0] - 2025-01-15

### Added
- Initial release of the ZipTax Go SDK
- `GetSalesTaxByAddress` - Tax rate lookup by street address
- `GetSalesTaxByGeoLocation` - Tax rate lookup by latitude/longitude
- `GetRatesByPostalCode` - Tax rate lookup by US postal code
- `GetAccountMetrics` - Account usage metrics
- Functional options pattern for client configuration
- Automatic retry with exponential backoff
- Context support for timeouts and cancellation
- Input validation for addresses, coordinates, postal codes, and dates
- Structured error types: `APIError`, `ValidationError`, sentinel errors
- Request/response logging via `WithLogger()` option
- Concurrent-safe client design
- Examples for basic usage, concurrent requests, error handling, and context timeouts
