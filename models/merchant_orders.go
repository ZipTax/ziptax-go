package models

// Order kinds accepted when recording a merchant order.
const (
	// OrderKindOrder records a sale. This is the API default.
	OrderKindOrder = "order"

	// OrderKindCredit records a credit order.
	OrderKindCredit = "credit"
)

// ExpandRefunds asks GetMerchantOrder to include the order's refunds in the response.
const ExpandRefunds = "refunds"

// MerchantCreateOrderFromCartRequest is the request payload for capturing a cart
// previously calculated with CalculateMerchantCart as a recorded order.
//
// Available only for TaxCloud-connected merchants. A self-managed merchant has no
// TaxCloud connection to record the order in, so the request is refused with HTTP 403.
type MerchantCreateOrderFromCartRequest struct {
	// MerchantID is the UUID of the merchant (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// CartID is the cartId returned by, or supplied to, CalculateMerchantCart (required).
	CartID string `json:"cartId"`

	// OrderID is your identifier for the resulting order (required). Used later with
	// GetMerchantOrder, UpdateMerchantOrder, and CreateMerchantRefund.
	OrderID string `json:"orderId"`

	// Completed indicates the order has shipped, creating a tax liability (optional).
	// Defaults to false. Ignored when CompletedDate is provided.
	Completed *bool `json:"completed,omitempty"`

	// CompletedDate is the RFC3339 datetime the order shipped on (optional).
	// Takes precedence over Completed when provided.
	CompletedDate string `json:"completedDate,omitempty"`

	// Kind is the kind of order (optional): OrderKindOrder or OrderKindCredit.
	// Defaults to OrderKindOrder.
	Kind string `json:"kind,omitempty"`
}

// MerchantCreateOrderRequest is the request payload for recording an order directly,
// without a prior cart calculation. The tax amounts on each line item are the
// amounts your checkout collected.
//
// Available only for TaxCloud-connected merchants. A self-managed merchant has no
// TaxCloud connection to record the order in, so the request is refused with HTTP 403.
type MerchantCreateOrderRequest struct {
	// MerchantID is the UUID of the merchant (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// OrderID is your identifier for the order (required). Used later with
	// GetMerchantOrder, UpdateMerchantOrder, and CreateMerchantRefund.
	OrderID string `json:"orderId"`

	// CustomerID is your identifier for the customer (required).
	// Used to match exemption certificates and order history.
	CustomerID string `json:"customerId"`

	// TransactionDate is the RFC3339 datetime the order was purchased on (required).
	TransactionDate string `json:"transactionDate"`

	// CompletedDate is the RFC3339 datetime the order shipped on, which created
	// the tax liability (required).
	CompletedDate string `json:"completedDate"`

	// Currency is the currency the line item prices are denominated in (required).
	Currency Currency `json:"currency"`

	// Origin is the ship-from address (required).
	Origin TaxCloudAddress `json:"origin"`

	// Destination is the ship-to address (required).
	Destination TaxCloudAddress `json:"destination"`

	// LineItems is the array of line items with the tax your checkout collected (required).
	LineItems []MerchantOrderLineItem `json:"lineItems"`

	// BatchID groups this order with related orders (optional).
	BatchID string `json:"batchId,omitempty"`

	// Channel is the sales channel the order came from (optional). Pass one of
	// amazon, ebay, or walmart to exclude marketplace-collected tax from filing.
	Channel string `json:"channel,omitempty"`

	// DeliveredBySeller indicates the seller delivers the order directly rather
	// than via common carrier (optional).
	DeliveredBySeller *bool `json:"deliveredBySeller,omitempty"`

	// Discounts holds line-item and order-level discounts (optional).
	Discounts *MerchantDiscounts `json:"discounts,omitempty"`

	// ExcludeFromFiling excludes the order from tax filing (optional).
	ExcludeFromFiling *bool `json:"excludeFromFiling,omitempty"`

	// Exemption is the exemption information for the order (optional).
	Exemption *Exemption `json:"exemption,omitempty"`

	// Kind is the kind of order (optional): OrderKindOrder or OrderKindCredit.
	// Defaults to OrderKindOrder.
	Kind string `json:"kind,omitempty"`
}

// MerchantOrderLineItem is a line item on a directly recorded order, carrying the
// tax amount and rate your checkout collected.
type MerchantOrderLineItem struct {
	// Index is the zero-based position of the item within the order (required).
	// Each line item must have a unique index.
	Index int64 `json:"index"`

	// ItemID is your unique identifier for the line item (required).
	// Referenced by refunds and discounts.
	ItemID string `json:"itemId"`

	// Price is the unit price of the item that tax was calculated on (required).
	Price float64 `json:"price"`

	// Quantity is the quantity of the item (required). Fractional quantities are allowed.
	Quantity float64 `json:"quantity"`

	// Tax is the rate and amount collected for this line item (required).
	Tax Tax `json:"tax"`

	// TIC is the Taxability Information Code classifying the product (optional).
	// Defaults to 0 (general tangible goods).
	TIC *int64 `json:"tic,omitempty"`

	// ProductID is the product's ID in the merchant's TaxCloud product catalog (optional).
	ProductID *string `json:"productId,omitempty"`
}

// MerchantGetOrderRequest is the request payload for retrieving a recorded order.
type MerchantGetOrderRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// OrderID is your identifier for the order to retrieve (required),
	// as supplied when the order was created.
	OrderID string `json:"orderId"`

	// Expand set to ExpandRefunds includes the order's refunds in the response (optional).
	Expand string `json:"expand,omitempty"`
}

