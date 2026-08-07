package validation

import (
	"fmt"
)

// Limits enforced by the merchant endpoints, mirrored here so callers get a
// local error instead of a round trip that ends in a 400.
const (
	// MaxMerchantNameLength is the longest merchant name the API accepts.
	MaxMerchantNameLength = 255

	// MaxMerchantCarts is the most carts a single cart calculation may contain.
	MaxMerchantCarts = 100

	// MaxCertificateListLimit is the largest page size the certificate list accepts.
	MaxCertificateListLimit = 100

	// MaxReasonDescriptionLength is the longest exemption reason description the API accepts.
	MaxReasonDescriptionLength = 20
)

// ValidateMerchantID validates that a merchant identifier is present.
func ValidateMerchantID(merchantID string) error {
	if merchantID == "" {
		return fmt.Errorf("merchantId is required")
	}
	return nil
}

// ValidateMerchantName validates a merchant's legal or trading name.
func ValidateMerchantName(name string) error {
	if name == "" {
		return fmt.Errorf("merchantName is required")
	}
	if len(name) > MaxMerchantNameLength {
		return fmt.Errorf("merchantName must not exceed %d characters, got %d", MaxMerchantNameLength, len(name))
	}
	return nil
}

// ValidateMerchantType validates the compliance model selected at creation time.
// An empty value is allowed: the API defaults it to "taxcloud".
func ValidateMerchantType(merchantType string) error {
	switch merchantType {
	case "", "taxcloud", "self-managed":
		return nil
	default:
		return fmt.Errorf("merchant_type must be 'taxcloud' or 'self-managed', got %q", merchantType)
	}
}

// MerchantCredentialsInput holds the fields needed to validate a set-credentials request.
type MerchantCredentialsInput struct {
	MerchantID   string
	ConnectionID string
	APIKey       string
}

// ValidateMerchantCredentials validates a request to store a merchant's TaxCloud credentials.
func ValidateMerchantCredentials(input *MerchantCredentialsInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.ConnectionID == "" {
		return fmt.Errorf("connectionId is required")
	}
	if input.APIKey == "" {
		return fmt.Errorf("apiKey is required")
	}
	return nil
}

// MerchantCartInput holds the fields needed to validate a merchant cart calculation.
type MerchantCartInput struct {
	MerchantID string
	Carts      []MerchantCartItemInput
}

// MerchantCartItemInput holds the fields needed to validate a single cart.
type MerchantCartItemInput struct {
	CustomerID       string
	DestinationLine1 string
	DestinationCity  string
	DestinationState string
	DestinationZip   string
	OriginLine1      string
	OriginCity       string
	OriginState      string
	OriginZip        string
	LineItems        []MerchantCartLineItemInput
}

// MerchantCartLineItemInput holds the fields needed to validate a cart line item.
type MerchantCartLineItemInput struct {
	ItemID   string
	Price    float64
	Quantity float64
}

// ValidateMerchantCartRequest validates the input for a CalculateMerchantCart request.
func ValidateMerchantCartRequest(input *MerchantCartInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}

	cartCount := len(input.Carts)
	if cartCount == 0 {
		return fmt.Errorf("items must contain at least 1 cart")
	}
	if cartCount > MaxMerchantCarts {
		return fmt.Errorf("items must not exceed %d carts, got %d", MaxMerchantCarts, cartCount)
	}

	for i, cart := range input.Carts {
		if cart.CustomerID == "" {
			return fmt.Errorf("items[%d].customerId is required", i)
		}
		if err := validateStructuredAddress(fmt.Sprintf("items[%d].destination", i),
			cart.DestinationLine1, cart.DestinationCity, cart.DestinationState, cart.DestinationZip); err != nil {
			return err
		}
		if err := validateStructuredAddress(fmt.Sprintf("items[%d].origin", i),
			cart.OriginLine1, cart.OriginCity, cart.OriginState, cart.OriginZip); err != nil {
			return err
		}

		if len(cart.LineItems) == 0 {
			return fmt.Errorf("items[%d].lineItems must contain at least 1 item", i)
		}
		for j, li := range cart.LineItems {
			if li.ItemID == "" {
				return fmt.Errorf("items[%d].lineItems[%d].itemId is required", i, j)
			}
			if li.Price < 0 {
				return fmt.Errorf("items[%d].lineItems[%d].price must not be negative, got %v", i, j, li.Price)
			}
			if li.Quantity <= 0 {
				return fmt.Errorf("items[%d].lineItems[%d].quantity must be greater than 0, got %v", i, j, li.Quantity)
			}
		}
	}

	return nil
}

// validateStructuredAddress checks the fields the API marks required on a TaxCloud-style address.
func validateStructuredAddress(field, line1, city, state, zip string) error {
	if line1 == "" {
		return fmt.Errorf("%s.line1 is required", field)
	}
	if city == "" {
		return fmt.Errorf("%s.city is required", field)
	}
	if state == "" {
		return fmt.Errorf("%s.state is required", field)
	}
	if zip == "" {
		return fmt.Errorf("%s.zip is required", field)
	}
	return nil
}

// MerchantOrderInput holds the routing fields shared by the merchant order endpoints.
type MerchantOrderInput struct {
	MerchantID string
	OrderID    string
}

