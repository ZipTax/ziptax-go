package ziptax

import (
	"context"

	"github.com/ziptax/ziptax-go/internal/validation"
	"github.com/ziptax/ziptax-go/models"
)

// Merchant Management lets platforms and SaaS businesses provision tax compliance
// for their own customers. You create a merchant for each seller on your platform,
// then manage that merchant's compliance through ZipTax.
//
// All merchant endpoints live on the ZipTax API and authenticate with the ZipTax
// API key supplied to NewClient. They do not require TaxCloud credentials on the
// client: a merchant's TaxCloud credentials are stored server-side with
// SetMerchantCredentials, or established by the TaxCloud invite flow.
//
// Merchant Management is a Private Preview feature. Contact support@zip.tax to
// gain access. Request and response shapes may change before general availability.

// postZipTax performs a POST against the ZipTax API with the client's API key.
//
// Post() takes an explicit base URL and headers because it was built for
// multi-API routing (ZipTax vs TaxCloud); unlike Get(), it does not auto-set
// the API key. This helper supplies both for the ZipTax API.
//
// The request follows the client's retry policy. Use postZipTaxOnce for
// operations that must not be repeated automatically.
func (c *Client) postZipTax(ctx context.Context, path string, body, result interface{}) error {
	return c.httpClient.Post(ctx, c.config.BaseURL, path, map[string]string{
		APIKeyHeader: c.config.APIKey,
	}, body, result)
}

// postZipTaxOnce performs a POST against the ZipTax API using exactly one
// attempt, ignoring the client's retry policy.
//
// It is used by the operations whose repetition creates an additional record
// server-side rather than converging on the same state, so an automatic retry
// after a timeout or 5xx could silently duplicate it: refunds, and the creates
// whose identifier the server generates (merchants and exemption certificates).
// Operations keyed by a caller-supplied identifier, such as order creation, are
// not in this group; resubmitting those conflicts rather than duplicating.
func (c *Client) postZipTaxOnce(ctx context.Context, path string, body, result interface{}) error {
	return c.httpClient.PostOnce(ctx, c.config.BaseURL, path, map[string]string{
		APIKeyHeader: c.config.APIKey,
	}, body, result)
}

