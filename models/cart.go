package models

// CalculateCartResult is the interface implemented by both *CalculateCartResponse and
// *TaxCloudCalculateCartResponse. It is used as the return type of Client.CalculateCart
// to support dual API routing.
//
// Use a type switch or type assertion to access the concrete response type.
type CalculateCartResult interface {
	isCalculateCartResult()
}

// CalculateCartRequest represents the request payload for calculating sales tax on a shopping cart.
// This is the shared input contract for both the ZipTax and TaxCloud backends.
// The SDK routes the request to the appropriate API based on client configuration.
//
// Example:
//
//	req := &models.CalculateCartRequest{
//		Items: []models.CartItem{
//			{
//				CustomerID: "customer-453",
//				Currency:   models.CartCurrency{CurrencyCode: "USD"},
//				Destination: models.CartAddress{
//					Address: "200 Spectrum Center Dr, Irvine, CA 92618-1905",
//				},
//				Origin: models.CartAddress{
//					Address: "323 Washington Ave N, Minneapolis, MN 55401-2427",
//				},
//				LineItems: []models.CartLineItem{
//					{
//						ItemID:   "item-1",
//						Price:    10.75,
//						Quantity: 1.5,
//					},
//				},
//			},
//		},
//	}
type CalculateCartRequest struct {
	// Items is the array of cart items (must contain exactly 1 element)
	Items []CartItem `json:"items"`
}

// CartItem represents a single cart containing customer info, addresses, currency,
// and line items for tax calculation.
type CartItem struct {
	// CustomerID is the customer identifier (required)
	CustomerID string `json:"customerId"`

	// Currency is the currency information (must be USD)
	Currency CartCurrency `json:"currency"`

	// Destination is the destination address used for tax calculation (required)
	Destination CartAddress `json:"destination"`

	// Origin is the origin address of the seller/shipper (required)
	Origin CartAddress `json:"origin"`

	// LineItems is the array of line items in the cart (1-250 items)
	LineItems []CartLineItem `json:"lineItems"`
}

// CartAddress represents a simple address structure for cart tax calculation (single string format).
type CartAddress struct {
	// Address is the full address string for geocoding (required)
	Address string `json:"address"`
}

// CartCurrency represents currency information for the cart request.
type CartCurrency struct {
	// CurrencyCode is the ISO currency code (must be "USD")
	CurrencyCode string `json:"currencyCode"`
}

// CartLineItem represents a line item in the cart request with product details for tax calculation.
type CartLineItem struct {
	// ItemID is the unique identifier for the line item (required)
	ItemID string `json:"itemId"`

	// Price is the unit price of the item (must be positive, greater than 0)
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (must be positive, greater than 0, supports fractional values)
	Quantity float64 `json:"quantity"`

	// TaxabilityCode is the taxability code for product-specific tax rules (optional).
	// Passed to the v60 tax engine if provided (ZipTax) or mapped to TIC (TaxCloud).
	TaxabilityCode *int64 `json:"taxabilityCode,omitempty"`
}

// CalculateCartResponse represents the response from ZipTax cart tax calculation
// containing per-item tax details.
//
// This type is returned by Client.CalculateCart when TaxCloud credentials are NOT configured.
// It implements the CalculateCartResult interface.
type CalculateCartResponse struct {
	// Items is the array of cart results (mirrors request items array order)
	Items []CartItemResponse `json:"items"`
}

// isCalculateCartResult implements the CalculateCartResult interface for CalculateCartResponse.
func (*CalculateCartResponse) isCalculateCartResult() {}

// CartItemResponse represents a single cart response with calculated tax information per line item.
type CartItemResponse struct {
	// CartID is the server-generated UUID identifying this cart calculation
	CartID string `json:"cartId"`

	// CustomerID is the customer identifier (echoed from request)
	CustomerID string `json:"customerId"`

	// Destination is the destination address (echoed from request)
	Destination CartAddress `json:"destination"`

	// Origin is the origin address (echoed from request)
	Origin CartAddress `json:"origin"`

	// LineItems is the array of line items with calculated tax information
	LineItems []CartLineItemResponse `json:"lineItems"`
}

// CartLineItemResponse represents a line item in the cart response with calculated tax rate and amount.
type CartLineItemResponse struct {
	// ItemID is the unique identifier for the line item (echoed from request)
	ItemID string `json:"itemId"`

	// Price is the unit price of the item (echoed from request)
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (echoed from request)
	Quantity float64 `json:"quantity"`

	// Tax is the calculated tax information for this line item
	Tax CartTax `json:"tax"`
}

// CartTax represents calculated tax details for a cart line item.
type CartTax struct {
	// Rate is the calculated sales tax rate
	Rate float64 `json:"rate"`

	// Amount is the calculated tax amount: (price x quantity) x rate
	Amount float64 `json:"amount"`
}
