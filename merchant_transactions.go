package ziptax

import (
	"context"

	"github.com/ziptax/ziptax-go/internal/validation"
	"github.com/ziptax/ziptax-go/models"
)

// Merchant Transactions calculate cart tax, record orders, manage exemption
// certificates, and issue refunds on behalf of a merchant.
//
// Only CalculateMerchantCart serves both compliance models. Every other function
// in this file requires a TaxCloud-connected merchant: a self-managed merchant has
// no TaxCloud connection to store or read the state in, so the API refuses those
// requests with HTTP 403.
//
// Merchant Management is a Private Preview feature. Contact support@zip.tax to
// gain access. Request and response shapes may change before general availability.

// CalculateMerchantCart calculates sales tax for one or more carts on behalf of a merchant.
//
// The request contract is the same for both compliance models, so a caller does
// not have to know which mode a merchant is in. For a TaxCloud-connected merchant
// the cart is forwarded to TaxCloud, and the returned CartID can be captured as an
// order with CreateMerchantOrderFromCart. For a self-managed merchant the cart is
// calculated by the ZipTax rate engine, US destinations only, and nothing is
// persisted, so the returned CartID cannot be captured as an order. Check
// MerchantCalculateCartResponse.IsSelfManaged to tell the two apart.
//
// Calculation has no lasting side effect in either mode and is safe to retry.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.CalculateMerchantCart(ctx, &models.MerchantCalculateCartRequest{
//		MerchantID: "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		Items: []models.MerchantCart{
//			{
//				CartID:     "my-cart-1",
//				CustomerID: "customer-453",
//				Currency:   models.Currency{},
//				Origin: models.TaxCloudAddress{
//					Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
//				},
//				Destination: models.TaxCloudAddress{
//					Line1: "200 Spectrum Center Dr", City: "Irvine", State: "CA", Zip: "92618",
//				},
//				LineItems: []models.MerchantCartLineItem{
//					{Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5},
//				},
//			},
//		},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to calculate merchant cart: %w", err)
//	}
//	fmt.Printf("Tax: %.2f\n", resp.Items[0].LineItems[0].Tax.Amount)
func (c *Client) CalculateMerchantCart(ctx context.Context, request *models.MerchantCalculateCartRequest) (*models.MerchantCalculateCartResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantCalculateCartRequest")
	}

	carts := make([]validation.MerchantCartItemInput, 0, len(request.Items))
	for _, item := range request.Items {
		lineItems := make([]validation.MerchantCartLineItemInput, 0, len(item.LineItems))
		for _, li := range item.LineItems {
			lineItems = append(lineItems, validation.MerchantCartLineItemInput{
				Index:    li.Index,
				ItemID:   li.ItemID,
				Price:    li.Price,
				Quantity: li.Quantity,
			})
		}
		carts = append(carts, validation.MerchantCartItemInput{
			CustomerID:       item.CustomerID,
			DestinationLine1: item.Destination.Line1,
			DestinationCity:  item.Destination.City,
			DestinationState: item.Destination.State,
			DestinationZip:   item.Destination.Zip,
			OriginLine1:      item.Origin.Line1,
			OriginCity:       item.Origin.City,
			OriginState:      item.Origin.State,
			OriginZip:        item.Origin.Zip,
			LineItems:        lineItems,
		})
	}

	if err := validation.ValidateMerchantCartRequest(&validation.MerchantCartInput{
		MerchantID: request.MerchantID,
		Carts:      carts,
	}); err != nil {
		return nil, newValidationError("MerchantCalculateCartRequest", "merchantId="+request.MerchantID, err)
	}

	var response models.MerchantCalculateCartResponse
	if err := c.postZipTax(ctx, "/merchant/cart/calculate", request, &response); err != nil {
		return nil, wrapError("failed to calculate merchant cart", err)
	}

	return &response, nil
}

// CreateMerchantOrderFromCart captures a cart previously calculated with
// CalculateMerchantCart as a recorded order.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	order, err := client.CreateMerchantOrderFromCart(ctx, &models.MerchantCreateOrderFromCartRequest{
//		MerchantID: "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		CartID:     "my-cart-1",
//		OrderID:    "my-order-1",
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create order from cart: %w", err)
//	}
//	fmt.Printf("Recorded order %s\n", order.OrderID)
func (c *Client) CreateMerchantOrderFromCart(ctx context.Context, request *models.MerchantCreateOrderFromCartRequest) (*models.MerchantOrderResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantCreateOrderFromCartRequest")
	}

	if err := validation.ValidateMerchantOrderFromCartRequest(&validation.MerchantOrderFromCartInput{
		MerchantID: request.MerchantID,
		CartID:     request.CartID,
		OrderID:    request.OrderID,
	}); err != nil {
		return nil, newValidationError("MerchantCreateOrderFromCartRequest", "cartId="+request.CartID, err)
	}

	var response models.MerchantOrderResponse
	if err := c.postZipTax(ctx, "/merchant/order/create-from-cart", request, &response); err != nil {
		return nil, wrapError("failed to create merchant order from cart", err)
	}

	return &response, nil
}

