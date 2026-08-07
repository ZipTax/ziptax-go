package models

// Merchant Management lets platforms provision tax compliance for their own
// customers. Every merchant uses one of two compliance models, chosen once at
// creation time via the MerchantType field on CreateMerchantRequest.
//
// Merchant Management is a Private Preview feature. Contact support@zip.tax to
// gain access. Request and response shapes may change before general availability.

// MerchantType selects the compliance model a merchant operates under.
// It is set once, at creation time, and cannot be changed afterwards.
const (
	// MerchantTypeTaxCloud creates a TaxCloud-connected merchant (the API default).
	// Creation kicks off the TaxCloud invite process: the merchant sets up their own
	// TaxCloud account and connects it, and TaxCloud handles registration, filing,
	// and remittance in the merchant's own account.
	MerchantTypeTaxCloud = "taxcloud"

	// MerchantTypeSelfManaged creates a self-managed merchant, active immediately
	// with no TaxCloud invite. Your platform tracks the merchant's nexus footprint
	// while the merchant remains responsible for their own registration, filing,
	// and remittance.
	//
	// Self-managed merchants support CalculateMerchantCart only. Every other
	// merchant transaction endpoint returns HTTP 403 for them.
	MerchantTypeSelfManaged = "self-managed"
)

// Merchant lifecycle statuses returned on reads.
const (
	// MerchantStatusTaxCloudInvited means the TaxCloud invite was sent but not yet accepted.
	MerchantStatusTaxCloudInvited = "taxcloud_invited"

	// MerchantStatusTaxCloudConnected means TaxCloud credentials are set and active.
	MerchantStatusTaxCloudConnected = "taxcloud_connected"

	// MerchantStatusTaxCloudDisconnected means the merchant's TaxCloud connection is no longer active.
	MerchantStatusTaxCloudDisconnected = "taxcloud_disconnected"

	// MerchantStatusExternalCompliance is the status of a self-managed merchant.
	MerchantStatusExternalCompliance = "external_compliance"
)

// CreateMerchantRequest is the request payload for creating a merchant.
type CreateMerchantRequest struct {
	// MerchantName is the legal or trading name of the merchant business (required, 1-255 characters).
	MerchantName string `json:"merchantName"`

	// ContactFirst is the first name of the merchant's primary contact (optional).
	ContactFirst string `json:"contactFirst,omitempty"`

	// ContactLast is the last name of the merchant's primary contact (optional).
	ContactLast string `json:"contactLast,omitempty"`

	// ContactEmail is the email address of the merchant's primary contact (optional).
	// Used for TaxCloud invitations and notifications.
	ContactEmail string `json:"contactEmail,omitempty"`

	// SendTaxCloudInvite sends an invite to set up and connect a TaxCloud account
	// to a merchant who does not already use TaxCloud (optional).
	//
	// Note the JSON key is "sendTaxcloudInvite" to match the API.
	SendTaxCloudInvite *bool `json:"sendTaxcloudInvite,omitempty"`

	// ReferenceID is the ID you use in your own system to identify this merchant (optional).
	ReferenceID string `json:"referenceId,omitempty"`

	// MerchantType is the compliance model for the merchant (optional, defaults to
	// MerchantTypeTaxCloud). Use MerchantTypeTaxCloud or MerchantTypeSelfManaged.
	//
	// Note the JSON key is "merchant_type" to match the API.
	MerchantType string `json:"merchant_type,omitempty"`
}

// MerchantOperationResponse is the acknowledgement returned by the merchant
// mutation endpoints: create, update, delete, and the credentials operations.
type MerchantOperationResponse struct {
	// MerchantID is the UUID of the merchant the operation applied to.
	MerchantID string `json:"merchantId"`

	// Message is a human-readable description of the result.
	Message string `json:"message"`

	// Status is the result status of the operation (e.g. "success").
	Status string `json:"status"`
}

// UpdateMerchantRequest is the request payload for updating a merchant.
type UpdateMerchantRequest struct {
	// MerchantID is the UUID of the merchant to update (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// Update holds the new merchant details (required).
	Update MerchantUpdate `json:"update"`
}

// MerchantUpdate holds the mutable fields of a merchant.
type MerchantUpdate struct {
	// MerchantName is the new legal or trading name of the merchant business
	// (required, 1-255 characters).
	MerchantName string `json:"merchantName"`

	// ContactFirst is the updated first name of the merchant's primary contact (optional).
	ContactFirst string `json:"contactFirst,omitempty"`

	// ContactLast is the updated last name of the merchant's primary contact (optional).
	ContactLast string `json:"contactLast,omitempty"`

	// ContactEmail is the updated email address of the merchant's primary contact (optional).
	ContactEmail string `json:"contactEmail,omitempty"`

	// ReferenceID is the ID you use in your own system to identify this merchant (optional).
	ReferenceID string `json:"referenceId,omitempty"`
}

// DeleteMerchantRequest is the request payload for soft-deleting a merchant.
type DeleteMerchantRequest struct {
	// MerchantID is the UUID of the merchant to soft-delete (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`
}

// GetMerchantRequest is the request payload for retrieving a single merchant.
type GetMerchantRequest struct {
	// MerchantID is the UUID of the merchant to retrieve (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`
}

// Merchant is a merchant record as returned by GetMerchant and ListMerchants.
type Merchant struct {
	// MerchantID is the UUID of the merchant.
	MerchantID string `json:"merchantId"`

	// MerchantName is the legal or trading name of the merchant business.
	MerchantName string `json:"merchantName"`

	// ContactFirst is the first name of the merchant's primary contact.
	ContactFirst string `json:"contactFirst,omitempty"`

	// ContactLast is the last name of the merchant's primary contact.
	ContactLast string `json:"contactLast,omitempty"`

	// ContactEmail is the email address of the merchant's primary contact.
	ContactEmail string `json:"contactEmail,omitempty"`

	// ReferenceID is the ID you use in your own system to identify this merchant.
	ReferenceID string `json:"referenceId,omitempty"`

	// Status is the derived lifecycle status of the merchant. It is one of
	// MerchantStatusTaxCloudInvited, MerchantStatusTaxCloudConnected,
	// MerchantStatusTaxCloudDisconnected, or MerchantStatusExternalCompliance.
	Status string `json:"status"`
}

// IsSelfManaged reports whether the merchant operates under the self-managed
// compliance model, in which case only CalculateMerchantCart is available.
func (m *Merchant) IsSelfManaged() bool {
	return m.Status == MerchantStatusExternalCompliance
}

// SetMerchantCredentialsRequest is the request payload for storing a merchant's
// TaxCloud credentials. Credentials are stored encrypted at rest by the API.
type SetMerchantCredentialsRequest struct {
	// MerchantID is the UUID of the merchant whose TaxCloud credentials are being set (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// ConnectionID is the TaxCloud connection ID that pairs with the API key to
	// identify the merchant's TaxCloud integration (required).
	ConnectionID string `json:"connectionId"`

	// APIKey is the TaxCloud API key to associate with the merchant (required).
	APIKey string `json:"apiKey"`
}

// DeleteMerchantCredentialsRequest is the request payload for removing a
// merchant's stored TaxCloud credentials.
type DeleteMerchantCredentialsRequest struct {
	// MerchantID is the UUID of the merchant whose TaxCloud credentials are being
	// deleted (required). The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`
}
