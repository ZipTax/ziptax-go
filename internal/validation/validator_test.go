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