// CreateMerchantOrder records an order directly, without a prior cart calculation.
// The tax amounts on each line item are the amounts your checkout collected.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	order, err := client.CreateMerchantOrder(ctx, &models.MerchantCreateOrderRequest{
//		MerchantID:      "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		OrderID:         "my-order-1",
//		CustomerID:      "customer-453",
//		TransactionDate: "2026-08-01T14:00:00Z",
//		CompletedDate:   "2026-08-02T09:15:00Z",
//		Currency:        models.Currency{},
//		Origin: models.TaxCloudAddress{
//			Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
//		},
//		Destination: models.TaxCloudAddress{
//			Line1: "200 Spectrum Center Dr", City: "Irvine", State: "CA", Zip: "92618",
//		},
//		LineItems: []models.MerchantOrderLineItem{
//			{
//				Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5,
//				Tax: models.Tax{Amount: 1.31, Rate: 0.0813},
//			},
//		},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create merchant order: %w", err)
//	}
//	fmt.Printf("Recorded order %s\n", order.OrderID)
func (c *Client) CreateMerchantOrder(ctx context.Context, request *models.MerchantCreateOrderRequest) (*models.MerchantOrderResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantCreateOrderRequest")
	}

	lineItems := make([]validation.MerchantCartLineItemInput, 0, len(request.LineItems))
	for _, li := range request.LineItems {
		lineItems = append(lineItems, validation.MerchantCartLineItemInput{
			Index:    li.Index,
			ItemID:   li.ItemID,
			Price:    li.Price,
			Quantity: li.Quantity,
		})
	}

	if err := validation.ValidateMerchantCreateOrderRequest(&validation.MerchantCreateOrderInput{
		MerchantID:      request.MerchantID,
		OrderID:         request.OrderID,
		CustomerID:      request.CustomerID,
		TransactionDate: request.TransactionDate,
		CompletedDate:   request.CompletedDate,
		LineItems:       lineItems,
	}); err != nil {
		return nil, newValidationError("MerchantCreateOrderRequest", "orderId="+request.OrderID, err)
	}

	var response models.MerchantOrderResponse
	if err := c.postZipTax(ctx, "/merchant/order/create", request, &response); err != nil {
		return nil, wrapError("failed to create merchant order", err)
	}

	return &response, nil
}

// GetMerchantOrder retrieves a recorded order. This is a read and is safe to retry.
//
// Set Expand to models.ExpandRefunds to include the order's refunds in the response.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	order, err := client.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
//		MerchantID: "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		OrderID:    "my-order-1",
//		Expand:     models.ExpandRefunds,
//	})
//	if err != nil {
//		return fmt.Errorf("failed to get merchant order: %w", err)
//	}
//	fmt.Printf("Order %s has %d refunds\n", order.OrderID, len(order.Refunds))
func (c *Client) GetMerchantOrder(ctx context.Context, request *models.MerchantGetOrderRequest) (*models.MerchantOrderResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantGetOrderRequest")
	}

	if err := validation.ValidateMerchantOrderRequest(&validation.MerchantOrderInput{
		MerchantID: request.MerchantID,
		OrderID:    request.OrderID,
	}); err != nil {
		return nil, newValidationError("MerchantGetOrderRequest", "orderId="+request.OrderID, err)
	}

	var response models.MerchantOrderResponse
	if err := c.postZipTax(ctx, "/merchant/order/get", request, &response); err != nil {
		return nil, wrapError("failed to get merchant order", err)
	}

	return &response, nil
}

// UpdateMerchantOrder modifies a recorded order. Currently the completed date can
// be set, marking the order shipped and creating the tax liability.
//
// Updates overwrite the fields you send, so do not retry them blindly.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	order, err := client.UpdateMerchantOrder(ctx, &models.MerchantUpdateOrderRequest{
//		MerchantID:    "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		OrderID:       "my-order-1",
//		CompletedDate: "2026-08-03T10:00:00Z",
//	})
//	if err != nil {
//		return fmt.Errorf("failed to update merchant order: %w", err)
//	}
//	fmt.Printf("Order completed on %s\n", order.CompletedDate)
func (c *Client) UpdateMerchantOrder(ctx context.Context, request *models.MerchantUpdateOrderRequest) (*models.MerchantOrderResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantUpdateOrderRequest")
	}

	if err := validation.ValidateMerchantOrderRequest(&validation.MerchantOrderInput{
		MerchantID: request.MerchantID,
		OrderID:    request.OrderID,
	}); err != nil {
		return nil, newValidationError("MerchantUpdateOrderRequest", "orderId="+request.OrderID, err)
	}

	var response models.MerchantOrderResponse
	if err := c.postZipTax(ctx, "/merchant/order/update", request, &response); err != nil {
		return nil, wrapError("failed to update merchant order", err)
	}

	return &response, nil
}

