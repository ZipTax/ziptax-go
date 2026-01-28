package models

// V60PostalCodeResponse represents the response for postal code lookup.
// It returns a flat structure with multiple results for different cities
// that may share the same postal code.
type V60PostalCodeResponse struct {
	Version       string                     `json:"version"`
	RCode         int                        `json:"rCode"`
	Results       []V60PostalCodeResult      `json:"results"`
	AddressDetail V60PostalCodeAddressDetail `json:"addressDetail"`
}

// V60PostalCodeResult represents an individual tax rate result for a postal code location.
// A single postal code may return multiple results for different cities.
type V60PostalCodeResult struct {
	GeoPostalCode     string  `json:"geoPostalCode"`
	GeoCity           string  `json:"geoCity"`
	GeoCounty         string  `json:"geoCounty"`
	GeoState          string  `json:"geoState"`
	TaxSales          float64 `json:"taxSales"`
	TaxUse            float64 `json:"taxUse"`
	TxbService        string  `json:"txbService"`
	TxbFreight        string  `json:"txbFreight"`
	StateSalesTax     float64 `json:"stateSalesTax"`
	StateUseTax       float64 `json:"stateUseTax"`
	CitySalesTax      float64 `json:"citySalesTax"`
	CityUseTax        float64 `json:"cityUseTax"`
	CityTaxCode       string  `json:"cityTaxCode"`
	CountySalesTax    float64 `json:"countySalesTax"`
	CountyUseTax      float64 `json:"countyUseTax"`
	CountyTaxCode     string  `json:"countyTaxCode"`
	DistrictSalesTax  float64 `json:"districtSalesTax"`
	DistrictUseTax    float64 `json:"districtUseTax"`
	District1Code     string  `json:"district1Code"`
	District1SalesTax float64 `json:"district1SalesTax"`
	District1UseTax   float64 `json:"district1UseTax"`
	District2Code     string  `json:"district2Code"`
	District2SalesTax float64 `json:"district2SalesTax"`
	District2UseTax   float64 `json:"district2UseTax"`
	District3Code     string  `json:"district3Code"`
	District3SalesTax float64 `json:"district3SalesTax"`
	District3UseTax   float64 `json:"district3UseTax"`
	District4Code     string  `json:"district4Code"`
	District4SalesTax float64 `json:"district4SalesTax"`
	District4UseTax   float64 `json:"district4UseTax"`
	District5Code     string  `json:"district5Code"`
	District5SalesTax float64 `json:"district5SalesTax"`
	District5UseTax   float64 `json:"district5UseTax"`
	OriginDestination string  `json:"originDestination"`
}

// V60PostalCodeAddressDetail contains address details for postal code lookups.
// Note: For postal code lookups, normalized address and incorporation status
// are not available, and coordinates are returned as 0.
type V60PostalCodeAddressDetail struct {
	NormalizedAddress string  `json:"normalizedAddress"`
	Incorporated      string  `json:"incorporated"`
	GeoLat            float64 `json:"geoLat"`
	GeoLng            float64 `json:"geoLng"`
}
