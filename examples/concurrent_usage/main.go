package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/ziptax/ziptax-go"
)

type Result struct {
	Address string
	Rate    float64
	Err     error
}

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

	// List of addresses to look up
	addresses := []string{
		"200 Spectrum Center Drive, Irvine, CA 92618",
		"1 Apple Park Way, Cupertino, CA 95014",
		"1600 Amphitheatre Parkway, Mountain View, CA 94043",
		"410 Terry Avenue North, Seattle, WA 98109",
		"1 Microsoft Way, Redmond, WA 98052",
	}

	fmt.Println("=== Concurrent Sales Tax Lookups ===")
	fmt.Printf("Looking up tax rates for %d addresses concurrently...\n\n", len(addresses))

	// Create channels for results
	results := make(chan Result, len(addresses))
	var wg sync.WaitGroup

	// Launch goroutines for concurrent requests
	ctx := context.Background()
	for _, addr := range addresses {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()

			// Perform the API request
			response, err := client.GetSalesTaxByAddress(ctx, address)
			if err != nil {
				results <- Result{Address: address, Err: err}
				return
			}

			// Extract the total tax rate
			var totalRate float64
			if len(response.TaxSummaries) > 0 {
				totalRate = response.TaxSummaries[0].Rate
			}

			results <- Result{
				Address: response.AddressDetail.NormalizedAddress,
				Rate:    totalRate,
			}
		}(addr)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and display results
	successCount := 0
	errorCount := 0

	for result := range results {
		if result.Err != nil {
			fmt.Printf("❌ Error for %s: %v\n", result.Address, result.Err)
			errorCount++
		} else {
			fmt.Printf("✓ %s: %.4f%%\n", result.Address, result.Rate*100)
			successCount++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", errorCount)
	fmt.Printf("Total: %d\n", len(addresses))
}
