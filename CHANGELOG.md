# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.3-beta] - 2026-03-20

### Added
- **Product Code Search (TIC)**: `SearchProductCodes` method for searching Taxability Information Codes by natural language description
  - Posts to ZipTax `POST /search/tic`
  - Returns ranked and scored results with TIC codes, labels, descriptions, and documentation
  - Uses ZipTax API key only (no TaxCloud credentials required)
- **Product Code Recommendation**: `RecommendProductCode` method for AI-powered TIC recommendation
  - Posts to ZipTax `POST /search/tic/recommend`
  - Returns a single best-match recommendation with higher accuracy
  - Slightly higher latency than `SearchProductCodes` due to AI processing
- Product code models in `models/product_codes.go`:
  - `ProductCodeSearchRequest`, `ProductCodeSearchResponse`, `ProductCodeSearchResult`
  - `ProductCodeRecommendation`, `ProductCodeRecommendationResponse`
- `ValidateProductQuery()` validation helper for non-empty query strings
- Product code search example in `examples/product_code_search/`

### Changed
- Test coverage improved from 92.3% to 93.0% on root package

## [0.2.2-beta] - 2026-03-11

### Added
- **Create Order from Cart**: `CreateOrderFromCart` method for converting previously calculated TaxCloud carts into finalized orders
  - Posts to TaxCloud `POST /tax/connections/{connectionId}/carts/orders`
  - Accepts `CreateOrderFromCartRequest` with `cartId` (required) and `orderId` (required)
  - Returns existing `OrderResponse` type for consistent return types
  - Validates `cartId` and `orderId` before making API call
  - Requires TaxCloud credentials to be configured during client initialization

## [0.2.1-beta] - 2026-02-27

### Added
- **Cart Tax Calculation**: `CalculateCart` method with dual API routing
  - Routes to ZipTax `/calculate/cart` by default
  - Automatically routes to TaxCloud `/tax/connections/{connectionId}/carts` when TaxCloud credentials are configured
  - Same input contract (`CalculateCartRequest`) regardless of backend
- `CalculateCartResult` interface for polymorphic return types (`*CalculateCartResponse` or `*TaxCloudCalculateCartResponse`)
- Cart request/response models in `models/cart.go`:
  - `CalculateCartRequest`, `CartItem`, `CartAddress`, `CartCurrency`, `CartLineItem`
  - `CalculateCartResponse`, `CartItemResponse`, `CartLineItemResponse`, `CartTax`
- TaxCloud cart models in `models/taxcloud_cart.go`:
  - `TaxCloudCalculateCartResponse`, `TaxCloudCartItemResponse`, `TaxCloudCartLineItemResponse`
  - `TaxCloudCalculateCartRequest`, `TaxCloudCartItem`, `TaxCloudCartLineItem` (internal transformation types)
- Address parsing utility (`ParseAddress`) for transforming single-string addresses to TaxCloud structured format
- Cart request validation: items count, currency, addresses, line item fields (price > 0, quantity > 0, max 250 items)
- TaxabilityCode to TIC mapping with nil-safe default (0)
- Auto-generated 0-based line item indices for TaxCloud requests
- Comprehensive test coverage for cart calculation (52 new tests across 3 test files)

### Changed
- Test coverage improved from 88.5% to 92.7% on root package
- Validation package now at 100% coverage

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
