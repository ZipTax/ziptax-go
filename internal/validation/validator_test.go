package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "valid address",
			address: "200 Spectrum Center Dr, Irvine, CA 92618",
			wantErr: false,
		},
		{
			name:    "empty address",
			address: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			address: "   ",
			wantErr: true,
		},
		{
			name:    "address too long",
			address: strings.Repeat("a", 101),
			wantErr: true,
		},
		{
			name:    "max length address",
			address: strings.Repeat("a", 100),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     string
		lng     string
		wantErr bool
	}{
		{
			name:    "valid coordinates",
			lat:     "33.65253",
			lng:     "-117.74794",
			wantErr: false,
		},
		{
			name:    "empty lat",
			lat:     "",
			lng:     "-117.74794",
			wantErr: true,
		},
		{
			name:    "empty lng",
			lat:     "33.65253",
			lng:     "",
			wantErr: true,
		},
		{
			name:    "both empty",
			lat:     "",
			lng:     "",
			wantErr: true,
		},
		{
			name:    "lat too long",
			lat:     strings.Repeat("1", 101),
			lng:     "-117.74794",
			wantErr: true,
		},
		{
			name:    "lng too long",
			lat:     "33.65253",
			lng:     strings.Repeat("1", 101),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoordinates(tt.lat, tt.lng)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHistoricalDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date",
			date:    "2024-01",
			wantErr: false,
		},
		{
			name:    "valid date with leading zeros",
			date:    "2024-09",
			wantErr: false,
		},
		{
			name:    "empty date (optional)",
			date:    "",
			wantErr: false,
		},
		{
			name:    "invalid format - full date",
			date:    "2024-01-15",
			wantErr: true,
		},
		{
			name:    "invalid format - year only",
			date:    "2024",
			wantErr: true,
		},
		{
			name:    "invalid format - text",
			date:    "January 2024",
			wantErr: true,
		},
		{
			name:    "invalid month - 00",
			date:    "2024-00",
			wantErr: true,
		},
		{
			name:    "invalid month - 13",
			date:    "2024-13",
			wantErr: true,
		},
		{
			name:    "valid month - 12",
			date:    "2024-12",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHistoricalDate(tt.date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCountryCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "USA",
			code:    "USA",
			wantErr: false,
		},
		{
			name:    "CAN",
			code:    "CAN",
			wantErr: false,
		},
		{
			name:    "empty (optional)",
			code:    "",
			wantErr: false,
		},
		{
			name:    "invalid code",
			code:    "MEX",
			wantErr: true,
		},
		{
			name:    "lowercase",
			code:    "usa",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCountryCode(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "json",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "xml",
			format:  "xml",
			wantErr: false,
		},
		{
			name:    "empty (optional)",
			format:  "",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "yaml",
			wantErr: true,
		},
		{
			name:    "uppercase",
			format:  "JSON",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFormat(tt.format)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePostalCode(t *testing.T) {
	tests := []struct {
		name       string
		postalCode string
		wantErr    bool
	}{
		{
			name:       "valid 5-digit postal code",
			postalCode: "92694",
			wantErr:    false,
		},
		{
			name:       "valid 9-digit postal code",
			postalCode: "92694-1234",
			wantErr:    false,
		},
		{
			name:       "valid postal code with leading zeros",
			postalCode: "00001",
			wantErr:    false,
		},
		{
			name:       "valid 9-digit postal code with leading zeros",
			postalCode: "00001-0001",
			wantErr:    false,
		},
		{
			name:       "empty postal code",
			postalCode: "",
			wantErr:    true,
		},
		{
			name:       "whitespace only",
			postalCode: "   ",
			wantErr:    true,
		},
		{
			name:       "too short",
			postalCode: "1234",
			wantErr:    true,
		},
		{
			name:       "too long",
			postalCode: "123456",
			wantErr:    true,
		},
		{
			name:       "invalid 9-digit format - missing dash",
			postalCode: "926941234",
			wantErr:    true,
		},
		{
			name:       "invalid 9-digit format - wrong extension length",
			postalCode: "92694-123",
			wantErr:    true,
		},
		{
			name:       "contains letters",
			postalCode: "9269A",
			wantErr:    true,
		},
		{
			name:       "contains spaces",
			postalCode: "92 694",
			wantErr:    true,
		},
		{
			name:       "valid postal code with surrounding whitespace",
			postalCode: "  92694  ",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePostalCode(tt.postalCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizePostalCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "5-digit postal code unchanged",
			input:    "92694",
			expected: "92694",
		},
		{
			name:     "9-digit postal code stripped to 5-digit",
			input:    "92694-1234",
			expected: "92694",
		},
		{
			name:     "postal code with whitespace trimmed",
			input:    "  92694  ",
			expected: "92694",
		},
		{
			name:     "9-digit with whitespace",
			input:    "  92694-1234  ",
			expected: "92694",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePostalCode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
