package models

// Currency codes accepted on merchant carts and orders.
const (
	// CurrencyCodeUSD is the United States dollar.
	CurrencyCodeUSD = "USD"

	// CurrencyCodeCAD is the Canadian dollar.
	CurrencyCodeCAD = "CAD"
)

// Country codes accepted on merchant addresses.
const (
	// CountryCodeUS is the ISO 3166-1 alpha-2 code for the United States.
	CountryCodeUS = "US"

	// CountryCodeCA is the ISO 3166-1 alpha-2 code for Canada.
	CountryCodeCA = "CA"
)

// Discount types accepted on merchant carts and orders.
const (
	// DiscountTypePercentage expresses a discount as a decimal fraction, e.g. 0.1 for 10% off.
	DiscountTypePercentage = "percentage"

	// DiscountTypeAmount expresses a discount as a fixed currency amount.
	DiscountTypeAmount = "amount"
)

// MerchantCalculateCartRequest is the request payload for calculating sales tax
// on one or more carts on behalf of a merchant.
//
// The same request contract serves both compliance models. ZipTax routes on the
// merchant's model, so you never branch on merchant type when building a request.
// A self-managed merchant rejects, rather than ignores, fields it cannot honour:
// discounts, exemption, deliveredBySeller, productId, and any currency other
// than USD.
type MerchantCalculateCartRequest struct {
	// MerchantID is the UUID of the merchant to calculate on behalf of (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// Items is the array of carts to calculate tax for (required, 1-100 carts).
	// Most integrations send a single cart.
	Items []MerchantCart `json:"items"`

	// TransactionDate is the RFC3339 datetime the carts are calculated for (optional).
	// Defaults to the current time when omitted.
	TransactionDate string `json:"transactionDate,omitempty"`
}

// MerchantCart is a single cart submitted for tax calculation.
type MerchantCart struct {
	// CartID is your identifier for this cart (optional). If omitted, TaxCloud
	// generates one and returns it in the response. Pass it to CreateMerchantOrderFromCart
	// to capture the cart as an order.
	CartID string `json:"cartId,omitempty"`

	// CustomerID is your identifier for the customer (required).
	// Used to match exemption certificates and order history.
	CustomerID string `json:"customerId"`

	// Currency is the currency the line item prices are denominated in (required).
	// Defaults to USD when CurrencyCode is omitted.
	Currency Currency `json:"currency"`

	// Origin is the ship-from address (required).
	Origin TaxCloudAddress `json:"origin"`

	// Destination is the ship-to address (required).
	Destination TaxCloudAddress `json:"destination"`

	// LineItems is the array of line items in the cart (required, at least 1).
	LineItems []MerchantCartLineItem `json:"lineItems"`

	// DeliveredBySeller indicates the seller delivers the order directly rather
	// than via common carrier (optional). Affects taxability of delivery charges
	// in some states. Not supported for self-managed merchants unless false.
	DeliveredBySeller *bool `json:"deliveredBySeller,omitempty"`

	// Discounts holds line-item and order-level discounts (optional).
	// Not supported for self-managed merchants.
	Discounts *MerchantDiscounts `json:"discounts,omitempty"`

	// Exemption is the exemption information for the cart (optional).
	// Not supported for self-managed merchants unless it claims no exemption.
	Exemption *Exemption `json:"exemption,omitempty"`
}

// MerchantCartLineItem is a single line item submitted for tax calculation.
type MerchantCartLineItem struct {
	// Index is the zero-based position of the item within the cart (required).
	// Each line item must have a unique index.
	Index int64 `json:"index"`

	// ItemID is your unique identifier for the line item, e.g. a SKU (required).
	// Used to match line items in later order, refund, and discount operations.
	ItemID string `json:"itemId"`

	// Price is the unit price of the item in the cart's currency (required).
	// When discounts are provided this must be the pre-discount price.
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (required). Fractional quantities are allowed.
	Quantity float64 `json:"quantity"`

	// TIC is the Taxability Information Code classifying the product (optional).
	// Defaults to 0 (general tangible goods).
	//
	// The TIC vocabulary differs by compliance model. Self-managed carts use
	// ZipTax TICs, where 10001 is shipping and 11000 is handling; TaxCloud's
	// shipping TICs 11010-11015 and the Colorado retail delivery fee TIC 11098
	// are rejected for self-managed merchants.
	TIC *int64 `json:"tic,omitempty"`

	// ProductID is the product's ID in the merchant's TaxCloud product catalog (optional).
	// Must match an existing catalog product. Not supported for self-managed merchants.
	ProductID *string `json:"productId,omitempty"`
}

