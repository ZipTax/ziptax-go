# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0-beta] - 2026-08-07

Aligns the SDK with the v6.0 API surface documented at https://docs.zip.tax.

The headline change is **Merchant Management**: platforms can now provision tax
compliance for their own customers through the ZipTax API, using only the ZipTax
API key. This supersedes the direct TaxCloud integration, which is now deprecated
but still works. See the Migration section of the README.

Merchant Management is a Private Preview feature and Self-Managed Cart Calculation
is still in active development. Contact support@zip.tax for access. Request and
response shapes on those endpoints may change before general availability.

### Added
- **Merchant Management** (`merchant.go`, `models/merchant.go`):
  - `CreateMerchant`, `UpdateMerchant`, `DeleteMerchant`, `GetMerchant`, `ListMerchants`
  - `SetMerchantCredentials`, `DeleteMerchantCredentials`
  - Two compliance models via `MerchantType`: `MerchantTypeTaxCloud` (default) and `MerchantTypeSelfManaged`
  - `Merchant.IsSelfManaged()` helper and the four `MerchantStatus*` lifecycle constants
- **Merchant Transactions** (`merchant_transactions.go`):
  - `CalculateMerchantCart` posts to `POST /merchant/cart/calculate` and serves both
    compliance models from one request contract. `MerchantCalculateCartResponse.IsSelfManaged()`
    distinguishes a ZipTax in-process calculation from a TaxCloud one.
  - `CreateMerchantOrder`, `CreateMerchantOrderFromCart`, `GetMerchantOrder`, `UpdateMerchantOrder`
  - `CreateMerchantRefund`
  - `CreateMerchantCertificate`, `GetMerchantCertificate`, `ListMerchantCertificates`,
    `DeleteMerchantCertificate` for exemption certificates, with cursor pagination
  - Line-item and order-level discounts via `MerchantDiscounts`
- **Reference data and system endpoints** (`system.go`, `models/system.go`):
  - `GetTICCodes` for the full TIC catalog (`GET /data/tic`)
  - `GetTICSearchSchema` for the product code search JSON Schema (`GET /schemas/ticsearch`)
  - `GetHealth` (`GET /system/health`) with a `HealthResponse.IsHealthy()` helper
  - `GetSystemMetadata` (`GET /system/metadata`)
  - `GetDetailedAccountMetrics` (`GET /account/metrics`) reporting usage per entitlement,
    including the new merchant request counters
- Merchant validation helpers in `internal/validation/merchant.go`, covering merchant IDs,
  names, compliance models, credentials, carts, orders, certificates, and list pagination
- Merchant management example in `examples/merchant_management/`

### Changed
- **Fixed `V60AccountMetrics` to match what `/account/v60/metrics` actually returns.**
  The struct carried the unversioned endpoint's core/geo field set, so
  `CoreRequestCount`, `CoreRequestLimit`, `CoreUsagePercent`, `GeoEnabled`,
  `GeoRequestCount`, `GeoRequestLimit`, and `GeoUsagePercent` never populated and
  always read as zero. They are replaced by the fields the endpoint really sends:
  `RequestCount`, `RequestLimit`, and `UsagePercent`. `IsActive` and `Message` are
  unchanged. This breaks compilation for code reading the old fields; those fields
  were already returning zero, so no correct behavior depended on them. For the
  core/geo/merchant breakdown, use the new `GetDetailedAccountMetrics`.
- Test coverage: 91.4% on the root package, 100% on `internal/validation`

### Fixed
- **Retried requests no longer send an empty body.** `DoWithRetry` cloned the
  request, but `http.Request.Clone` copies `Body` by reference, so once the first
  attempt drained it every later attempt sent 0 bytes while `ContentLength` still
  claimed the original size. `net/http` rejected that locally with
  `ContentLength=N with Body length 0`, so the retry never reached the server and
  the real API error was replaced by a confusing transport error. The body is now
  rewound from `Request.GetBody` before each attempt. This affected every retried
  POST and PATCH; GET was unaffected.
- **Non-idempotent operations are no longer retried.** With retries enabled, a
  transport error or 5xx on a refund could resubmit it and record a duplicate,
  with nothing in the returned error to indicate it. These now make exactly one
  attempt regardless of `WithMaxRetries`:
  `CreateMerchantRefund`, `RefundOrder` (deprecated), `CreateMerchant`, and
  `CreateMerchantCertificate`. The first two duplicate a refund; the last two
  have server-assigned IDs, so a repeat creates a second record. Reads, cart
  calculation, and the order operations (keyed by a caller-supplied `orderId`)
  keep the configured retry behavior. See the README for how to reconcile.

### Deprecated
The direct TaxCloud integration still works and is unchanged in behavior. It calls
`api.v3.taxcloud.com` with a connection ID and TaxCloud API key held on the client,
a path no longer covered by the ZipTax API documentation.

| Deprecated | Replacement |
| --- | --- |
| `CreateOrder` | `CreateMerchantOrder` |
| `GetOrder` | `GetMerchantOrder` |
| `UpdateOrder` | `UpdateMerchantOrder` |
| `RefundOrder` | `CreateMerchantRefund` |
| `CreateOrderFromCart` | `CreateMerchantOrderFromCart` |
| `CalculateCart` (TaxCloud branch only) | `CalculateMerchantCart` |
| `WithTaxCloudConnectionID`, `WithTaxCloudAPIKey` | `SetMerchantCredentials` |
| `WithTaxCloudBaseURL` | `WithBaseURL` |
| `Config.HasTaxCloudCredentials` | not needed; merchant endpoints use the ZipTax key |

`CalculateCart` itself is not deprecated: its default ZipTax branch (`POST /calculate/cart`)
remains supported. Only the TaxCloud routing branch is superseded.

### Notes
- `/merchant/credentials/get` exists in the API but is not part of the public
  documentation, so it is deliberately not exposed by the SDK.
- `GET /data/tic` can also serve XML. The SDK requests JSON only.

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
