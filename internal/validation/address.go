package validation

import (
	"fmt"
	"strings"
)

// ParsedAddress represents the result of parsing a single-string address into structured components.
type ParsedAddress struct {
	Line1       string
	City        string
	State       string
	Zip         string
	CountryCode string
}

// ParseAddress parses a single-string US address into structured components for TaxCloud.
//
// Expected format: "street, city, state zip" or "street, city, state zip-plus4"
// Examples:
//
//	"200 Spectrum Center Dr, Irvine, CA 92618"
//	"200 Spectrum Center Dr, Irvine, CA 92618-1905"
//	"323 Washington Ave N, Minneapolis, MN 55401-2427"
//
// The function splits by comma, expecting at least 3 segments:
//   - Segment 1: street address (line1)
//   - Segment 2: city
//   - Segment 3: state and zip code (e.g., "CA 92618" or "MN 55401-2427")
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

	// Extract line1 (first segment)
	line1 := strings.TrimSpace(parts[0])
	if line1 == "" {
		return nil, fmt.Errorf("street address (first segment) cannot be empty in %q", address)
	}

	// Extract city (second segment)
	city := strings.TrimSpace(parts[1])
	if city == "" {
		return nil, fmt.Errorf("city (second segment) cannot be empty in %q", address)
	}

	// Extract state and zip from last segment (e.g., "CA 92618" or "CA 92618-1905")
	// Use the last segment to handle addresses with extra commas in the middle
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

	return &ParsedAddress{
		Line1:       line1,
		City:        city,
		State:       strings.ToUpper(state),
		Zip:         zip,
		CountryCode: "US",
	}, nil
}