// MerchantUpdateOrderRequest is the request payload for modifying a recorded order.
// Currently the completedDate can be set, to mark the order shipped and create the
// tax liability.
//
// Updates overwrite the fields you send, so do not retry them blindly.
type MerchantUpdateOrderRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// OrderID is your identifier for the order to update (required),
	// as supplied when the order was created.
	OrderID string `json:"orderId"`

	// CompletedDate is the RFC3339 datetime the order shipped on, which creates
	// the tax liability.
	CompletedDate string `json:"completedDate,omitempty"`
}

// MerchantOrderResponse is a recorded order as returned by the merchant order endpoints.
type MerchantOrderResponse struct {
	// OrderID is your identifier for the order.
	OrderID string `json:"orderId"`

	// CustomerID is your identifier for the customer.
	CustomerID string `json:"customerId"`

	// ConnectionID is the TaxCloud connection the order was recorded under.
	ConnectionID string `json:"connectionId"`

	// TransactionDate is the RFC3339 datetime the order was purchased on.
	TransactionDate string `json:"transactionDate,omitempty"`

	// CompletedDate is the RFC3339 datetime the order shipped on, creating the tax
	// liability. Empty for orders that are not yet completed.
	CompletedDate string `json:"completedDate,omitempty"`

	// Origin is the ship-from address.
	Origin TaxCloudAddressResponse `json:"origin"`

	// Destination is the ship-to address.
	Destination TaxCloudAddressResponse `json:"destination"`

	// Currency is the currency the prices and tax amounts are denominated in.
	Currency CurrencyResponse `json:"currency"`

	// LineItems holds the order's line items, each with its tax rate and amount.
	LineItems []MerchantOrderLineItemResponse `json:"lineItems"`

	// Kind is the kind of order: OrderKindOrder or OrderKindCredit.
	Kind string `json:"kind"`

	// Channel is the sales channel the order came from. Nil when none was recorded.
	Channel *string `json:"channel"`

	// BatchID groups this order with related orders, if one was supplied.
	BatchID string `json:"batchId,omitempty"`

	// DeliveredBySeller is whether the seller delivered the order directly.
	DeliveredBySeller bool `json:"deliveredBySeller"`

	// ExcludeFromFiling is whether the order is excluded from tax filing.
	ExcludeFromFiling bool `json:"excludeFromFiling"`

	// Exemption is the exemption information applied to the order.
	Exemption *Exemption `json:"exemption,omitempty"`

	// Refunds holds the order's refunds. Populated on GetMerchantOrder when
	// Expand is set to ExpandRefunds.
	Refunds []MerchantRefundResponse `json:"refunds,omitempty"`
}

// MerchantOrderLineItemResponse is a line item on a recorded order.
type MerchantOrderLineItemResponse struct {
	// Index is the zero-based position of the item within the order.
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

	// Tax is the rate and amount calculated for this line item.
	Tax Tax `json:"tax"`

	// TIC is the Taxability Information Code the item was calculated under.
	// Nil when no TIC applies.
	TIC *int64 `json:"tic"`
}

// MerchantCreateRefundRequest is the request payload for refunding all or part of
// a recorded order.
//
// Refund prices and tax amounts are calculated automatically from the order; when
// the order had discounts, refunds use the discounted prices actually paid.
//
// Do not retry refunds blindly: a duplicate submission records a duplicate refund.
type MerchantCreateRefundRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// OrderID is your identifier for the order to refund (required),
	// as supplied when the order was created.
	OrderID string `json:"orderId"`

	// Items scopes a partial refund to specific line items and quantities (optional).
	// Omit or leave empty to refund the entire order.
	Items []MerchantRefundRequestItem `json:"items,omitempty"`

	// BatchID groups this refund with related refunds (optional).
	BatchID string `json:"batchId,omitempty"`

	// ReturnedDate should be included only if this return amends a previously filed
	// sales tax return (optional). Providing it triggers an Amended Sales Tax Return,
	// which is not typically recommended.
	ReturnedDate string `json:"returnedDate,omitempty"`
}

// MerchantRefundRequestItem scopes a partial refund to a line item and quantity.
type MerchantRefundRequestItem struct {
	// ItemID is the itemId of the line item to refund (required).
	// Must match an itemId from the original order.
	ItemID string `json:"itemId"`

	// Quantity is the quantity of the item to refund (required). May be fractional
	// and must not exceed the quantity on the original order.
	Quantity float64 `json:"quantity"`
}

// MerchantRefundResponse is a recorded refund.
type MerchantRefundResponse struct {
	// ConnectionID is the TaxCloud connection the refund was recorded under.
	ConnectionID string `json:"connectionId"`

	// CreatedDate is the RFC3339 datetime the refund was created.
	CreatedDate string `json:"createdDate,omitempty"`

	// Items holds the refunded line items, each with the refunded price, quantity,
	// and tax amount.
	Items []MerchantRefundItemResponse `json:"items"`

	// BatchID groups this refund with related refunds, if one was supplied.
	BatchID string `json:"batchId,omitempty"`

	// ReturnedDate is the RFC3339 datetime the refund took effect.
	ReturnedDate string `json:"returnedDate,omitempty"`
}

// MerchantRefundItemResponse is a refunded line item.
type MerchantRefundItemResponse struct {
	// Index is the zero-based position of the item within the refund.
	Index int64 `json:"index"`

	// ItemID is the itemId of the refunded line item, matching the original order.
	ItemID string `json:"itemId"`

	// Price is the unit price refunded, calculated automatically from the order.
	// When the order had discounts, this reflects the discounted amount actually paid.
	Price float64 `json:"price"`

	// Quantity is the quantity refunded.
	Quantity float64 `json:"quantity"`

	// Tax is the tax amount refunded for the item, calculated proportionally from
	// the order's tax.
	Tax RefundTax `json:"tax"`

	// TIC is the Taxability Information Code of the refunded item.
	TIC *int64 `json:"tic,omitempty"`
}
