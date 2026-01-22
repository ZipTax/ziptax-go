package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ziptax/ziptax-go"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ZIPTAX_API_KEY")
	if apiKey == "" {
		log.Fatal("ZIPTAX_API_KEY environment variable is required")
	}

	// Create a new client
	client, err := ziptax.NewClient(apiKey)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Example 1: Handle validation errors
	fmt.Println("=== Example 1: Validation Error ===")
	_, err = client.GetSalesTaxByAddress(ctx, "") // Empty address
	if err != nil {
		var validationErr *ziptax.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("Validation error:\n")
			fmt.Printf("  Field: %s\n", validationErr.Field)
			fmt.Printf("  Value: %s\n", validationErr.Value)
			fmt.Printf("  Message: %s\n", validationErr.Message)
		} else {
			fmt.Printf("Other error: %v\n", err)
		}
	}

	// Example 2: Handle invalid coordinates
	fmt.Println("\n=== Example 2: Invalid Coordinates ===")
	_, err = client.GetSalesTaxByGeoLocation(ctx, "", "")
	if err != nil {
		var validationErr *ziptax.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("Validation error: %s\n", validationErr.Message)
		}
	}

	// Example 3: Handle API errors (using invalid API key)
	fmt.Println("\n=== Example 3: API Error (Invalid API Key) ===")
	invalidClient, err := ziptax.NewClient("invalid-api-key-12345")
	if err != nil {
		fmt.Printf("Client creation failed: %v\n", err)
	} else {
		_, err = invalidClient.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
		if err != nil {
			var apiErr *ziptax.APIError
			if errors.As(err, &apiErr) {
				fmt.Printf("API error:\n")
				fmt.Printf("  Status Code: %d\n", apiErr.StatusCode)
				fmt.Printf("  Code: %d\n", apiErr.Code)
				fmt.Printf("  Name: %s\n", apiErr.Name)
				fmt.Printf("  Message: %s\n", apiErr.Message)
			} else {
				fmt.Printf("Other error: %v\n", err)
			}
		}
	}

	// Example 4: Successful request with proper error checking
	fmt.Println("\n=== Example 4: Successful Request ===")
	response, err := client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
	if err != nil {
		// Check for specific error types
		if errors.Is(err, ziptax.ErrInvalidAPIKey) {
			fmt.Println("Error: Invalid API key")
			return
		}
		if errors.Is(err, ziptax.ErrRateLimitExceeded) {
			fmt.Println("Error: Rate limit exceeded")
			return
		}

		// Handle validation errors
		var validationErr *ziptax.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("Validation error: %s\n", validationErr.Message)
			return
		}

		// Handle API errors
		var apiErr *ziptax.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("API error: %s\n", apiErr.Message)
			return
		}

		// Generic error
		fmt.Printf("Unexpected error: %v\n", err)
		return
	}

	// Success!
	fmt.Printf("✓ Successfully retrieved tax rates for %s\n", response.AddressDetail.NormalizedAddress)
	if len(response.TaxSummaries) > 0 {
		fmt.Printf("  Total tax rate: %.4f%%\n", response.TaxSummaries[0].Rate*100)
	}
}