package models

// V60Response represents the main response structure for v6.0 API.
type V60Response struct {
	Metadata      V60Metadata       `json:"metadata"`
	BaseRates     []V60BaseRate     `json:"baseRates,omitempty"`
	Service       V60Service        `json:"service"`
	Shipping      V60Shipping       `json:"shipping"`
	SourcingRules *V60SourcingRules `json:"sourcingRules,omitempty"`
	TaxSummaries  []V60TaxSummary   `json:"taxSummaries,omitempty"`
	AddressDetail V60AddressDetail  `json:"addressDetail"`
}

// V60Metadata contains metadata about the API response.
type V60Metadata struct {
	Version  string          `json:"version"`
	Response V60ResponseInfo `json:"response"`
}

// V60ResponseInfo contains detailed response information nested in metadata.
type V60ResponseInfo struct {
	Code       int    `json:"code"`
	Name       string `json:"name"`
	Message    string `json:"message"`
	Definition string `json:"definition"`
}

// V60AddressDetail contains geocoded address information.
type V60AddressDetail struct {
	NormalizedAddress string  `json:"normalizedAddress"`
	Incorporated      string  `json:"incorporated"`
	GeoLat            float64 `json:"geoLat"`
	GeoLng            float64 `json:"geoLng"`
}
