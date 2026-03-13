package models

// CreateOrderRequest represents the request payload for creating an order in TaxCloud.
type CreateOrderRequest struct {
	// OrderID is the order ID in external system (required)
	OrderID string `json:"orderId"`

	// CustomerID is the customer ID in external system (required)
	CustomerID string `json:"customerId"`

	// TransactionDate is the RFC3339 datetime string when order was purchased (required)
	TransactionDate string `json:"transactionDate"`

	// CompletedDate is the RFC3339 datetime string when order was shipped/completed (required)
	CompletedDate string `json:"completedDate"`

	// Origin is the origin address of the order (required)
	Origin TaxCloudAddress `json:"origin"`

	// Destination is the destination address of the order (required)
	Destination TaxCloudAddress `json:"destination"`

	// LineItems is the array of line items in the order with tax calculations (required)
	LineItems []CartItemWithTax `json:"lineItems"`

	// Currency is the currency information for the order (required)
	Currency *Currency `json:"currency,omitempty"`

	// Channel is the sales channel (e.g., amazon, ebay, walmart) for tax exclusion rules (optional)
	Channel *string `json:"channel,omitempty"`

	// DeliveredBySeller indicates whether the seller directly delivered the order (optional)
	DeliveredBySeller *bool `json:"deliveredBySeller,omitempty"`

	// ExcludeFromFiling indicates whether to exclude this order from tax filing (optional)
	ExcludeFromFiling *bool `json:"excludeFromFiling,omitempty"`

	// Exemption is the exemption certificate information for the order (optional)
	Exemption *Exemption `json:"exemption,omitempty"`
}

// OrderResponse represents the response after successfully creating an order in TaxCloud.
type OrderResponse struct {
	// OrderID is the order ID in external system
	OrderID string `json:"orderId"`

	// CustomerID is the customer ID in external system
	CustomerID string `json:"customerId"`

	// ConnectionID is the TaxCloud connection ID used for this order
	ConnectionID string `json:"connectionId"`

	// TransactionDate is the RFC3339 datetime string when order was purchased
	TransactionDate string `json:"transactionDate"`

	// CompletedDate is the RFC3339 datetime string when order was shipped/completed
	CompletedDate string `json:"completedDate"`

	// Origin is the origin address of the order
	Origin TaxCloudAddressResponse `json:"origin"`

	// Destination is the destination address of the order
	Destination TaxCloudAddressResponse `json:"destination"`

	// LineItems is the array of line items with tax calculations
	LineItems []CartItemWithTaxResponse `json:"lineItems"`

	// Currency is the currency information
	Currency CurrencyResponse `json:"currency"`

	// Channel is the sales channel for the order
	Channel *string `json:"channel,omitempty"`

	// DeliveredBySeller indicates whether seller directly delivered the order
	DeliveredBySeller bool `json:"deliveredBySeller"`

	// ExcludeFromFiling indicates whether order is excluded from tax filing
	ExcludeFromFiling bool `json:"excludeFromFiling"`

	// Exemption is the exemption information
	Exemption *Exemption `json:"exemption,omitempty"`
}

// TaxCloudAddress represents an address structure for TaxCloud orders.
type TaxCloudAddress struct {
	// Line1 is the first line of address (street, PO Box, or building) (required)
	Line1 string `json:"line1"`

	// Line2 is the second line of address (apartment or suite number) (optional)
	Line2 *string `json:"line2,omitempty"`

	// City is the city or post-town (required)
	City string `json:"city"`

	// State is the state, province, county or large territorial division (required)
	State string `json:"state"`

	// Zip is the postal or ZIP code (required)
	Zip string `json:"zip"`

	// CountryCode is the ISO 3166-1 alpha-2 country code (optional, defaults to US)
	CountryCode *string `json:"countryCode,omitempty"`
}

// TaxCloudAddressResponse represents an address response structure from TaxCloud.
type TaxCloudAddressResponse struct {
	// Line1 is the first line of address
	Line1 string `json:"line1"`

	// Line2 is the second line of address
	Line2 *string `json:"line2,omitempty"`

	// City is the city or post-town
	City string `json:"city"`

	// State is the state abbreviation
	State string `json:"state"`

	// Zip is the postal or ZIP code
	Zip string `json:"zip"`

	// CountryCode is the ISO 3166-1 alpha-2 country code
	CountryCode string `json:"countryCode"`
}

// CartItemWithTax represents a cart line item with tax calculation for order creation.
type CartItemWithTax struct {
	// Index is the position/index of item within the cart (required)
	Index int64 `json:"index"`

	// ItemID is the unique identifier for the cart item (required)
	ItemID string `json:"itemId"`

	// Price is the unit price of the item (required)
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (required)
	Quantity float64 `json:"quantity"`

	// Tax is the tax information for the item (required)
	Tax Tax `json:"tax"`

	// ProductID is the product ID from product catalog (optional, must match existing product)
	ProductID *string `json:"productId,omitempty"`

	// TIC is the Taxability Information Code (optional, defaults to 0 if not provided)
	TIC *int64 `json:"tic,omitempty"`
}

