package validation

import (
	"fmt"
)

// CartValidationInput holds the fields needed to validate a CalculateCartRequest.
// This avoids importing the models package into the validation package.
type CartValidationInput struct {
	Items []CartItemInput
}

// CartItemInput holds the fields needed to validate a single cart item.
type CartItemInput struct {
	CustomerID      string
	CurrencyCode    string
	DestinationAddr string
	OriginAddr      string
	LineItems       []CartLineItemInput
}

// CartLineItemInput holds the fields needed to validate a cart line item.
type CartLineItemInput struct {
	ItemID   string
	Price    float64
	Quantity float64
}

// ValidateCalculateCartRequest validates the input for a CalculateCart request.
func ValidateCalculateCartRequest(input *CartValidationInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}

	if len(input.Items) != 1 {
		return fmt.Errorf("items array must contain exactly 1 cart element, got %d", len(input.Items))
	}

	item := input.Items[0]

	if item.CustomerID == "" {
		return fmt.Errorf("customerId is required")
	}

	if item.CurrencyCode != "USD" {
		return fmt.Errorf("currency.currencyCode must be 'USD', got %q", item.CurrencyCode)
	}

	if item.DestinationAddr == "" {
		return fmt.Errorf("destination.address is required")
	}

	if item.OriginAddr == "" {
		return fmt.Errorf("origin.address is required")
	}

	lineItemCount := len(item.LineItems)
	if lineItemCount == 0 {
		return fmt.Errorf("lineItems must contain at least 1 item")
	}
	if lineItemCount > 250 {
		return fmt.Errorf("lineItems must not exceed 250 items, got %d", lineItemCount)
	}

	for i, li := range item.LineItems {
		if li.ItemID == "" {
			return fmt.Errorf("lineItems[%d].itemId is required", i)
		}
		if li.Price <= 0 {
			return fmt.Errorf("lineItems[%d].price must be greater than 0, got %v", i, li.Price)
		}
		if li.Quantity <= 0 {
			return fmt.Errorf("lineItems[%d].quantity must be greater than 0, got %v", i, li.Quantity)
		}
	}

	return nil
}

// CreateOrderFromCartValidationInput holds the fields needed to validate a CreateOrderFromCart request.
// This avoids importing the models package into the validation package.
type CreateOrderFromCartValidationInput struct {
	CartID  string
	OrderID string
}

// ValidateCreateOrderFromCartRequest validates the input for a CreateOrderFromCart request.
func ValidateCreateOrderFromCartRequest(input *CreateOrderFromCartValidationInput) error {
	if input == nil {
		return fmt.Errorf("validation input cannot be nil")
	}

	if input.CartID == "" {
		return fmt.Errorf("cartId is required")
	}

	if input.OrderID == "" {
		return fmt.Errorf("orderId is required")
	}

	return nil
}
