package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCalculateCartRequest(t *testing.T) {
	validInput := func() *CartValidationInput {
		return &CartValidationInput{
			Items: []CartItemInput{
				{
					CustomerID:      "customer-453",
					CurrencyCode:    "USD",
					DestinationAddr: "200 Spectrum Center Dr, Irvine, CA 92618",
					OriginAddr:      "323 Washington Ave N, Minneapolis, MN 55401",
					LineItems: []CartLineItemInput{
						{
							ItemID:   "item-1",
							Price:    10.75,
							Quantity: 1.5,
						},
					},
				},
			},
		}
	}

	t.Run("valid request", func(t *testing.T) {
		err := ValidateCalculateCartRequest(validInput())
		require.NoError(t, err)
	})

	t.Run("valid request with multiple line items", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems = []CartLineItemInput{
			{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
			{ItemID: "item-2", Price: 5.00, Quantity: 2.0},
			{ItemID: "item-3", Price: 100.00, Quantity: 0.5},
		}
		err := ValidateCalculateCartRequest(input)
		require.NoError(t, err)
	})

	t.Run("nil input", func(t *testing.T) {
		err := ValidateCalculateCartRequest(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty items array", func(t *testing.T) {
		input := &CartValidationInput{Items: nil}
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 1 cart element")
		assert.Contains(t, err.Error(), "got 0")
	})

	t.Run("multiple cart items", func(t *testing.T) {
		input := validInput()
		input.Items = append(input.Items, input.Items[0])
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exactly 1 cart element")
		assert.Contains(t, err.Error(), "got 2")
	})

	t.Run("empty customer ID", func(t *testing.T) {
		input := validInput()
		input.Items[0].CustomerID = ""
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "customerId is required")
	})

	t.Run("invalid currency code", func(t *testing.T) {
		input := validInput()
		input.Items[0].CurrencyCode = "EUR"
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "currency.currencyCode must be 'USD'")
	})

	t.Run("empty destination address", func(t *testing.T) {
		input := validInput()
		input.Items[0].DestinationAddr = ""
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "destination.address is required")
	})

	t.Run("empty origin address", func(t *testing.T) {
		input := validInput()
		input.Items[0].OriginAddr = ""
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "origin.address is required")
	})

	t.Run("no line items", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems = nil
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least 1 item")
	})

	t.Run("too many line items", func(t *testing.T) {
		input := validInput()
		items := make([]CartLineItemInput, 251)
		for i := range items {
			items[i] = CartLineItemInput{
				ItemID:   fmt.Sprintf("item-%d", i),
				Price:    10.0,
				Quantity: 1.0,
			}
		}
		input.Items[0].LineItems = items
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not exceed 250")
	})

	t.Run("empty item ID", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems[0].ItemID = ""
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[0].itemId is required")
	})

	t.Run("zero price", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems[0].Price = 0
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[0].price must be greater than 0")
	})

	t.Run("negative price", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems[0].Price = -5.0
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[0].price must be greater than 0")
	})

	t.Run("zero quantity", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems[0].Quantity = 0
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[0].quantity must be greater than 0")
	})

	t.Run("negative quantity", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems[0].Quantity = -1.0
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[0].quantity must be greater than 0")
	})

	t.Run("second line item validation", func(t *testing.T) {
		input := validInput()
		input.Items[0].LineItems = append(input.Items[0].LineItems, CartLineItemInput{
			ItemID:   "",
			Price:    10.0,
			Quantity: 1.0,
		})
		err := ValidateCalculateCartRequest(input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lineItems[1].itemId is required")
	})
}