// ValidateMerchantOrderRequest validates the merchantId and orderId pair that the
// merchant order and refund endpoints route on.
func ValidateMerchantOrderRequest(input *MerchantOrderInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.OrderID == "" {
		return fmt.Errorf("orderId is required")
	}
	return nil
}

// MerchantOrderFromCartInput holds the fields needed to validate a create-order-from-cart request.
type MerchantOrderFromCartInput struct {
	MerchantID string
	CartID     string
	OrderID    string
}

// ValidateMerchantOrderFromCartRequest validates a request to capture a calculated cart as an order.
func ValidateMerchantOrderFromCartRequest(input *MerchantOrderFromCartInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.CartID == "" {
		return fmt.Errorf("cartId is required")
	}
	if input.OrderID == "" {
		return fmt.Errorf("orderId is required")
	}
	return nil
}

// MerchantCreateOrderInput holds the fields needed to validate a direct order creation.
type MerchantCreateOrderInput struct {
	MerchantID      string
	OrderID         string
	CustomerID      string
	TransactionDate string
	CompletedDate   string
	LineItems       []MerchantCartLineItemInput
}

// ValidateMerchantCreateOrderRequest validates a request to record an order directly.
func ValidateMerchantCreateOrderRequest(input *MerchantCreateOrderInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantOrderRequest(&MerchantOrderInput{
		MerchantID: input.MerchantID,
		OrderID:    input.OrderID,
	}); err != nil {
		return err
	}
	if input.CustomerID == "" {
		return fmt.Errorf("customerId is required")
	}
	if input.TransactionDate == "" {
		return fmt.Errorf("transactionDate is required")
	}
	if input.CompletedDate == "" {
		return fmt.Errorf("completedDate is required")
	}
	if len(input.LineItems) == 0 {
		return fmt.Errorf("lineItems must contain at least 1 item")
	}
	for i, li := range input.LineItems {
		if li.ItemID == "" {
			return fmt.Errorf("lineItems[%d].itemId is required", i)
		}
		if li.Price < 0 {
			return fmt.Errorf("lineItems[%d].price must not be negative, got %v", i, li.Price)
		}
		if li.Quantity <= 0 {
			return fmt.Errorf("lineItems[%d].quantity must be greater than 0, got %v", i, li.Quantity)
		}
	}
	return nil
}

// MerchantCertificateInput holds the fields needed to validate an exemption certificate creation.
type MerchantCertificateInput struct {
	MerchantID        string
	CustomerID        string
	CustomerName      string
	Reason            string
	ReasonDescription string
	BusinessType      string
	AddressLine1      string
	AddressCity       string
	AddressState      string
	AddressZip        string
	States            []string
}

// ValidateMerchantCertificateRequest validates a request to create an exemption certificate.
func ValidateMerchantCertificateRequest(input *MerchantCertificateInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.CustomerID == "" {
		return fmt.Errorf("customerId is required")
	}
	if input.CustomerName == "" {
		return fmt.Errorf("customerName is required")
	}
	if input.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	if input.ReasonDescription == "" {
		return fmt.Errorf("reasonDescription is required")
	}
	if len(input.ReasonDescription) > MaxReasonDescriptionLength {
		return fmt.Errorf("reasonDescription must not exceed %d characters, got %d",
			MaxReasonDescriptionLength, len(input.ReasonDescription))
	}
	if input.BusinessType == "" {
		return fmt.Errorf("customerBusinessType is required")
	}
	if err := validateStructuredAddress("address",
		input.AddressLine1, input.AddressCity, input.AddressState, input.AddressZip); err != nil {
		return err
	}
	if len(input.States) == 0 {
		return fmt.Errorf("states must contain at least 1 state")
	}
	for i, s := range input.States {
		if s == "" {
			return fmt.Errorf("states[%d].abbreviation is required", i)
		}
	}
	return nil
}

// MerchantCertificateRefInput holds the fields that identify a single certificate.
type MerchantCertificateRefInput struct {
	MerchantID    string
	CertificateID string
}

// ValidateMerchantCertificateRef validates the merchantId and certificateId pair
// used by the certificate get and delete endpoints.
func ValidateMerchantCertificateRef(input *MerchantCertificateRefInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.CertificateID == "" {
		return fmt.Errorf("certificateId is required")
	}
	return nil
}

// MerchantCertificateListInput holds the fields needed to validate a certificate list request.
type MerchantCertificateListInput struct {
	MerchantID string
	Limit      int64
	SortBy     string
}

// ValidateMerchantCertificateListRequest validates a request to page through
// a merchant's exemption certificates.
func ValidateMerchantCertificateListRequest(input *MerchantCertificateListInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}
	if err := ValidateMerchantID(input.MerchantID); err != nil {
		return err
	}
	if input.Limit < 0 {
		return fmt.Errorf("limit must not be negative, got %d", input.Limit)
	}
	if input.Limit > MaxCertificateListLimit {
		return fmt.Errorf("limit must not exceed %d, got %d", MaxCertificateListLimit, input.Limit)
	}
	switch input.SortBy {
	case "", "createdDate", "id":
	default:
		return fmt.Errorf("sortBy must be 'createdDate' or 'id', got %q", input.SortBy)
	}
	return nil
}
