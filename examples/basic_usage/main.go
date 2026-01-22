package main

import (
	"context"
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

	// Example 1: Get sales tax by address
	fmt.Println("=== Get Sales Tax by Address ===")
	ctx := context.Background()
	response, err := client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
	if err != nil {
		log.Fatalf("Failed to get tax rates: %v", err)
	}

	fmt.Printf("API Version: %s\n", response.Metadata.Version)
	fmt.Printf("Response Code: %d - %s\n", response.Metadata.Response.Code, response.Metadata.Response.Name)
	fmt.Printf("Normalized Address: %s\n", response.AddressDetail.NormalizedAddress)
	fmt.Printf("Coordinates: (%.6f, %.6f)\n", response.AddressDetail.GeoLat, response.AddressDetail.GeoLng)

	if len(response.BaseRates) > 0 {
		fmt.Println("\nBase Rates:")
		for _, rate := range response.BaseRates {
			fmt.Printf("  - %s: %.4f%% (%s)\n", rate.JurName, rate.Rate*100, rate.JurType)
		}
	}

	if len(response.TaxSummaries) > 0 {
		fmt.Println("\nTax Summaries:")
		for _, summary := range response.TaxSummaries {
			fmt.Printf("  - %s: %.4f%%\n", summary.SummaryName, summary.Rate*100)
		}
	}

	// Example 2: Get sales tax by geolocation
	fmt.Println("\n=== Get Sales Tax by Geolocation ===")
	geoResponse, err := client.GetSalesTaxByGeoLocation(ctx, "33.65253", "-117.74794")
	if err != nil {
		log.Fatalf("Failed to get tax rates by geolocation: %v", err)
	}

	fmt.Printf("Normalized Address: %s\n", geoResponse.AddressDetail.NormalizedAddress)
	fmt.Printf("Incorporated: %s\n", geoResponse.AddressDetail.Incorporated)

	// Example 3: Get account metrics
	fmt.Println("\n=== Get Account Metrics ===")
	metrics, err := client.GetAccountMetrics(ctx)
	if err != nil {
		log.Fatalf("Failed to get account metrics: %v", err)
	}

	fmt.Printf("Core Requests: %d / %d (%.2f%%)\n",
		metrics.CoreRequestCount, metrics.CoreRequestLimit, metrics.CoreUsagePercent)
	fmt.Printf("Geo Requests: %d / %d (%.2f%%)\n",
		metrics.GeoRequestCount, metrics.GeoRequestLimit, metrics.GeoUsagePercent)
	fmt.Printf("Geo Enabled: %v\n", metrics.GeoEnabled)
	fmt.Printf("Account Active: %v\n", metrics.IsActive)
	fmt.Printf("Message: %s\n", metrics.Message)
}