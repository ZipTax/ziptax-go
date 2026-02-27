package validation

import (
	"fmt"
	"strings"
)

// ParsedAddress represents the result of parsing a single-string address into structured components.
type ParsedAddress struct {
	Line1       string
	Line2       string // Optional second line (suite, apartment, unit, etc.)
	City        string
	State       string
	Zip         string
	CountryCode string
}

// ParseAddress parses a single-string US address into structured components for TaxCloud.
//
// The parsing strategy works from the edges inward:
//   - First segment: street address (Line1)
//   - Last segment: state and zip code (e.g., "CA 92618" or "MN 55401-2427")
//   - Second-to-last segment: city
//   - Any segments between first and second-to-last: joined into Line2
//
// This correctly handles addresses with or without a suite/unit line:
//
//	"200 Spectrum Center Dr, Irvine, CA 92618"
//	  -> Line1: "200 Spectrum Center Dr", City: "Irvine"
//
//	"200 Spectrum Center Dr, Suite 100, Irvine, CA 92618"
//	  -> Line1: "200 Spectrum Center Dr", Line2: "Suite 100", City: "Irvine"
//
//	"200 Spectrum Center Dr, Bldg A, Suite 100, Irvine, CA 92618"
//	  -> Line1: "200 Spectrum Center Dr", Line2: "Bldg A, Suite 100", City: "Irvine"
//
// CountryCode defaults to "US".
func ParseAddress(address string) (*ParsedAddress, error) {
	address = strings.TrimSpace(address)
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	// Split by comma
	parts := strings.Split(address, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf(
			"address must contain at least 3 comma-separated segments (street, city, state zip): got %d segment(s) in %q",
			len(parts), address,
		)
	}

	// First segment: street address (Line1)
	line1 := strings.TrimSpace(parts[0])
	if line1 == "" {
		return nil, fmt.Errorf("street address (first segment) cannot be empty in %q", address)
	}

	// Last segment: state and zip (e.g., "CA 92618" or "MN 55401-2427")
	stateZip := strings.TrimSpace(parts[len(parts)-1])
	if stateZip == "" {
		return nil, fmt.Errorf("state and zip (last segment) cannot be empty in %q", address)
	}

	// Split the last segment by whitespace to get state and zip
	stateZipParts := strings.Fields(stateZip)
	if len(stateZipParts) < 2 {
		return nil, fmt.Errorf(
			"last segment must contain both state and zip code separated by space (e.g., 'CA 92618'): got %q in %q",
			stateZip, address,
		)
	}

	state := stateZipParts[0]
	zip := stateZipParts[1]

	// Validate state is 2 characters
	if len(state) != 2 {
		return nil, fmt.Errorf(
			"state must be a 2-letter abbreviation (e.g., 'CA'): got %q in %q",
			state, address,
		)
	}

	// Validate zip code format (5-digit or 9-digit with hyphen)
	if err := ValidatePostalCode(zip); err != nil {
		return nil, fmt.Errorf("invalid zip code in address %q: %w", address, err)
	}

	// Second-to-last segment: city
	city := strings.TrimSpace(parts[len(parts)-2])
	if city == "" {
		return nil, fmt.Errorf("city (second-to-last segment) cannot be empty in %q", address)
	}

	// Middle segments (between first and second-to-last): Line2
	var line2 string
	if len(parts) > 3 {
		middleParts := make([]string, 0, len(parts)-3)
		for _, p := range parts[1 : len(parts)-2] {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				middleParts = append(middleParts, trimmed)
			}
		}
		if len(middleParts) > 0 {
			line2 = strings.Join(middleParts, ", ")
		}
	}

	return &ParsedAddress{
		Line1:       line1,
		Line2:       line2,
		City:        city,
		State:       strings.ToUpper(state),
		Zip:         zip,
		CountryCode: "US",
	}, nil
}