// MerchantDiscounts holds the discounts applied to a cart or order.
// Line-item discounts are applied before any order-level discount.
type MerchantDiscounts struct {
	// LineItemDiscounts are discounts applied to specific line items (optional).
	LineItemDiscounts []MerchantLineItemDiscount `json:"lineItemDiscounts,omitempty"`

	// OrderDiscount is a discount applied to the order total (optional).
	OrderDiscount *MerchantOrderDiscount `json:"orderDiscount,omitempty"`
}

// MerchantLineItemDiscount is a discount applied to a single line item.
type MerchantLineItemDiscount struct {
	// ItemID is the itemId of the line item this discount applies to (required).
	// Must match an itemId in the cart or order's line items.
	ItemID string `json:"itemId"`

	// Type is the kind of discount (required): DiscountTypePercentage or DiscountTypeAmount.
	Type string `json:"type"`

	// Value is the discount value (required): a decimal fraction between 0 and 1
	// for a percentage discount, or a currency amount for an amount discount.
	Value float64 `json:"value"`
}

// MerchantOrderDiscount is a discount applied to an order total.
type MerchantOrderDiscount struct {
	// Type is the kind of discount (required): DiscountTypePercentage or DiscountTypeAmount.
	Type string `json:"type"`

	// Value is the discount value (required): a decimal fraction between 0 and 1
	// for a percentage discount, or a currency amount for an amount discount.
	Value float64 `json:"value"`
}

// MerchantCalculateCartResponse is the response from a merchant cart calculation.
//
// The response shape depends on the merchant's compliance model. For a
// TaxCloud-connected merchant it is TaxCloud's response relayed verbatim. For a
// self-managed merchant it is the ZipTax calculation, which has no ConnectionID,
// Exemption, or DeliveredBySeller. Use IsSelfManaged to tell the two apart.
type MerchantCalculateCartResponse struct {
	// ConnectionID is the TaxCloud connection the calculation ran under.
	// Empty for a self-managed merchant, which has no TaxCloud connection.
	ConnectionID string `json:"connectionId,omitempty"`

	// TransactionDate is the RFC3339 datetime the carts were calculated for.
	TransactionDate string `json:"transactionDate,omitempty"`

	// Items holds one calculated cart per submitted cart, in the same order.
	Items []MerchantCartResult `json:"items"`
}

// IsSelfManaged reports whether the calculation was performed in-process by the
// ZipTax rate engine for a self-managed merchant, rather than by TaxCloud.
//
// A self-managed calculation is stateless: nothing is persisted, and the returned
// CartID cannot be captured as an order with CreateMerchantOrderFromCart.
func (r *MerchantCalculateCartResponse) IsSelfManaged() bool {
	return r.ConnectionID == ""
}

// MerchantCartResult is a single calculated cart.
type MerchantCartResult struct {
	// CartID identifies the calculated cart: the cartId you submitted, or a
	// generated one when you omitted it.
	CartID string `json:"cartId"`

	// CustomerID is your identifier for the customer, as submitted.
	CustomerID string `json:"customerId"`

	// Origin is the ship-from address, as submitted.
	Origin TaxCloudAddressResponse `json:"origin"`

	// Destination is the ship-to address, as submitted.
	Destination TaxCloudAddressResponse `json:"destination"`

	// Currency is the currency the prices and tax amounts are denominated in.
	Currency CurrencyResponse `json:"currency"`

	// LineItems holds the submitted line items, each with its calculated tax.
	LineItems []MerchantCartLineItemResult `json:"lineItems"`

	// DeliveredBySeller is whether the seller delivers the order directly, as submitted.
	// Absent for a self-managed merchant.
	DeliveredBySeller *bool `json:"deliveredBySeller,omitempty"`

	// Exemption is the exemption information applied to the calculation.
	// Absent for a self-managed merchant.
	Exemption *Exemption `json:"exemption,omitempty"`
}

// MerchantCartLineItemResult is a line item with its calculated tax.
type MerchantCartLineItemResult struct {
	// Index is the zero-based position of the item within the cart, as submitted.
	Index int64 `json:"index"`

	// ItemID is your unique identifier for the line item, as submitted.
	ItemID string `json:"itemId"`

	// Price is the unit price tax was calculated on. When discounts were applied,
	// this is the discounted unit price.
	Price float64 `json:"price"`

	// OriginalPrice is the original, pre-discount unit price, as submitted.
	OriginalPrice float64 `json:"originalPrice"`

	// Quantity is the quantity of the item.
	Quantity float64 `json:"quantity"`

	// Tax is the calculated rate and amount for this line item.
	Tax Tax `json:"tax"`

	// TIC is the Taxability Information Code the item was calculated under.
	// Nil when no TIC applies.
	TIC *int64 `json:"tic"`
}