// CreateMerchant creates a merchant under the calling account.
//
// The MerchantType field selects the compliance model and cannot be changed later.
// It defaults to models.MerchantTypeTaxCloud, which starts the TaxCloud invite
// process; use models.MerchantTypeSelfManaged for a merchant that is active
// immediately and handles its own registration, filing, and remittance.
//
// This call is never retried automatically, whatever WithMaxRetries is set to,
// because the server assigns the merchant ID: a resubmission would create a
// second merchant. If it fails without a clear outcome, call ListMerchants to
// check whether the merchant was created before trying again.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.CreateMerchant(ctx, &models.CreateMerchantRequest{
//		MerchantName: "Acme Supply Co",
//		ContactEmail: "ops@acme.example",
//		MerchantType: models.MerchantTypeTaxCloud,
//	})
//	if err != nil {
//		return fmt.Errorf("failed to create merchant: %w", err)
//	}
//	fmt.Printf("Created merchant %s\n", resp.MerchantID)
func (c *Client) CreateMerchant(ctx context.Context, request *models.CreateMerchantRequest) (*models.MerchantOperationResponse, error) {
	if request == nil {
		return nil, newNilRequestError("CreateMerchantRequest")
	}

	if err := validation.ValidateMerchantName(request.MerchantName); err != nil {
		return nil, &ValidationError{
			Field:   "merchantName",
			Value:   request.MerchantName,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateMerchantType(request.MerchantType); err != nil {
		return nil, &ValidationError{
			Field:   "merchant_type",
			Value:   request.MerchantType,
			Message: err.Error(),
		}
	}

	// Not retried: the server generates the merchantId, so a repeated submission
	// creates a second merchant rather than returning the first.
	var response models.MerchantOperationResponse
	if err := c.postZipTaxOnce(ctx, "/merchant/create", request, &response); err != nil {
		return nil, wrapError("failed to create merchant", err)
	}

	return &response, nil
}

// UpdateMerchant updates a merchant's name, contact details, or reference ID.
// The merchant must be owned by the calling account.
//
// MerchantName is required on the nested Update value even when it is unchanged.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.UpdateMerchant(ctx, &models.UpdateMerchantRequest{
//		MerchantID: "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		Update: models.MerchantUpdate{
//			MerchantName: "Acme Supply Co",
//			ContactEmail: "billing@acme.example",
//		},
//	})
//	if err != nil {
//		return fmt.Errorf("failed to update merchant: %w", err)
//	}
//	fmt.Println(resp.Message)
func (c *Client) UpdateMerchant(ctx context.Context, request *models.UpdateMerchantRequest) (*models.MerchantOperationResponse, error) {
	if request == nil {
		return nil, newNilRequestError("UpdateMerchantRequest")
	}

	if err := validation.ValidateMerchantID(request.MerchantID); err != nil {
		return nil, &ValidationError{
			Field:   "merchantId",
			Value:   request.MerchantID,
			Message: err.Error(),
		}
	}
	if err := validation.ValidateMerchantName(request.Update.MerchantName); err != nil {
		return nil, &ValidationError{
			Field:   "update.merchantName",
			Value:   request.Update.MerchantName,
			Message: err.Error(),
		}
	}

	var response models.MerchantOperationResponse
	if err := c.postZipTax(ctx, "/merchant/update", request, &response); err != nil {
		return nil, wrapError("failed to update merchant", err)
	}

	return &response, nil
}

// DeleteMerchant soft-deletes a merchant owned by the calling account.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.DeleteMerchant(ctx, "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e")
//	if err != nil {
//		return fmt.Errorf("failed to delete merchant: %w", err)
//	}
//	fmt.Println(resp.Status)
func (c *Client) DeleteMerchant(ctx context.Context, merchantID string) (*models.MerchantOperationResponse, error) {
	if err := validation.ValidateMerchantID(merchantID); err != nil {
		return nil, &ValidationError{
			Field:   "merchantId",
			Value:   merchantID,
			Message: err.Error(),
		}
	}

	var response models.MerchantOperationResponse
	request := &models.DeleteMerchantRequest{MerchantID: merchantID}
	if err := c.postZipTax(ctx, "/merchant/delete", request, &response); err != nil {
		return nil, wrapError("failed to delete merchant", err)
	}

	return &response, nil
}

// GetMerchant retrieves a single merchant owned by the calling account.
//
// Example:
//
//	ctx := context.Background()
//	merchant, err := client.GetMerchant(ctx, "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e")
//	if err != nil {
//		return fmt.Errorf("failed to get merchant: %w", err)
//	}
//	fmt.Printf("%s is %s\n", merchant.MerchantName, merchant.Status)
func (c *Client) GetMerchant(ctx context.Context, merchantID string) (*models.Merchant, error) {
	if err := validation.ValidateMerchantID(merchantID); err != nil {
		return nil, &ValidationError{
			Field:   "merchantId",
			Value:   merchantID,
			Message: err.Error(),
		}
	}

	var response models.Merchant
	request := &models.GetMerchantRequest{MerchantID: merchantID}
	if err := c.postZipTax(ctx, "/merchant/get", request, &response); err != nil {
		return nil, wrapError("failed to get merchant", err)
	}

	return &response, nil
}

// ListMerchants returns every merchant owned by the calling account.
//
// Example:
//
//	ctx := context.Background()
//	merchants, err := client.ListMerchants(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to list merchants: %w", err)
//	}
//	for _, m := range merchants {
//		fmt.Printf("%s: %s (%s)\n", m.MerchantID, m.MerchantName, m.Status)
//	}
func (c *Client) ListMerchants(ctx context.Context) ([]models.Merchant, error) {
	var response []models.Merchant
	if err := c.httpClient.Get(ctx, "/merchant/list", nil, &response); err != nil {
		return nil, wrapError("failed to list merchants", err)
	}

	return response, nil
}

// SetMerchantCredentials stores a merchant's TaxCloud connection ID and API key
// so that ZipTax can act on the merchant's behalf. The API stores them encrypted
// at rest.
//
// Use this when a merchant already has a TaxCloud account. Merchants created with
// an invite establish their credentials by accepting it instead.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.SetMerchantCredentials(ctx, &models.SetMerchantCredentialsRequest{
//		MerchantID:   "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
//		ConnectionID: "25eb9b97-5acb-492d-b720-c03e79cf715a",
//		APIKey:       taxCloudKey,
//	})
//	if err != nil {
//		return fmt.Errorf("failed to set merchant credentials: %w", err)
//	}
//	fmt.Println(resp.Status)
func (c *Client) SetMerchantCredentials(ctx context.Context, request *models.SetMerchantCredentialsRequest) (*models.MerchantOperationResponse, error) {
	if request == nil {
		return nil, newNilRequestError("SetMerchantCredentialsRequest")
	}

	if err := validation.ValidateMerchantCredentials(&validation.MerchantCredentialsInput{
		MerchantID:   request.MerchantID,
		ConnectionID: request.ConnectionID,
		APIKey:       request.APIKey,
	}); err != nil {
		// The API key is a secret, so the ValidationError carries only the
		// non-sensitive identifiers.
		return nil, &ValidationError{
			Field:   "SetMerchantCredentialsRequest",
			Value:   "merchantId=" + request.MerchantID,
			Message: err.Error(),
		}
	}

	var response models.MerchantOperationResponse
	if err := c.postZipTax(ctx, "/merchant/credentials/set", request, &response); err != nil {
		return nil, wrapError("failed to set merchant credentials", err)
	}

	return &response, nil
}

// DeleteMerchantCredentials removes a merchant's stored TaxCloud credentials.
// After this call the merchant's transaction endpoints are no longer available
// until credentials are set again.
//
// Example:
//
//	ctx := context.Background()
//	resp, err := client.DeleteMerchantCredentials(ctx, "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e")
//	if err != nil {
//		return fmt.Errorf("failed to delete merchant credentials: %w", err)
//	}
//	fmt.Println(resp.Status)
func (c *Client) DeleteMerchantCredentials(ctx context.Context, merchantID string) (*models.MerchantOperationResponse, error) {
	if err := validation.ValidateMerchantID(merchantID); err != nil {
		return nil, &ValidationError{
			Field:   "merchantId",
			Value:   merchantID,
			Message: err.Error(),
		}
	}

	var response models.MerchantOperationResponse
	request := &models.DeleteMerchantCredentialsRequest{MerchantID: merchantID}
	if err := c.postZipTax(ctx, "/merchant/credentials/delete", request, &response); err != nil {
		return nil, wrapError("failed to delete merchant credentials", err)
	}

	return &response, nil
}

// newNilRequestError builds the ValidationError returned when a caller passes a nil request.
func newNilRequestError(typeName string) *ValidationError {
	return &ValidationError{
		Field:   typeName,
		Value:   "",
		Message: "request cannot be nil",
	}
}
