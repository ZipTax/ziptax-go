package models

// V60AccountMetrics represents the account metrics returned by GET /account/v60/metrics.
//
// The v6.0 endpoint reports a single combined request counter. For the per-entitlement
// breakdown (core, geocoding, and merchant counters) use AccountMetrics via
// Client.GetDetailedAccountMetrics.
type V60AccountMetrics struct {
	// RequestCount is the number of requests consumed in the current period.
	RequestCount int64 `json:"request_count"`

	// RequestLimit is the maximum requests allowed for the account in the current period.
	RequestLimit int64 `json:"request_limit"`

	// UsagePercent is request usage as a percentage of the limit (0-100).
	UsagePercent float64 `json:"usage_percent"`

	// IsActive is whether the account is currently active.
	IsActive bool `json:"is_active"`

	// Message is an informational message about the account.
	Message string `json:"message"`
}

// AccountMetrics represents the account metrics returned by GET /account/metrics.
//
// This endpoint reports usage per entitlement: core tax lookups, geocoding, and
// merchant requests. For the simpler combined counter, use V60AccountMetrics via
// Client.GetAccountMetrics.
type AccountMetrics struct {
	// CoreRequestCount is the number of core (tax lookup) requests consumed in the
	// current billing period.
	CoreRequestCount int64 `json:"core_request_count"`

	// CoreRequestLimit is the maximum core requests allowed in the current period,
	// from the account's entitlement.
	CoreRequestLimit int64 `json:"core_request_limit"`

	// CoreUsagePercent is core request usage as a percentage of the limit (0-100).
	CoreUsagePercent float64 `json:"core_usage_percent"`

	// GeoEnabled is whether the account has the geocoding entitlement that allows
	// address and coordinate lookups.
	GeoEnabled bool `json:"geo_enabled"`

	// GeoRequestCount is the number of geocoding requests consumed in the current period.
	GeoRequestCount int64 `json:"geo_request_count"`

	// GeoRequestLimit is the maximum geocoding requests allowed in the current period.
	GeoRequestLimit int64 `json:"geo_request_limit"`

	// GeoUsagePercent is geocoding request usage as a percentage of the limit (0-100).
	GeoUsagePercent float64 `json:"geo_usage_percent"`

	// MerchantRequestCount is the number of merchant requests consumed in the current period.
	MerchantRequestCount int64 `json:"merchant_request_count"`

	// MerchantRequestLimit is the maximum merchant requests allowed in the current period.
	MerchantRequestLimit int64 `json:"merchant_request_limit"`

	// MerchantUsagePercent is merchant request usage as a percentage of the limit (0-100).
	MerchantUsagePercent float64 `json:"merchant_usage_percent"`

	// IsActive is whether the account is currently active and able to make requests.
	IsActive bool `json:"is_active"`

	// Message is an informational message about the account.
	Message string `json:"message"`
}
