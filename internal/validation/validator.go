package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// historicalDatePattern matches YYYY-MM format
	historicalDatePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}$`)
	// postalCodePattern matches US postal codes in 5-digit or 9-digit format
	postalCodePattern = regexp.MustCompile(`^[0-9]{5}(-[0-9]{4})?$`)
)

// ValidateAddress validates an address string.
func ValidateAddress(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if len(address) > 100 {
		return fmt.Errorf("address exceeds maximum length of 100 characters")
	}
	return nil
}

// ValidateCoordinates validates latitude and longitude values.
func ValidateCoordinates(lat, lng string) error {
	if lat == "" {
		return fmt.Errorf("latitude cannot be empty")
	}
	if lng == "" {
		return fmt.Errorf("longitude cannot be empty")
	}
	if len(lat) > 100 {
		return fmt.Errorf("latitude exceeds maximum length of 100 characters")
	}
	if len(lng) > 100 {
		return fmt.Errorf("longitude exceeds maximum length of 100 characters")
	}
	return nil
}

// ValidateHistoricalDate validates a historical date in YYYY-MM format.
// Month must be between 01 and 12. Year must be a positive number.
func ValidateHistoricalDate(date string) error {
	if date == "" {
		return nil // Optional parameter
	}
	if !historicalDatePattern.MatchString(date) {
		return fmt.Errorf("historical date must be in YYYY-MM format")
	}

	// Validate month range (01-12)
	parts := strings.SplitN(date, "-", 2)
	month, _ := strconv.Atoi(parts[1])
	if month < 1 || month > 12 {
		return fmt.Errorf("historical date month must be between 01 and 12")
	}

	return nil
}

// ValidateCountryCode validates a country code.
func ValidateCountryCode(code string) error {
	if code == "" {
		return nil // Optional parameter with default
	}
	validCodes := map[string]bool{
		"USA": true,
		"CAN": true,
	}
	if !validCodes[code] {
		return fmt.Errorf("country code must be USA or CAN")
	}
	return nil
}

// ValidateFormat validates a response format.
func ValidateFormat(format string) error {
	if format == "" {
		return nil // Optional parameter with default
	}
	validFormats := map[string]bool{
		"json": true,
		"xml":  true,
	}
	if !validFormats[format] {
		return fmt.Errorf("format must be json or xml")
	}
	return nil
}

// ValidatePostalCode validates a US postal code in 5-digit or 9-digit format.
func ValidatePostalCode(postalCode string) error {
	postalCode = strings.TrimSpace(postalCode)
	if postalCode == "" {
		return fmt.Errorf("postal code cannot be empty")
	}
	if !postalCodePattern.MatchString(postalCode) {
		return fmt.Errorf("postal code must be in 5-digit (e.g., 92694) or 9-digit (e.g., 92694-1234) format")
	}
	return nil
}

// NormalizePostalCode trims whitespace and strips the 9-digit suffix from a postal code,
// returning only the 5-digit base. The API only supports 5-digit postal codes.
func NormalizePostalCode(postalCode string) string {
	postalCode = strings.TrimSpace(postalCode)
	if idx := strings.Index(postalCode, "-"); idx != -1 {
		return postalCode[:idx]
	}
	return postalCode
}