// CreateMerchantRefund refunds all or part of a recorded order. Omit Items to
// refund the entire order.
//
// Refund prices and tax amounts are calculated automatically from the order; when
// the order had discounts, refunds use the discounted prices actually paid.
//
// A duplicate submission records a duplicate refund, so this call is never
// retried automatically, whatever WithMaxRetries is set to. A transport error
// or 5xx is surfaced instead, leaving the outcome genuinely unknown rather than
// silently doubled. Before resubmitting, read the order back with
// GetMerchantOrder using models.ExpandRefunds to see whether the refund landed.
//
// Requires a TaxCloud-connected merchant.
//
// Example (partial refund):
//
//	ctx := context.Background()
//	refund, err := client.CreateMerchantRefund(ctx, &models.MerchantCreateRefundRequest{
//		MerchantID: "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		OrderID:    "my-order-1",
//		Items: []models.MerchantRefundRequestItem{
//			{ItemID: "item-1", Quantity: 1},
//		},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create merchant refund: %w", err)
//	}
//	fmt.Printf("Refunded %d items\n", len(refund.Items))
func (c *Client) CreateMerchantRefund(ctx context.Context, request *models.MerchantCreateRefundRequest) (*models.MerchantRefundResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantCreateRefundRequest")
	}

	if err := validation.ValidateMerchantOrderRequest(&validation.MerchantOrderInput{
		MerchantID: request.MerchantID,
		OrderID:    request.OrderID,
	}); err != nil {
		return nil, newValidationError("MerchantCreateRefundRequest", "orderId="+request.OrderID, err)
	}

	// Not retried: a repeated submission records a second refund against the
	// same order, which the caller cannot detect from the returned error.
	var response models.MerchantRefundResponse
	if err := c.postZipTaxOnce(ctx, "/merchant/refund/create", request, &response); err != nil {
		return nil, wrapError("failed to create merchant refund", err)
	}

	return &response, nil
}

// CreateMerchantCertificate creates an exemption certificate for one of a
// merchant's customers.
//
// Reference the returned CertificateID as the ExemptionID on carts and orders to
// apply the exemption.
//
// This call is never retried automatically, whatever WithMaxRetries is set to,
// because the server assigns the certificate ID: a resubmission would create a
// second certificate for the same customer. If it fails without a clear
// outcome, use ListMerchantCertificates filtered by CustomerID to check before
// trying again.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	cert, err := client.CreateMerchantCertificate(ctx, &models.MerchantCreateCertificateRequest{
//		MerchantID:           "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		CustomerID:           "customer-453",
//		CustomerName:         "Acme Reseller LLC",
//		CustomerBusinessType: models.BusinessTypeRetailTrade,
//		Reason:               models.ExemptionReasonResale,
//		ReasonDescription:    "Resale",
//		Address: models.TaxCloudAddress{
//			Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
//		},
//		States: []models.CertificateState{{Abbreviation: "MN"}},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create exemption certificate: %w", err)
//	}
//	fmt.Printf("Certificate %s\n", cert.CertificateID)
func (c *Client) CreateMerchantCertificate(ctx context.Context, request *models.MerchantCreateCertificateRequest) (*models.MerchantCertificate, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantCreateCertificateRequest")
	}

	states := make([]string, 0, len(request.States))
	for _, s := range request.States {
		states = append(states, s.Abbreviation)
	}

	if err := validation.ValidateMerchantCertificateRequest(&validation.MerchantCertificateInput{
		MerchantID:        request.MerchantID,
		CustomerID:        request.CustomerID,
		CustomerName:      request.CustomerName,
		Reason:            request.Reason,
		ReasonDescription: request.ReasonDescription,
		BusinessType:      request.CustomerBusinessType,
		AddressLine1:      request.Address.Line1,
		AddressCity:       request.Address.City,
		AddressState:      request.Address.State,
		AddressZip:        request.Address.Zip,
		States:            states,
	}); err != nil {
		return nil, newValidationError("MerchantCreateCertificateRequest", "customerId="+request.CustomerID, err)
	}

	// Not retried: the server generates the certificateId, so a repeated
	// submission creates a second certificate for the same customer.
	var response models.MerchantCertificate
	if err := c.postZipTaxOnce(ctx, "/merchant/cert/create", request, &response); err != nil {
		return nil, wrapError("failed to create merchant exemption certificate", err)
	}

	return &response, nil
}

