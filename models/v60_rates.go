package models

// V60BaseRate represents a base tax rate for a specific jurisdiction.
type V60BaseRate struct {
	Rate           float64 `json:"rate"`
	RateID         *string `json:"rateId,omitempty"`
	JurType        string  `json:"jurType"`
	JurName        string  `json:"jurName"`
	JurDescription *string `json:"jurDescription,omitempty"`
	JurTaxCode     *string `json:"jurTaxCode,omitempty"`
}

// V60Service represents service taxability information.
type V60Service struct {
	AdjustmentType string `json:"adjustmentType"`
	Taxable        string `json:"taxable"`
	Description    string `json:"description"`
}

// V60Shipping represents shipping taxability information.
type V60Shipping struct {
	AdjustmentType string `json:"adjustmentType"`
	Taxable        string `json:"taxable"`
	Description    string `json:"description"`
}

// V60SourcingRules represents sourcing rules (origin vs destination).
type V60SourcingRules struct {
	AdjustmentType string `json:"adjustmentType"`
	Description    string `json:"description"`
	Value          string `json:"value"`
}

// V60TaxSummary represents a tax rate summary.
type V60TaxSummary struct {
	Rate         float64          `json:"rate"`
	TaxType      string           `json:"taxType"`
	SummaryName  string           `json:"summaryName"`
	DisplayRates []V60DisplayRate `json:"displayRates"`
}

// V60DisplayRate represents a display rate breakdown within a tax summary.
type V60DisplayRate struct {
	Name string  `json:"name"`
	Rate float64 `json:"rate"`
}
