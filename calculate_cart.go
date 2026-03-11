package ziptax

import (
	"context"
	"fmt"

	"github.com/ziptax/ziptax-go/internal/validation"
	"github.com/ziptax/ziptax-go/models"
)

// CalculateCartResult is the interface implemented by both *models.CalculateCartResponse and
// *models.TaxCloudCalculateCartResponse. The concrete type returned by CalculateCart depends
// on whether TaxCloud credentials are configured on the client.
//
// Use a type assertion to access the concrete response:
//
//	result, err := client.CalculateCart(ctx, req)
//	if err != nil {
//		return err
//	}
//	switch r := result.(type) {
//	case *models.CalculateCartResponse:
//		// ZipTax response
//	case *models.TaxCloudCalculateCartResponse:
//		// TaxCloud response
//	}
type CalculateCartResult = models.CalculateCartResult

// CalculateCart calculates sales tax for a shopping cart with one or more line items.
//
// This function supports two backends based on client configuration:
//   - ZipTax API (default): When the client is NOT configured with TaxCloud credentials,
//     the request is sent to the ZipTax /calculate/cart endpoint. Returns *models.CalculateCartResponse.
//   - TaxCloud API: When the client IS configured with TaxCloud credentials
//     (WithTaxCloudConnectionID and WithTaxCloudAPIKey), the request is automatically
//     routed to TaxCloud. The SDK transforms the request internally (parsing addresses,
//     mapping taxabilityCode to TIC, adding line item indices). Returns *models.TaxCloudCalculateCartResponse.
//
// The input contract (CalculateCartRequest) is the same regardless of which backend is used.
//
// Example (ZipTax):
//
//	ctx := context.Background()
//	req := &models.CalculateCartRequest{
//		Items: []models.CartItem{
//			{
//				CustomerID: "customer-453",
//				Currency:   models.CartCurrency{CurrencyCode: "USD"},
//				Destination: models.CartAddress{Address: "200 Spectrum Center Dr, Irvine, CA 92618"},
//				Origin:      models.CartAddress{Address: "323 Washington Ave N, Minneapolis, MN 55401"},
//				LineItems: []models.CartLineItem{
//					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
//				},
//			},
//		},
//	}
//	result, err := client.CalculateCart(ctx, req)
//	if err != nil {
//		return fmt.Errorf("failed to calculate cart: %w", err)
//	}
//	resp := result.(*models.CalculateCartResponse)
//	fmt.Printf("Tax rate: %.5f\n", resp.Items[0].LineItems[0].Tax.Rate)
//
// Example (TaxCloud):
//
//	// Client configured with TaxCloud credentials
//	result, err := client.CalculateCart(ctx, req)
//	if err != nil {
//		return fmt.Errorf("failed to calculate cart: %w", err)
//	}
//	resp := result.(*models.TaxCloudCalculateCartResponse)
//	fmt.Printf("Connection: %s, Tax: %.2f\n", resp.ConnectionID, resp.Items[0].LineItems[0].Tax.Amount)
func (c *Client) CalculateCart(ctx context.Context, request *models.CalculateCartRequest) (CalculateCartResult, error) {
	// Validate request
	if request == nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: "request cannot be nil",
		}
	}

	// Build validation input
	validationInput := &validation.CartValidationInput{
		Items: make([]validation.CartItemInput, len(request.Items)),
	}
	for i, item := range request.Items {
		lineItems := make([]validation.CartLineItemInput, len(item.LineItems))
		for j, li := range item.LineItems {
			lineItems[j] = validation.CartLineItemInput{
				ItemID:   li.ItemID,
				Price:    li.Price,
				Quantity: li.Quantity,
			}
		}
		validationInput.Items[i] = validation.CartItemInput{
			CustomerID:      item.CustomerID,
			CurrencyCode:    item.Currency.CurrencyCode,
			DestinationAddr: item.Destination.Address,
			OriginAddr:      item.Origin.Address,
			LineItems:       lineItems,
		}
	}

	if err := validation.ValidateCalculateCartRequest(validationInput); err != nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: err.Error(),
		}
	}

	// Route to TaxCloud if credentials are configured
	if c.config.HasTaxCloudCredentials() {
		return c.calculateCartTaxCloud(ctx, request)
	}

	// Default: route to ZipTax
	return c.calculateCartZipTax(ctx, request)
}

