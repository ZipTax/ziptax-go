package models

// V60AccountMetrics represents account metrics and usage statistics.
type V60AccountMetrics struct {
	CoreRequestCount int64   `json:"core_request_count"`
	CoreRequestLimit int64   `json:"core_request_limit"`
	CoreUsagePercent float64 `json:"core_usage_percent"`
	GeoEnabled       bool    `json:"geo_enabled"`
	GeoRequestCount  int64   `json:"geo_request_count"`
	GeoRequestLimit  int64   `json:"geo_request_limit"`
	GeoUsagePercent  float64 `json:"geo_usage_percent"`
	IsActive         bool    `json:"is_active"`
	Message          string  `json:"message"`
}
