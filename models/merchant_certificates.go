package models

// Customer business types accepted on an exemption certificate.
// Use BusinessTypeOther together with CustomerBusinessDescription when none of
// the specific types apply.
const (
	BusinessTypeAccommodationAndFoodServices           = "AccommodationAndFoodServices"
	BusinessTypeAgriculturalForestryFishingHunting     = "AgriculturalForestryFishingHunting"
	BusinessTypeConstruction                           = "Construction"
	BusinessTypeFinanceAndInsurance                    = "FinanceAndInsurance"
	BusinessTypeInformationPublishingAndCommunications = "InformationPublishingAndCommunications"
	BusinessTypeManufacturing                          = "Manufacturing"
	BusinessTypeMining                                 = "Mining"
	BusinessTypeRealEstate                             = "RealEstate"
	BusinessTypeRentalAndLeasing                       = "RentalAndLeasing"
	BusinessTypeRetailTrade                            = "RetailTrade"
	BusinessTypeTransportationAndWarehousing           = "TransportationAndWarehousing"
	BusinessTypeUtilities                              = "Utilities"
	BusinessTypeWholesaleTrade                         = "WholesaleTrade"
	BusinessTypeBusinessServices                       = "BusinessServices"
	BusinessTypeProfessionalServices                   = "ProfessionalServices"
	BusinessTypeEducationAndHealthCareServices         = "EducationAndHealthCareServices"
	BusinessTypeNonprofitOrganization                  = "NonprofitOrganization"
	BusinessTypeGovernment                             = "Government"
	BusinessTypeNotABusiness                           = "NotABusiness"
	BusinessTypeOther                                  = "Other"
)

// Exemption reasons accepted on an exemption certificate.
const (
	ExemptionReasonFederalGovernment                   = "FederalGovernment"
	ExemptionReasonStateOrLocalGovernment              = "StateOrLocalGovernment"
	ExemptionReasonTribalGovernment                    = "TribalGovernment"
	ExemptionReasonForeignDiplomat                     = "ForeignDiplomat"
	ExemptionReasonCharitableOrganization              = "CharitableOrganization"
	ExemptionReasonEducationalOrganization             = "EducationalOrganization"
	ExemptionReasonReligiousOrganization               = "ReligiousOrganization"
	ExemptionReasonResale                              = "Resale"
	ExemptionReasonAgriculturalProduction              = "AgriculturalProduction"
	ExemptionReasonIndustrialProductionOrManufacturing = "IndustrialProductionOrManufacturing"
	ExemptionReasonDirectPayPermit                     = "DirectPayPermit"
	ExemptionReasonDirectMail                          = "DirectMail"
	ExemptionReasonOther                               = "Other"
)

// Sort fields accepted when listing exemption certificates.
const (
	// CertificateSortByID sorts by certificate ID. This is the API default.
	CertificateSortByID = "id"

	// CertificateSortByCreatedDate sorts by the date the certificate was created.
	CertificateSortByCreatedDate = "createdDate"
)

// MerchantCreateCertificateRequest is the request payload for creating an exemption
// certificate for one of a merchant's customers.
//
// Reference the returned CertificateID as the ExemptionID on carts and orders to
// apply the exemption.
//
// Available only for TaxCloud-connected merchants. A self-managed merchant holds
// no exemption certificates, so the request is refused with HTTP 403.
type MerchantCreateCertificateRequest struct {
	// MerchantID is the UUID of the merchant (required).
	// The merchant must be owned by the calling account.
	MerchantID string `json:"merchantId"`

	// CustomerID is your identifier for the exempt customer (required).
	// Carts and orders submitted with this customerId can use the certificate.
	CustomerID string `json:"customerId"`

	// CustomerName is the name of the customer or organization the certificate is
	// issued to (required).
	CustomerName string `json:"customerName"`

	// CustomerBusinessType is the type of business the customer is (required).
	// Use one of the BusinessType constants.
	CustomerBusinessType string `json:"customerBusinessType"`

	// CustomerBusinessDescription is a free-text description of the business (optional).
	// Provide it when CustomerBusinessType is BusinessTypeOther.
	CustomerBusinessDescription string `json:"customerBusinessDescription,omitempty"`

	// Reason is why the customer is exempt (required).
	// Use one of the ExemptionReason constants.
	Reason string `json:"reason"`

	// ReasonDescription is a short free-text elaboration of the exemption reason
	// (required, maximum 20 characters).
	ReasonDescription string `json:"reasonDescription"`

	// Address is the customer's address (required).
	Address TaxCloudAddress `json:"address"`

	// States lists the states the exemption certificate is valid in (required).
	States []CertificateState `json:"states"`
}