// calculateCartZipTax sends the cart calculation request to the ZipTax API.
func (c *Client) calculateCartZipTax(ctx context.Context, request *models.CalculateCartRequest) (*models.CalculateCartResponse, error) {
	var response models.CalculateCartResponse
	if err := c.httpClient.Post(ctx, c.config.BaseURL, "/calculate/cart", map[string]string{
		"X-API-Key": c.config.APIKey,
	}, request, &response); err != nil {
		return nil, wrapError("failed to calculate cart", err)
	}

	return &response, nil
}

// calculateCartTaxCloud transforms the request and sends it to the TaxCloud API.
func (c *Client) calculateCartTaxCloud(ctx context.Context, request *models.CalculateCartRequest) (*models.TaxCloudCalculateCartResponse, error) {
	// Transform the request from ZipTax format to TaxCloud format
	tcRequest, err := transformToTaxCloudCartRequest(request)
	if err != nil {
		return nil, &ValidationError{
			Field:   "request",
			Value:   "",
			Message: fmt.Sprintf("failed to transform request for TaxCloud: %s", err.Error()),
		}
	}

	// Build the path with connection ID
	path := fmt.Sprintf("/tax/connections/%s/carts", c.config.TaxCloudConnectionID)

	// Set up authentication headers
	headers := map[string]string{
		"X-API-Key": c.config.TaxCloudAPIKey,
	}

	// Make the POST request to TaxCloud API
	var response models.TaxCloudCalculateCartResponse
	if err := c.httpClient.Post(ctx, c.config.TaxCloudBaseURL, path, headers, tcRequest, &response); err != nil {
		return nil, wrapError("failed to calculate cart via TaxCloud", err)
	}

	return &response, nil
}

// transformToTaxCloudCartRequest transforms a CalculateCartRequest into the TaxCloud request format.
// This involves:
//   - Parsing single-string addresses into structured TaxCloudAddress components
//   - Mapping taxabilityCode to TIC (defaulting to 0 if nil)
//   - Adding 0-based index to each line item
func transformToTaxCloudCartRequest(request *models.CalculateCartRequest) (*models.TaxCloudCalculateCartRequest, error) {
	tcItems := make([]models.TaxCloudCartItem, len(request.Items))

	for i, item := range request.Items {
		// Parse destination address
		destAddr, err := validation.ParseAddress(item.Destination.Address)
		if err != nil {
			return nil, fmt.Errorf("destination address: %w", err)
		}

		// Parse origin address
		originAddr, err := validation.ParseAddress(item.Origin.Address)
		if err != nil {
			return nil, fmt.Errorf("origin address: %w", err)
		}

		// Transform line items
		tcLineItems := make([]models.TaxCloudCartLineItem, len(item.LineItems))
		for j, li := range item.LineItems {
			tic := int64(0)
			if li.TaxabilityCode != nil {
				tic = *li.TaxabilityCode
			}
			tcLineItems[j] = models.TaxCloudCartLineItem{
				Index:    int64(j),
				ItemID:   li.ItemID,
				Price:    li.Price,
				Quantity: li.Quantity,
				TIC:      tic,
			}
		}

		// Build destination TaxCloudAddress
		destTCAddr := models.TaxCloudAddress{
			Line1:       destAddr.Line1,
			City:        destAddr.City,
			State:       destAddr.State,
			Zip:         destAddr.Zip,
			CountryCode: &destAddr.CountryCode,
		}
		if destAddr.Line2 != "" {
			destTCAddr.Line2 = &destAddr.Line2
		}

		// Build origin TaxCloudAddress
		originTCAddr := models.TaxCloudAddress{
			Line1:       originAddr.Line1,
			City:        originAddr.City,
			State:       originAddr.State,
			Zip:         originAddr.Zip,
			CountryCode: &originAddr.CountryCode,
		}
		if originAddr.Line2 != "" {
			originTCAddr.Line2 = &originAddr.Line2
		}

		tcItems[i] = models.TaxCloudCartItem{
			CustomerID:  item.CustomerID,
			Currency:    item.Currency,
			Destination: destTCAddr,
			Origin:      originTCAddr,
			LineItems:   tcLineItems,
		}
	}

	return &models.TaxCloudCalculateCartRequest{
		Items: tcItems,
	}, nil
}
