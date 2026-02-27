package validation

import (
	"fmt"
)

// CartValidationInput holds the fields needed to validate a CalculateCartRequest.
// This avoids importing the models package into the validation package.
type CartValidationInput struct {
	ItemCount int
	Items     []CartItemInput
}

// CartItemInput holds the fields needed to validate a single cart item.
type CartItemInput struct {
	CustomerID      string
	CurrencyCode    string
	DestinationAddr string
	OriginAddr      string
	LineItemCount   int
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
	if input.ItemCount == 0 {
		return fmt.Errorf("items array must contain exactly 1 cart element, got 0")
	}
	if input.ItemCount != 1 {
		return fmt.Errorf("items array must contain exactly 1 cart element, got %d", input.ItemCount)
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

	if item.LineItemCount == 0 {
		return fmt.Errorf("lineItems must contain at least 1 item")
	}
	if item.LineItemCount > 250 {
		return fmt.Errorf("lineItems must not exceed 250 items, got %d", item.LineItemCount)
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
