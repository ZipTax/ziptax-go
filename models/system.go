package models

// Component health values reported by GET /system/health.
const (
	// HealthStatusOK means the component is healthy.
	HealthStatusOK = "ok"

	// HealthStatusConfigError means the AWS config could not be loaded (DynamoDB only).
	HealthStatusConfigError = "config_error"

	// HealthStatusConnectionError means the table could not be described (DynamoDB only).
	HealthStatusConnectionError = "connection_error"

	// HealthStatusEmpty means the tax-data cache is not loaded.
	HealthStatusEmpty = "empty"

	// HealthStatusPartial means the tax-data cache loaded fewer than the expected
	// number of records.
	HealthStatusPartial = "partial"
)

// HealthResponse is the response from the health check endpoint.
type HealthResponse struct {
	// Status is the overall health of the API. It is always "ok" on a served
	// response; component-level detail is in Components.
	Status string `json:"status"`

	// Components holds the per-component health detail.
	Components HealthComponents `json:"components"`
}

// HealthComponents holds per-component health detail.
type HealthComponents struct {
	// Dynamo is the DynamoDB connectivity status: HealthStatusOK,
	// HealthStatusConfigError, or HealthStatusConnectionError.
	Dynamo string `json:"dynamo"`

	// TaxData is the tax-data cache status: HealthStatusOK, HealthStatusEmpty,
	// or HealthStatusPartial.
	TaxData string `json:"taxdata"`

	// TaxDataCount is the number of tax-data records currently loaded in the
	// in-memory cache.
	TaxDataCount int64 `json:"taxdata_count"`
}

// IsHealthy reports whether every component is reporting HealthStatusOK.
func (h *HealthResponse) IsHealthy() bool {
	return h.Status == HealthStatusOK &&
		h.Components.Dynamo == HealthStatusOK &&
		h.Components.TaxData == HealthStatusOK
}

// SystemMetadata is the response from the system metadata endpoint.
type SystemMetadata struct {
	// GoVersion is the Go runtime version the running binary was built with.
	GoVersion string `json:"go_version"`

	// Hostname is the hostname of the instance serving the request.
	Hostname string `json:"hostname"`
}

// TICDataResponse is the response from the TIC codes endpoint. It holds the full
// catalog of Taxability Information Codes.
type TICDataResponse struct {
	// TICList is the catalog of TIC codes.
	TICList []TICListEntry `json:"tic_list"`
}

// TICListEntry wraps a single TIC in the catalog response.
type TICListEntry struct {
	// TIC is the Taxability Information Code detail.
	TIC TICDetail `json:"tic"`
}

// TICDetail describes a single Taxability Information Code.
type TICDetail struct {
	// ID is the TIC identifier, a numeric string. Use it as the taxabilityCode
	// parameter on rate requests, or as the TIC on merchant cart and order line items.
	ID string `json:"id"`

	// Parent is the TIC code of this code's parent category in the hierarchy.
	// Empty for top-level categories.
	Parent string `json:"parent"`

	// Title is the short, localized human-readable title of the TIC category.
	Title string `json:"title"`

	// Label is the longer, localized description of what the TIC category covers.
	Label string `json:"label"`

	// NLTitle is the non-localized (base English) title of the TIC category.
	NLTitle string `json:"nl_title"`

	// NLLabel is the non-localized (base English) description of the TIC category.
	NLLabel string `json:"nl_label"`
}