// CartItemWithTaxResponse represents a cart line item response from TaxCloud.
type CartItemWithTaxResponse struct {
	// Index is the position/index of item within the cart
	Index int64 `json:"index"`

	// ItemID is the unique identifier for the cart item
	ItemID string `json:"itemId"`

	// Price is the unit price of the item
	Price float64 `json:"price"`

	// Quantity is the quantity of the item
	Quantity float64 `json:"quantity"`

	// Tax is the tax information for the item
	Tax Tax `json:"tax"`

	// TIC is the Taxability Information Code
	TIC int64 `json:"tic"`
}

// Tax represents tax calculation details for a cart item.
type Tax struct {
	// Amount is the tax amount calculated for the item (required)
	Amount float64 `json:"amount"`

	// Rate is the tax rate applied in decimal format (required)
	Rate float64 `json:"rate"`
}

// Currency represents currency information for an order.
type Currency struct {
	// CurrencyCode is the ISO currency code (optional, defaults to USD)
	CurrencyCode *string `json:"currencyCode,omitempty"`
}

// CurrencyResponse represents currency response from TaxCloud.
type CurrencyResponse struct {
	// CurrencyCode is the ISO currency code
	CurrencyCode string `json:"currencyCode"`
}

// Exemption represents tax exemption certificate information.
type Exemption struct {
	// ExemptionID is the ID of exemption certificate used for customer (optional)
	// If provided, IsExempt is assumed true
	ExemptionID *string `json:"exemptionId,omitempty"`

	// IsExempt indicates whether customer is exempt from tax (optional)
	IsExempt *bool `json:"isExempt,omitempty"`
}

// CreateOrderFromCartRequest represents the request payload for creating an order from a
// previously calculated cart in TaxCloud. The user must have previously called CalculateCart
// with TaxCloud credentials and stored the returned cartId from the TaxCloudCartItemResponse.
//
// The TaxCloud /carts/orders endpoint only accepts cartId and orderId. To set a completed
// date on the order, use UpdateOrder after creation.
//
// Example:
//
//	req := &models.CreateOrderFromCartRequest{
//		CartID:  "ce4a1234-5678-90ab-cdef-1234567890ab",
//		OrderID: "my-order-1",
//	}
type CreateOrderFromCartRequest struct {
	// CartID is the cart ID from a previous TaxCloud CalculateCart response (required).
	// Identifies the cart to convert into an order.
	CartID string `json:"cartId"`

	// OrderID is the user's internal order ID for cross-referencing (required).
	// Must be unique per connection to ensure accurate tax reporting.
	OrderID string `json:"orderId"`
}

// UpdateOrderRequest represents the request payload for updating an order in TaxCloud.
// Currently only the completedDate can be updated.
type UpdateOrderRequest struct {
	// CompletedDate is the RFC3339 datetime string when order was shipped/completed (required)
	CompletedDate string `json:"completedDate"`
}

// RefundTransactionRequest represents the request payload for creating a refund against an order.
type RefundTransactionRequest struct {
	// Items is the array of items to refund (optional)
	// If empty list or omitted, entire order will be refunded
	Items []CartItemRefundWithTaxRequest `json:"items,omitempty"`

	// ReturnedDate should only be included if this return is a change to a previously filed sales tax return
	// This triggers an Amended Sales Tax Return (not typically recommended)
	ReturnedDate *string `json:"returnedDate,omitempty"`
}

// CartItemRefundWithTaxRequest represents a cart line item to be refunded.
type CartItemRefundWithTaxRequest struct {
	// ItemID is the unique identifier for the cart item to refund (required)
	ItemID string `json:"itemId"`

	// Quantity is the quantity of the item to refund (required)
	Quantity float64 `json:"quantity"`
}

// RefundTransactionResponse represents the response after successfully creating a refund.
type RefundTransactionResponse struct {
	// ConnectionID is the TaxCloud connection ID used for this refund
	ConnectionID string `json:"connectionId"`

	// CreatedDate is the RFC3339 datetime string when the refund was created
	CreatedDate string `json:"createdDate"`

	// Items is the array of refunded line items with tax calculations
	Items []CartItemRefundWithTaxResponse `json:"items"`

	// ReturnedDate is the RFC3339 datetime string when the refund took effect (optional)
	ReturnedDate *string `json:"returnedDate,omitempty"`
}

// CartItemRefundWithTaxResponse represents a refunded cart line item response from TaxCloud.
type CartItemRefundWithTaxResponse struct {
	// Index is the position/index of item within the cart
	Index int64 `json:"index"`

	// ItemID is the unique identifier for the cart item
	ItemID string `json:"itemId"`

	// Price is the price of the refunded item
	Price float64 `json:"price"`

	// Quantity is the quantity of the item refunded
	Quantity float64 `json:"quantity"`

	// Tax is the tax information for the refunded item
	Tax RefundTax `json:"tax"`

	// TIC is the Taxability Information Code (optional)
	TIC *int64 `json:"tic,omitempty"`
}

// RefundTax represents tax details for a refunded item.
type RefundTax struct {
	// Amount is the tax amount refunded for the item
	Amount float64 `json:"amount"`
}
