package models

// TaxCloudCalculateCartResponse represents the response from TaxCloud cart tax calculation.
// Returned by Client.CalculateCart when the client is configured with TaxCloud credentials.
//
// It contains the connectionId, transactionDate, and an array of cart results with
// TaxCloud-style structured addresses and line items with tax details.
// It implements the CalculateCartResult interface.
type TaxCloudCalculateCartResponse struct {
	// ConnectionID is the TaxCloud Connection ID used for this cart calculation
	ConnectionID string `json:"connectionId"`

	// Items is the array of cart results with calculated tax information
	Items []TaxCloudCartItemResponse `json:"items"`

	// TransactionDate is the RFC3339 datetime string the cart was calculated for
	TransactionDate string `json:"transactionDate"`
}

// isCalculateCartResult implements the CalculateCartResult interface for TaxCloudCalculateCartResponse.
func (*TaxCloudCalculateCartResponse) isCalculateCartResult() {}

// TaxCloudCartItemResponse represents a single cart response from TaxCloud
// with calculated tax information per line item.
type TaxCloudCartItemResponse struct {
	// CartID is the ID representing this cart (auto-generated if not provided in the request)
	CartID string `json:"cartId"`

	// CustomerID is the customer identifier (echoed from request)
	CustomerID string `json:"customerId"`

	// Currency is the currency information
	Currency CurrencyResponse `json:"currency"`

	// DeliveredBySeller indicates whether the seller directly delivered the order
	DeliveredBySeller bool `json:"deliveredBySeller"`

	// Destination is the destination address (structured format from TaxCloud)
	Destination TaxCloudAddressResponse `json:"destination"`

	// Origin is the origin address (structured format from TaxCloud)
	Origin TaxCloudAddressResponse `json:"origin"`

	// Exemption is the exemption information
	Exemption Exemption `json:"exemption"`

	// LineItems is the array of line items with calculated tax information
	LineItems []TaxCloudCartLineItemResponse `json:"lineItems"`
}

// TaxCloudCartLineItemResponse represents a line item in the TaxCloud cart response
// with calculated tax rate and amount.
type TaxCloudCartLineItemResponse struct {
	// Index is the position/index of item within the cart (0-based)
	Index int64 `json:"index"`

	// ItemID is the unique identifier for the line item (echoed from request)
	ItemID string `json:"itemId"`

	// Price is the unit price of the item (echoed from request)
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (echoed from request)
	Quantity float64 `json:"quantity"`

	// Tax is the calculated tax information for this line item
	Tax Tax `json:"tax"`

	// TIC is the Taxability Information Code (mapped from taxabilityCode, defaults to 0)
	TIC *int64 `json:"tic,omitempty"`
}

// TaxCloudCalculateCartRequest represents the internal request structure sent to the TaxCloud API
// after transforming from the shared CalculateCartRequest.
// This type is not exported to users; it is used internally by the SDK for request transformation.
type TaxCloudCalculateCartRequest struct {
	// Items is the array of cart items in TaxCloud format
	Items []TaxCloudCartItem `json:"items"`
}

// TaxCloudCartItem represents a single cart in TaxCloud format with structured addresses.
type TaxCloudCartItem struct {
	// CustomerID is the customer identifier
	CustomerID string `json:"customerId"`

	// Currency is the currency information
	Currency CartCurrency `json:"currency"`

	// Destination is the destination address (structured format for TaxCloud)
	Destination TaxCloudAddress `json:"destination"`

	// Origin is the origin address (structured format for TaxCloud)
	Origin TaxCloudAddress `json:"origin"`

	// LineItems is the array of line items in TaxCloud format
	LineItems []TaxCloudCartLineItem `json:"lineItems"`
}

// TaxCloudCartLineItem represents a line item in TaxCloud cart request format.
type TaxCloudCartLineItem struct {
	// Index is the position/index of item within the cart (0-based, auto-generated)
	Index int64 `json:"index"`

	// ItemID is the unique identifier for the line item
	ItemID string `json:"itemId"`

	// Price is the unit price of the item
	Price float64 `json:"price"`

	// Quantity is the quantity of the item
	Quantity float64 `json:"quantity"`

	// TIC is the Taxability Information Code (mapped from taxabilityCode, defaults to 0)
	TIC int64 `json:"tic"`
}
