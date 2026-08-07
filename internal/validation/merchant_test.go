package validation

import (
	"strings"
	"testing"
)

func TestValidateMerchantID(t *testing.T) {
	tests := []struct {
		name       string
		merchantID string
		wantErr    bool
	}{
		{"valid UUID", "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e", false},
		{"any non-empty value", "m1", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMerchantID(tt.merchantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMerchantID(%q) error = %v, wantErr %v", tt.merchantID, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMerchantName(t *testing.T) {
	tests := []struct {
		name         string
		merchantName string
		wantErr      bool
	}{
		{"valid", "Acme Supply Co", false},
		{"single character", "A", false},
		{"at the limit", strings.Repeat("a", MaxMerchantNameLength), false},
		{"empty", "", true},
		{"over the limit", strings.Repeat("a", MaxMerchantNameLength+1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMerchantName(tt.merchantName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMerchantName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMerchantType(t *testing.T) {
	tests := []struct {
		name         string
		merchantType string
		wantErr      bool
	}{
		{"empty defaults server-side", "", false},
		{"taxcloud", "taxcloud", false},
		{"self-managed", "self-managed", false},
		{"unknown", "hybrid", true},
		{"wrong case", "TaxCloud", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMerchantType(tt.merchantType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMerchantType(%q) error = %v, wantErr %v", tt.merchantType, err, tt.wantErr)
			}
		})
	}
}

func TestValidateMerchantCredentials(t *testing.T) {
	valid := func() *MerchantCredentialsInput {
		return &MerchantCredentialsInput{
			MerchantID:   "m1",
			ConnectionID: "c1",
			APIKey:       "k1",
		}
	}

	tests := []struct {
		name    string
		input   *MerchantCredentialsInput
		wantErr bool
	}{
		{"valid", valid(), false},
		{"nil input", nil, true},
		{"missing merchant ID", &MerchantCredentialsInput{ConnectionID: "c1", APIKey: "k1"}, true},
		{"missing connection ID", &MerchantCredentialsInput{MerchantID: "m1", APIKey: "k1"}, true},
		{"missing API key", &MerchantCredentialsInput{MerchantID: "m1", ConnectionID: "c1"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMerchantCredentials(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMerchantCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// validCartInput returns an input that passes, so each case mutates one field.
func validCartInput() *MerchantCartInput {
	return &MerchantCartInput{
		MerchantID: "m1",
		Carts: []MerchantCartItemInput{
			{
				CustomerID:       "customer-453",
				DestinationLine1: "200 Spectrum Center Dr",
				DestinationCity:  "Irvine",
				DestinationState: "CA",
				DestinationZip:   "92618",
				OriginLine1:      "323 Washington Ave N",
				OriginCity:       "Minneapolis",
				OriginState:      "MN",
				OriginZip:        "55401",
				LineItems: []MerchantCartLineItemInput{
					{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}
}

func TestValidateMerchantCartRequest(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*MerchantCartInput)
		wantErr string
	}{
		{"valid", func(*MerchantCartInput) {}, ""},
		{"missing merchant ID", func(i *MerchantCartInput) { i.MerchantID = "" }, "merchantId is required"},
		{"no carts", func(i *MerchantCartInput) { i.Carts = nil }, "at least 1 cart"},
		{"missing customer ID", func(i *MerchantCartInput) { i.Carts[0].CustomerID = "" }, "items[0].customerId is required"},
		{"missing destination line1", func(i *MerchantCartInput) { i.Carts[0].DestinationLine1 = "" }, "items[0].destination.line1 is required"},
		{"missing destination city", func(i *MerchantCartInput) { i.Carts[0].DestinationCity = "" }, "items[0].destination.city is required"},
		{"missing destination state", func(i *MerchantCartInput) { i.Carts[0].DestinationState = "" }, "items[0].destination.state is required"},
		{"missing destination zip", func(i *MerchantCartInput) { i.Carts[0].DestinationZip = "" }, "items[0].destination.zip is required"},
		{"missing origin line1", func(i *MerchantCartInput) { i.Carts[0].OriginLine1 = "" }, "items[0].origin.line1 is required"},
		{"no line items", func(i *MerchantCartInput) { i.Carts[0].LineItems = nil }, "items[0].lineItems must contain at least 1 item"},
		{"missing item ID", func(i *MerchantCartInput) { i.Carts[0].LineItems[0].ItemID = "" }, "items[0].lineItems[0].itemId is required"},
		{"negative price", func(i *MerchantCartInput) { i.Carts[0].LineItems[0].Price = -1 }, "must not be negative"},
		{"zero price is allowed", func(i *MerchantCartInput) { i.Carts[0].LineItems[0].Price = 0 }, ""},
		{"zero quantity", func(i *MerchantCartInput) { i.Carts[0].LineItems[0].Quantity = 0 }, "quantity must be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCartInput()
			tt.mutate(input)

			err := ValidateMerchantCartRequest(input)
			assertErrContains(t, err, tt.wantErr)
		})
	}

	t.Run("nil input", func(t *testing.T) {
		if err := ValidateMerchantCartRequest(nil); err == nil {
			t.Error("expected an error for nil input")
		}
	})

	t.Run("too many carts", func(t *testing.T) {
		input := validCartInput()
		cart := input.Carts[0]
		for len(input.Carts) <= MaxMerchantCarts {
			input.Carts = append(input.Carts, cart)
		}
		assertErrContains(t, ValidateMerchantCartRequest(input), "must not exceed 100 carts")
	})
}

func TestValidateMerchantOrderRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   *MerchantOrderInput
		wantErr string
	}{
		{"valid", &MerchantOrderInput{MerchantID: "m1", OrderID: "o1"}, ""},
		{"nil input", nil, "cannot be nil"},
		{"missing merchant ID", &MerchantOrderInput{OrderID: "o1"}, "merchantId is required"},
		{"missing order ID", &MerchantOrderInput{MerchantID: "m1"}, "orderId is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrContains(t, ValidateMerchantOrderRequest(tt.input), tt.wantErr)
		})
	}
}

func TestValidateMerchantOrderFromCartRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   *MerchantOrderFromCartInput
		wantErr string
	}{
		{"valid", &MerchantOrderFromCartInput{MerchantID: "m1", CartID: "c1", OrderID: "o1"}, ""},
		{"nil input", nil, "cannot be nil"},
		{"missing merchant ID", &MerchantOrderFromCartInput{CartID: "c1", OrderID: "o1"}, "merchantId is required"},
		{"missing cart ID", &MerchantOrderFromCartInput{MerchantID: "m1", OrderID: "o1"}, "cartId is required"},
		{"missing order ID", &MerchantOrderFromCartInput{MerchantID: "m1", CartID: "c1"}, "orderId is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrContains(t, ValidateMerchantOrderFromCartRequest(tt.input), tt.wantErr)
		})
	}
}

func validCreateOrderInput() *MerchantCreateOrderInput {
	return &MerchantCreateOrderInput{
		MerchantID:      "m1",
		OrderID:         "o1",
		CustomerID:      "customer-453",
		TransactionDate: "2026-08-01T14:00:00Z",
		CompletedDate:   "2026-08-02T09:15:00Z",
		LineItems: []MerchantCartLineItemInput{
			{ItemID: "item-1", Price: 10.75, Quantity: 1.5},
		},
	}
}

func TestValidateMerchantCreateOrderRequest(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*MerchantCreateOrderInput)
		wantErr string
	}{
		{"valid", func(*MerchantCreateOrderInput) {}, ""},
		{"missing merchant ID", func(i *MerchantCreateOrderInput) { i.MerchantID = "" }, "merchantId is required"},
		{"missing order ID", func(i *MerchantCreateOrderInput) { i.OrderID = "" }, "orderId is required"},
		{"missing customer ID", func(i *MerchantCreateOrderInput) { i.CustomerID = "" }, "customerId is required"},
		{"missing transaction date", func(i *MerchantCreateOrderInput) { i.TransactionDate = "" }, "transactionDate is required"},
		{"missing completed date", func(i *MerchantCreateOrderInput) { i.CompletedDate = "" }, "completedDate is required"},
		{"no line items", func(i *MerchantCreateOrderInput) { i.LineItems = nil }, "lineItems must contain at least 1 item"},
		{"missing item ID", func(i *MerchantCreateOrderInput) { i.LineItems[0].ItemID = "" }, "lineItems[0].itemId is required"},
		{"negative price", func(i *MerchantCreateOrderInput) { i.LineItems[0].Price = -0.01 }, "must not be negative"},
		{"zero quantity", func(i *MerchantCreateOrderInput) { i.LineItems[0].Quantity = 0 }, "quantity must be greater than 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCreateOrderInput()
			tt.mutate(input)
			assertErrContains(t, ValidateMerchantCreateOrderRequest(input), tt.wantErr)
		})
	}

	t.Run("nil input", func(t *testing.T) {
		if err := ValidateMerchantCreateOrderRequest(nil); err == nil {
			t.Error("expected an error for nil input")
		}
	})
}

func validCertificateInput() *MerchantCertificateInput {
	return &MerchantCertificateInput{
		MerchantID:        "m1",
		CustomerID:        "customer-453",
		CustomerName:      "Acme Reseller LLC",
		Reason:            "Resale",
		ReasonDescription: "Resale",
		BusinessType:      "RetailTrade",
		AddressLine1:      "323 Washington Ave N",
		AddressCity:       "Minneapolis",
		AddressState:      "MN",
		AddressZip:        "55401",
		States:            []string{"MN"},
	}
}

func TestValidateMerchantCertificateRequest(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*MerchantCertificateInput)
		wantErr string
	}{
		{"valid", func(*MerchantCertificateInput) {}, ""},
		{"missing merchant ID", func(i *MerchantCertificateInput) { i.MerchantID = "" }, "merchantId is required"},
		{"missing customer ID", func(i *MerchantCertificateInput) { i.CustomerID = "" }, "customerId is required"},
		{"missing customer name", func(i *MerchantCertificateInput) { i.CustomerName = "" }, "customerName is required"},
		{"missing reason", func(i *MerchantCertificateInput) { i.Reason = "" }, "reason is required"},
		{"missing reason description", func(i *MerchantCertificateInput) { i.ReasonDescription = "" }, "reasonDescription is required"},
		{
			name: "reason description at the limit",
			mutate: func(i *MerchantCertificateInput) {
				i.ReasonDescription = strings.Repeat("a", MaxReasonDescriptionLength)
			},
			wantErr: "",
		},
		{
			name: "reason description over the limit",
			mutate: func(i *MerchantCertificateInput) {
				i.ReasonDescription = strings.Repeat("a", MaxReasonDescriptionLength+1)
			},
			wantErr: "must not exceed 20 characters",
		},
		{"missing business type", func(i *MerchantCertificateInput) { i.BusinessType = "" }, "customerBusinessType is required"},
		{"missing address line1", func(i *MerchantCertificateInput) { i.AddressLine1 = "" }, "address.line1 is required"},
		{"missing address zip", func(i *MerchantCertificateInput) { i.AddressZip = "" }, "address.zip is required"},
		{"no states", func(i *MerchantCertificateInput) { i.States = nil }, "states must contain at least 1 state"},
		{"blank state", func(i *MerchantCertificateInput) { i.States = []string{""} }, "states[0].abbreviation is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validCertificateInput()
			tt.mutate(input)
			assertErrContains(t, ValidateMerchantCertificateRequest(input), tt.wantErr)
		})
	}

	t.Run("nil input", func(t *testing.T) {
		if err := ValidateMerchantCertificateRequest(nil); err == nil {
			t.Error("expected an error for nil input")
		}
	})
}

func TestValidateMerchantCertificateRef(t *testing.T) {
	tests := []struct {
		name    string
		input   *MerchantCertificateRefInput
		wantErr string
	}{
		{"valid", &MerchantCertificateRefInput{MerchantID: "m1", CertificateID: "cert-1"}, ""},
		{"nil input", nil, "cannot be nil"},
		{"missing merchant ID", &MerchantCertificateRefInput{CertificateID: "cert-1"}, "merchantId is required"},
		{"missing certificate ID", &MerchantCertificateRefInput{MerchantID: "m1"}, "certificateId is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrContains(t, ValidateMerchantCertificateRef(tt.input), tt.wantErr)
		})
	}
}

func TestValidateMerchantCertificateListRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   *MerchantCertificateListInput
		wantErr string
	}{
		{"valid with defaults", &MerchantCertificateListInput{MerchantID: "m1"}, ""},
		{"valid at the limit", &MerchantCertificateListInput{MerchantID: "m1", Limit: MaxCertificateListLimit}, ""},
		{"valid sort by createdDate", &MerchantCertificateListInput{MerchantID: "m1", SortBy: "createdDate"}, ""},
		{"valid sort by id", &MerchantCertificateListInput{MerchantID: "m1", SortBy: "id"}, ""},
		{"nil input", nil, "cannot be nil"},
		{"missing merchant ID", &MerchantCertificateListInput{}, "merchantId is required"},
		{"negative limit", &MerchantCertificateListInput{MerchantID: "m1", Limit: -1}, "must not be negative"},
		{"limit over the maximum", &MerchantCertificateListInput{MerchantID: "m1", Limit: 101}, "must not exceed 100"},
		{"unknown sort field", &MerchantCertificateListInput{MerchantID: "m1", SortBy: "customerName"}, "sortBy must be"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertErrContains(t, ValidateMerchantCertificateListRequest(tt.input), tt.wantErr)
		})
	}
}

// assertErrContains checks err against want: an empty want means no error is expected.
func assertErrContains(t *testing.T, err error, want string) {
	t.Helper()

	if want == "" {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		return
	}

	if err == nil {
		t.Fatalf("expected an error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), want)
	}
}