// CertificateState is a state an exemption certificate is valid in.
type CertificateState struct {
	// Abbreviation is the two-letter abbreviation of the state (required).
	Abbreviation string `json:"abbreviation"`
}

// MerchantGetCertificateRequest is the request payload for retrieving a single
// exemption certificate.
type MerchantGetCertificateRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// CertificateID is the certificateId returned when the certificate was created (required).
	CertificateID string `json:"certificateId"`
}

// MerchantDeleteCertificateRequest is the request payload for deleting (disabling)
// an exemption certificate so it can no longer be applied to new transactions.
type MerchantDeleteCertificateRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// CertificateID is the certificateId returned when the certificate was created (required).
	CertificateID string `json:"certificateId"`
}

// MerchantListCertificatesRequest is the request payload for paging through a
// merchant's exemption certificates.
type MerchantListCertificatesRequest struct {
	// MerchantID is the UUID of the merchant (required).
	MerchantID string `json:"merchantId"`

	// CustomerID filters results to certificates belonging to this customerId (optional).
	CustomerID string `json:"customerId,omitempty"`

	// Cursor is the opaque pagination cursor from the NextCursor field of a previous
	// response (optional). Omit to start at the first page.
	Cursor string `json:"cursor,omitempty"`

	// Limit is the maximum number of certificates to return per page (optional).
	// Defaults to 20; maximum 100.
	Limit int64 `json:"limit,omitempty"`

	// SortBy selects the sort field (optional): CertificateSortByID or
	// CertificateSortByCreatedDate. Defaults to CertificateSortByID.
	SortBy string `json:"sortBy,omitempty"`

	// Ascending sorts results in ascending order (optional). Defaults to false (descending).
	Ascending *bool `json:"ascending,omitempty"`

	// Disabled set to true lists disabled (revoked) certificates instead of active
	// ones (optional). Defaults to false.
	Disabled *bool `json:"disabled,omitempty"`
}

// MerchantListCertificatesResponse is a page of exemption certificates.
type MerchantListCertificatesResponse struct {
	// Items holds the exemption certificates on this page of results.
	Items []MerchantCertificate `json:"items"`

	// Limit is the maximum number of results per page that was applied.
	Limit int64 `json:"limit"`

	// NextCursor is the opaque cursor to pass as Cursor on the next call to fetch
	// the following page. Empty when there are no further results.
	NextCursor string `json:"nextCursor"`
}

// MerchantCertificate is an exemption certificate held for a merchant's customer.
type MerchantCertificate struct {
	// CertificateID is TaxCloud's identifier for the certificate. Use it with
	// GetMerchantCertificate and DeleteMerchantCertificate, and as the ExemptionID
	// on carts and orders.
	CertificateID string `json:"certificateId"`

	// ConnectionID is the TaxCloud connection the certificate belongs to.
	ConnectionID string `json:"connectionId"`

	// AccountID is the TaxCloud account ID the certificate belongs to.
	AccountID int64 `json:"accountId"`

	// CustomerID is your identifier for the exempt customer.
	CustomerID string `json:"customerId"`

	// CustomerName is the name of the customer the certificate was issued to.
	CustomerName string `json:"customerName"`

	// CustomerBusinessType is the type of business the customer is.
	CustomerBusinessType string `json:"customerBusinessType"`

	// CustomerBusinessDescription is a free-text description of the business,
	// present when CustomerBusinessType is BusinessTypeOther.
	CustomerBusinessDescription string `json:"customerBusinessDescription,omitempty"`

	// Reason is why the customer is exempt.
	Reason string `json:"reason"`

	// ReasonDescription is the free-text elaboration of the exemption reason.
	ReasonDescription string `json:"reasonDescription"`

	// Address is the customer's address.
	Address TaxCloudAddressResponse `json:"address"`

	// States lists the states the certificate is valid in.
	States []CertificateState `json:"states,omitempty"`

	// CreatedDate is the RFC3339 datetime the certificate was created.
	CreatedDate string `json:"createdDate"`

	// DisabledAt is the RFC3339 datetime the certificate was disabled,
	// or nil while it is active.
	DisabledAt *string `json:"disabledAt"`

	// SinglePurchase is whether the certificate covers a single purchase only,
	// rather than being a blanket certificate.
	SinglePurchase bool `json:"singlePurchase"`
}

// IsActive reports whether the certificate has not been disabled and can still be
// applied to new transactions.
func (c *MerchantCertificate) IsActive() bool {
	return c.DisabledAt == nil || *c.DisabledAt == ""
}