// GetMerchantCertificate retrieves a single exemption certificate.
// This is a read and is safe to retry.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	cert, err := client.GetMerchantCertificate(ctx, merchantID, certificateID)
//	if err != nil {
//		return fmt.Errorf("failed to get exemption certificate: %w", err)
//	}
//	fmt.Printf("%s active: %t\n", cert.CustomerName, cert.IsActive())
func (c *Client) GetMerchantCertificate(ctx context.Context, merchantID, certificateID string) (*models.MerchantCertificate, error) {
	if err := validation.ValidateMerchantCertificateRef(&validation.MerchantCertificateRefInput{
		MerchantID:    merchantID,
		CertificateID: certificateID,
	}); err != nil {
		return nil, newValidationError("MerchantGetCertificateRequest", "certificateId="+certificateID, err)
	}

	request := &models.MerchantGetCertificateRequest{
		MerchantID:    merchantID,
		CertificateID: certificateID,
	}

	var response models.MerchantCertificate
	if err := c.postZipTax(ctx, "/merchant/cert/get", request, &response); err != nil {
		return nil, wrapError("failed to get merchant exemption certificate", err)
	}

	return &response, nil
}

// ListMerchantCertificates pages through a merchant's exemption certificates.
// This is a read and is safe to retry.
//
// Pass the NextCursor from a response as the Cursor of the next request to fetch
// the following page; NextCursor is empty when there are no further results.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	req := &models.MerchantListCertificatesRequest{MerchantID: merchantID, Limit: 50}
//	for {
//		page, err := client.ListMerchantCertificates(ctx, req)
//		if err != nil {
//			return fmt.Errorf("failed to list exemption certificates: %w", err)
//		}
//		for _, cert := range page.Items {
//			fmt.Println(cert.CertificateID, cert.CustomerName)
//		}
//		if page.NextCursor == "" {
//			break
//		}
//		req.Cursor = page.NextCursor
//	}
func (c *Client) ListMerchantCertificates(ctx context.Context, request *models.MerchantListCertificatesRequest) (*models.MerchantListCertificatesResponse, error) {
	if request == nil {
		return nil, newNilRequestError("MerchantListCertificatesRequest")
	}

	if err := validation.ValidateMerchantCertificateListRequest(&validation.MerchantCertificateListInput{
		MerchantID: request.MerchantID,
		Limit:      request.Limit,
		SortBy:     request.SortBy,
	}); err != nil {
		return nil, newValidationError("MerchantListCertificatesRequest", "merchantId="+request.MerchantID, err)
	}

	var response models.MerchantListCertificatesResponse
	if err := c.postZipTax(ctx, "/merchant/cert/list", request, &response); err != nil {
		return nil, wrapError("failed to list merchant exemption certificates", err)
	}

	return &response, nil
}

// DeleteMerchantCertificate deletes (disables) an exemption certificate so it can
// no longer be applied to new transactions.
//
// The API relays TaxCloud's response verbatim and its shape is not fixed, so the
// decoded body is returned as a map. Most callers only need the error.
//
// Requires a TaxCloud-connected merchant.
//
// Example:
//
//	ctx := context.Background()
//	if _, err := client.DeleteMerchantCertificate(ctx, merchantID, certificateID); err != nil {
//		return fmt.Errorf("failed to delete exemption certificate: %w", err)
//	}
func (c *Client) DeleteMerchantCertificate(ctx context.Context, merchantID, certificateID string) (map[string]interface{}, error) {
	if err := validation.ValidateMerchantCertificateRef(&validation.MerchantCertificateRefInput{
		MerchantID:    merchantID,
		CertificateID: certificateID,
	}); err != nil {
		return nil, newValidationError("MerchantDeleteCertificateRequest", "certificateId="+certificateID, err)
	}

	request := &models.MerchantDeleteCertificateRequest{
		MerchantID:    merchantID,
		CertificateID: certificateID,
	}

	var response map[string]interface{}
	if err := c.postZipTax(ctx, "/merchant/cert/delete", request, &response); err != nil {
		return nil, wrapError("failed to delete merchant exemption certificate", err)
	}

	return response, nil
}

// newValidationError builds a ValidationError from a validation package error.
func newValidationError(field, value string, err error) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: err.Error(),
	}
}
