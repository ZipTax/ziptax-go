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
- 🧪 Comprehensive test coverage (>91%)
- 📝 Full type safety with Go structs
- 🔐 Secure API key authentication
- 🌐 Support for US and Canadian addresses
- 📍 Geolocation-based lookups
- 🏷️ **Product Code (TIC) Search** - Search and AI-powered recommendation for Taxability Information Codes
- 🛒 **Cart Tax Calculation** - Calculate sales tax on shopping carts
- 🏪 **Merchant Management** - Provision tax compliance for the merchants on your platform
- 🧾 **Merchant Transactions** - Orders, refunds, and exemption certificates on a merchant's behalf
- 🩺 **Reference data & health** - TIC catalog, JSON Schema, service health, and account usage
- 📦 **TaxCloud Order Management** (deprecated) - see [Migration](#migration-from-direct-taxcloud-to-merchant-management)

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

The same client covers every endpoint in this README, including Merchant Management.
Merchant endpoints authenticate with your ZipTax API key; a merchant's own TaxCloud
credentials are stored server-side with [`SetMerchantCredentials`](#store-taxcloud-credentials-for-a-merchant).

**Enable the deprecated direct TaxCloud integration (optional):**

> **Deprecated.** These options configure calls straight to `api.v3.taxcloud.com`.
> Use [Merchant Management](#merchant-management) instead. See
> [Migration](#migration-from-direct-taxcloud-to-merchant-management).

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

`GetAccountMetrics` reads `/account/v60/metrics`, which reports one combined counter:

```go
metrics, err := client.GetAccountMetrics(ctx)
fmt.Printf("Usage: %d / %d (%.2f%%)\n",
    metrics.RequestCount, metrics.RequestLimit, metrics.UsagePercent)
```

For usage per entitlement, including the merchant counters, use
`GetDetailedAccountMetrics`, which reads `/account/metrics`:

```go
detailed, err := client.GetDetailedAccountMetrics(ctx)
fmt.Printf("Core Usage: %.2f%%\n", detailed.CoreUsagePercent)
fmt.Printf("Geo Usage: %.2f%%\n", detailed.GeoUsagePercent)
fmt.Printf("Merchant Usage: %.2f%%\n", detailed.MerchantUsagePercent)
```

> **Changed in v0.3.0-beta.** `V60AccountMetrics` previously declared the core/geo
> fields that `/account/metrics` returns, not the ones `/account/v60/metrics` sends,
> so they always decoded as zero. They are now `RequestCount`, `RequestLimit`, and
> `UsagePercent`. Use `GetDetailedAccountMetrics` for the core/geo/merchant breakdown.

### System Health and Reference Data

These endpoints are public and need no API key; the SDK sends yours anyway and the
API ignores it.

```go
// Full TIC catalog
catalog, err := client.GetTICCodes(ctx)
for _, entry := range catalog.TICList {
    fmt.Printf("%s: %s\n", entry.TIC.ID, entry.TIC.Title)
}

// JSON Schema for the product code search response
schema, err := client.GetTICSearchSchema(ctx)

// Service health
health, err := client.GetHealth(ctx)
fmt.Printf("healthy: %t (%d tax data records)\n",
    health.IsHealthy(), health.Components.TaxDataCount)

// Serving instance metadata
meta, err := client.GetSystemMetadata(ctx)
fmt.Printf("%s running %s\n", meta.Hostname, meta.GoVersion)
```

## Product Code Search (TIC)

Search for Taxability Information Codes (TICs) using natural language product descriptions. No TaxCloud credentials required -- these endpoints use the ZipTax API.

### Search Product Codes

Returns all matching TICs ranked and scored by relevance:

```go
response, err := client.SearchProductCodes(ctx, "baked goods sold in plastic packaging")
if err != nil {
    log.Fatal(err)
}

for _, result := range response.Results {
    fmt.Printf("TIC %s: %s (rank %s, score %s)\n",
        result.TicID, result.Label, result.Rank, result.Score)
}
```

### Recommend Product Code

Get an AI-powered best-match recommendation (slightly higher latency):

```go
response, err := client.RecommendProductCode(ctx, "baked goods sold in plastic packaging")
if err != nil {
    log.Fatal(err)
}

prediction := response.Predictions[0]
if prediction.Status == "success" {
    fmt.Printf("Recommended TIC: %s (%s)\n", prediction.TicID, prediction.Label)
    fmt.Printf("Description: %s\n", prediction.TicDescription)
}
```

### Using TICs with Cart Line Items

Use the returned `TicID` as the `TaxabilityCode` in cart line items:

```go
import (
    "log"
    "strconv"
)

ticID := response.Results[0].TicID
tic, err := strconv.ParseInt(ticID, 10, 64)
if err != nil {
    log.Fatalf("failed to parse TIC ID %q: %v", ticID, err)
}

lineItem := models.CartLineItem{
    ItemID:          "item-1",
    Price:           10.00,
    Quantity:        1,
    TaxabilityCode:  &tic,
}
```

## Cart Tax Calculation

Calculate sales tax on a shopping cart. The SDK routes the request automatically based on your client configuration:

- **Without TaxCloud credentials**: Routes to the ZipTax `/calculate/cart` API
- **With TaxCloud credentials**: Routes to the TaxCloud `/tax/connections/{connectionId}/carts` API — **deprecated**, use [`CalculateMerchantCart`](#calculate-cart-tax-for-a-merchant)

`CalculateCart` itself is not deprecated. Its default ZipTax branch is fully supported;
only the TaxCloud routing branch is superseded by Merchant Management.

```go
cartReq := &models.CalculateCartRequest{
    Items: []models.CartItem{
        {
            CustomerID: "customer-453",
            Currency:   models.CartCurrency{CurrencyCode: "USD"},
            Destination: models.CartAddress{
                Address: "200 Spectrum Center Dr, Irvine, CA 92618",
            },
            Origin: models.CartAddress{
                Address: "323 Washington Ave N, Minneapolis, MN 55401-2427",
            },
            LineItems: []models.CartLineItem{
                {
                    ItemID:   "item-1",
                    Price:    10.75,
                    Quantity: 1.5,
                },
            },
        },
    },
}

result, err := client.CalculateCart(ctx, cartReq)
if err != nil {
    log.Fatal(err)
}

// Use a type switch to handle the polymorphic response
switch r := result.(type) {
case *models.CalculateCartResponse:
    fmt.Printf("ZipTax cart total tax: %.2f\n", r.Items[0].LineItems[0].Tax.Amount)
case *models.TaxCloudCalculateCartResponse:
    fmt.Printf("TaxCloud cart ID: %s\n", r.Items[0].CartID)
}
```

### Create Order from Cart

After calculating a cart with TaxCloud credentials, convert it into a finalized order. This requires the `cartId` returned from a previous `CalculateCart` call:

```go
req := &models.CreateOrderFromCartRequest{
    CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab", // from CalculateCart response
    OrderID: "my-order-1",                            // your internal order ID
}

order, err := client.CreateOrderFromCart(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order created: %s\n", order.OrderID)
```

To set a completed date on the order, use `UpdateOrder` after creation.

## Merchant Management

Merchant Management lets platforms and SaaS businesses provision tax compliance for
their own customers. You create a merchant for each seller on your platform, then
manage that merchant's compliance through ZipTax.

Every merchant endpoint authenticates with your ZipTax API key. No TaxCloud
credentials are configured on the client.

> **Private Preview.** Contact [support@zip.tax](mailto:support@zip.tax) for access.
> Self-Managed Cart Calculation is still in active development; request bodies,
> responses, and supported fields may change before general availability.

### Compliance models

Each merchant uses one of two models, chosen once at creation via `MerchantType`:

| | `MerchantTypeSelfManaged` | `MerchantTypeTaxCloud` (default) |
| --- | --- | --- |
| Activation | Active immediately, no invite | TaxCloud invite sent to `ContactEmail` |
| Status on reads | `external_compliance` | `taxcloud_invited` → `taxcloud_connected` |
| Registration, filing, remittance | The merchant handles their own | TaxCloud handles all three |
| Available endpoints | `CalculateMerchantCart` only | All merchant endpoints |

Everything except `CalculateMerchantCart` returns HTTP 403 for a self-managed
merchant, which has no TaxCloud connection to store or read state in.

### Create and manage merchants

```go
resp, err := client.CreateMerchant(ctx, &models.CreateMerchantRequest{
    MerchantName: "Acme Supply Co",
    ContactEmail: "ops@acme.example",
    ReferenceID:  "seller-42",              // your own identifier
    MerchantType: models.MerchantTypeTaxCloud,
})
if err != nil {
    log.Fatal(err)
}
merchantID := resp.MerchantID

// Read one, or list them all
merchant, err := client.GetMerchant(ctx, merchantID)
fmt.Printf("%s is %s\n", merchant.MerchantName, merchant.Status)

merchants, err := client.ListMerchants(ctx)

// Update mutable fields. MerchantName is required even when unchanged.
_, err = client.UpdateMerchant(ctx, &models.UpdateMerchantRequest{
    MerchantID: merchantID,
    Update: models.MerchantUpdate{
        MerchantName: "Acme Supply Co",
        ContactEmail: "billing@acme.example",
    },
})

// Soft-delete
_, err = client.DeleteMerchant(ctx, merchantID)
```

### Store TaxCloud credentials for a merchant

For a merchant that already has a TaxCloud account, store their credentials so
ZipTax can act on their behalf. The API encrypts them at rest. Merchants created
with an invite establish credentials by accepting it instead.

```go
_, err := client.SetMerchantCredentials(ctx, &models.SetMerchantCredentialsRequest{
    MerchantID:   merchantID,
    ConnectionID: "25eb9b97-5acb-492d-b720-c03e79cf715a",
    APIKey:       taxCloudAPIKey,
})

// Revoke them later
_, err = client.DeleteMerchantCredentials(ctx, merchantID)
```

### Calculate cart tax for a merchant

One request contract serves both compliance models, so you never branch on merchant
type when building a request. Check `IsSelfManaged()` on the response to tell which
engine ran.

```go
resp, err := client.CalculateMerchantCart(ctx, &models.MerchantCalculateCartRequest{
    MerchantID: merchantID,
    Items: []models.MerchantCart{
        {
            CartID:     "my-cart-1",
            CustomerID: "customer-453",
            Currency:   models.Currency{}, // defaults to USD
            Origin: models.TaxCloudAddress{
                Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
            },
            Destination: models.TaxCloudAddress{
                Line1: "200 Spectrum Center Dr", City: "Irvine", State: "CA", Zip: "92618",
            },
            LineItems: []models.MerchantCartLineItem{
                {Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5},
            },
        },
    },
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tax: %.2f\n", resp.Items[0].LineItems[0].Tax.Amount)

if resp.IsSelfManaged() {
    // Calculated in-process by the ZipTax rate engine. Stateless: this cartId
    // cannot be captured as an order.
} else {
    // Forwarded to TaxCloud. Capture the cartId with CreateMerchantOrderFromCart.
}
```

Self-managed calculation **rejects** rather than ignores fields it cannot honour:
discounts, exemptions, `DeliveredBySeller`, `ProductID`, and any currency other than
USD. It also uses ZipTax TICs (10001 shipping, 11000 handling) rather than TaxCloud's.

### Record orders

Capture a calculated cart:

```go
order, err := client.CreateMerchantOrderFromCart(ctx, &models.MerchantCreateOrderFromCartRequest{
    MerchantID: merchantID,
    CartID:     "my-cart-1",
    OrderID:    "my-order-1",
})
```

Or record an order directly, supplying the tax your checkout collected:

```go
order, err := client.CreateMerchantOrder(ctx, &models.MerchantCreateOrderRequest{
    MerchantID:      merchantID,
    OrderID:         "my-order-1",
    CustomerID:      "customer-453",
    TransactionDate: "2026-08-01T14:00:00Z",
    CompletedDate:   "2026-08-02T09:15:00Z",
    Currency:        models.Currency{},
    Origin:          origin,
    Destination:     destination,
    LineItems: []models.MerchantOrderLineItem{
        {
            Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5,
            Tax: models.Tax{Amount: 1.31, Rate: 0.08125},
        },
    },
})
```

Read and update:

```go
order, err := client.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
    MerchantID: merchantID,
    OrderID:    "my-order-1",
    Expand:     models.ExpandRefunds, // include refunds in the response
})

// Setting the completed date marks the order shipped, creating the tax liability.
order, err = client.UpdateMerchantOrder(ctx, &models.MerchantUpdateOrderRequest{
    MerchantID:    merchantID,
    OrderID:       "my-order-1",
    CompletedDate: "2026-08-03T10:00:00Z",
})
```

### Refunds

Omit `Items` to refund the whole order. Refund prices and tax are calculated
automatically from the order.

```go
refund, err := client.CreateMerchantRefund(ctx, &models.MerchantCreateRefundRequest{
    MerchantID: merchantID,
    OrderID:    "my-order-1",
    Items: []models.MerchantRefundRequestItem{
        {ItemID: "item-1", Quantity: 1},
    },
})
```

> Refunds are not idempotent. A duplicate submission records a duplicate refund, so
> do not retry them blindly.

### Exemption certificates

```go
cert, err := client.CreateMerchantCertificate(ctx, &models.MerchantCreateCertificateRequest{
    MerchantID:           merchantID,
    CustomerID:           "customer-453",
    CustomerName:         "Acme Reseller LLC",
    CustomerBusinessType: models.BusinessTypeRetailTrade,
    Reason:               models.ExemptionReasonResale,
    ReasonDescription:    "Resale",  // max 20 characters
    Address: models.TaxCloudAddress{
        Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
    },
    States: []models.CertificateState{{Abbreviation: "MN"}},
})
```

Apply it to a cart or order by setting `Exemption.ExemptionID` to `cert.CertificateID`
for the same `CustomerID`.

Page through them with a cursor:

```go
req := &models.MerchantListCertificatesRequest{MerchantID: merchantID, Limit: 50}
for {
    page, err := client.ListMerchantCertificates(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    for _, c := range page.Items {
        fmt.Println(c.CertificateID, c.CustomerName, c.IsActive())
    }
    if page.NextCursor == "" {
        break
    }
    req.Cursor = page.NextCursor
}

// Disable a certificate so it can no longer be applied
_, err = client.DeleteMerchantCertificate(ctx, merchantID, cert.CertificateID)
```

## Migration from direct TaxCloud to Merchant Management

The direct TaxCloud integration calls `api.v3.taxcloud.com` with a connection ID and
TaxCloud API key held on the client. That path is no longer covered by the ZipTax API
documentation. It still works and is unchanged in behavior, but is deprecated.

Merchant Management reaches the same TaxCloud capabilities through the ZipTax API,
authenticated with your ZipTax key alone, addressing merchants by ID.

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

**What changes in your code:**

1. Drop `WithTaxCloudConnectionID` and `WithTaxCloudAPIKey` from `NewClient`.
2. Create a merchant with `CreateMerchant`, then call `SetMerchantCredentials` once
   with the connection ID and TaxCloud key you were passing to the client. Store the
   returned `MerchantID`.
3. Pass that `MerchantID` on every merchant call in place of the client-level credentials.
4. Requests move from flat addresses to structured `TaxCloudAddress` values, and line
   items carry an explicit zero-based `Index`.

```go
// Before
client, _ := ziptax.NewClient(ziptaxKey,
    ziptax.WithTaxCloudConnectionID(connectionID),
    ziptax.WithTaxCloudAPIKey(taxCloudKey),
)
order, err := client.GetOrder(ctx, "my-order-1")

// After
client, _ := ziptax.NewClient(ziptaxKey)
// once, at onboarding:
//   m, _ := client.CreateMerchant(ctx, &models.CreateMerchantRequest{MerchantName: "Acme Supply Co"})
//   client.SetMerchantCredentials(ctx, &models.SetMerchantCredentialsRequest{
//       MerchantID: m.MerchantID, ConnectionID: connectionID, APIKey: taxCloudKey,
//   })
order, err := client.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
    MerchantID: merchantID,
    OrderID:    "my-order-1",
})
```

## TaxCloud Order Management

> **Deprecated.** These functions call TaxCloud directly and require TaxCloud
> credentials on the client. Use [Merchant Management](#merchant-management) instead;
> see [Migration](#migration-from-direct-taxcloud-to-merchant-management). They remain
> supported for now and their behavior is unchanged.

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

### Retries and non-idempotent operations

Retries are **off by default**; `WithMaxRetries` turns them on. When enabled, the
client retries transport errors, 5xx responses, and 429.

Some operations are never retried, whatever `WithMaxRetries` is set to, because
each request creates a new record server-side rather than converging on the same
state. Retrying one of those after a timeout would duplicate it, and nothing in
the returned error would tell you it happened:

| Never retried | Why |
| --- | --- |
| `CreateMerchantRefund` | A duplicate submission records a duplicate refund |
| `RefundOrder` (deprecated) | Same TaxCloud endpoint |
| `CreateMerchant` | The server assigns the merchant ID, so a repeat creates a second merchant |
| `CreateMerchantCertificate` | The server assigns the certificate ID, so a repeat creates a second certificate |

These surface the transport error or 5xx instead, leaving the outcome genuinely
unknown rather than silently doubled. Reconcile by reading current state before
resubmitting:

```go
refund, err := client.CreateMerchantRefund(ctx, req)
if err != nil {
    // The refund may or may not have been recorded. Check before retrying.
    order, getErr := client.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
        MerchantID: merchantID,
        OrderID:    orderID,
        Expand:     models.ExpandRefunds,
    })
    if getErr == nil && len(order.Refunds) > 0 {
        // It landed; do not resubmit.
    }
}
```

Order creation is **not** in this group: `CreateMerchantOrder` and
`CreateMerchantOrderFromCart` are keyed by the `OrderID` you supply, so a repeat
conflicts rather than duplicating. Cart calculation and all reads are documented
by the API as safe to retry.
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
- [Merchant Management](./examples/merchant_management) - Provision a merchant, calculate a cart, record an order, refund it
- [Product Code Search](./examples/product_code_search) - TIC search and AI recommendation
- [Concurrent Usage](./examples/concurrent_usage) - Parallel requests with goroutines
- [Error Handling](./examples/error_handling) - Proper error handling patterns
- [Context Timeout](./examples/context_timeout) - Using context for timeouts and cancellation
- [TaxCloud Order](./examples/taxcloud_order) - Direct TaxCloud order management (deprecated)

## API Reference

For detailed API documentation, see [pkg.go.dev](https://pkg.go.dev/github.com/ziptax/ziptax-go).

The API itself is documented at [docs.zip.tax](https://docs.zip.tax), with the
canonical OpenAPI document at [docs.zip.tax/openapi.json](https://docs.zip.tax/openapi.json).

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