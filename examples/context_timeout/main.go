package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ziptax/ziptax-go"
	"github.com/ziptax/ziptax-go/models"
)

func main() {
	// Get API key from environment variable
	apiKey := os.Getenv("ZIPTAX_API_KEY")
	if apiKey == "" {
		log.Fatal("ZIPTAX_API_KEY environment variable is required")
	}

	// Create a new client with a longer default timeout
	client, err := ziptax.NewClient(
		apiKey,
		ziptax.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: Request with timeout context
	fmt.Println("=== Example 1: Request with Timeout ===")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	response, err := client.GetSalesTaxByAddress(ctx, "200 Spectrum Center Drive, Irvine, CA 92618")
	duration := time.Since(start)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Request timed out after %.2fs\n", duration.Seconds())
		} else {
			fmt.Printf("Request failed: %v\n", err)
		}
	} else {
		fmt.Printf("✓ Request completed in %.2fs\n", duration.Seconds())
		fmt.Printf("  Address: %s\n", response.AddressDetail.NormalizedAddress)
		if len(response.TaxSummaries) > 0 {
			fmt.Printf("  Tax rate: %.4f%%\n", response.TaxSummaries[0].Rate*100)
		}
	}

	// Example 2: Request with cancellation
	fmt.Println("\n=== Example 2: Request with Cancellation ===")
	ctx2, cancel2 := context.WithCancel(context.Background())

	// Cancel the context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel2()
	}()

	start = time.Now()
	_, err = client.GetSalesTaxByAddress(ctx2, "1 Apple Park Way, Cupertino, CA 95014")
	duration = time.Since(start)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Printf("Request was canceled after %.2fs\n", duration.Seconds())
		} else {
			fmt.Printf("Request failed: %v\n", err)
		}
	} else {
		fmt.Printf("✓ Request completed in %.2fs\n", duration.Seconds())
	}

	// Example 3: Multiple requests with shared timeout
	fmt.Println("\n=== Example 3: Multiple Requests with Shared Timeout ===")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel3()

	addresses := []string{
		"200 Spectrum Center Drive, Irvine, CA 92618",
		"1 Apple Park Way, Cupertino, CA 95014",
		"1600 Amphitheatre Parkway, Mountain View, CA 94043",
	}

	start = time.Now()
	for i, addr := range addresses {
		// Check if context is already canceled
		if ctx3.Err() != nil {
			fmt.Printf("Context canceled, skipping remaining requests\n")
			break
		}

		resp, err := client.GetSalesTaxByAddress(ctx3, addr)
		if err != nil {
			fmt.Printf("%d. Error for %s: %v\n", i+1, addr, err)
			continue
		}
		fmt.Printf("%d. ✓ %s: %.4f%%\n", i+1, resp.AddressDetail.NormalizedAddress,
			getRate(resp)*100)
	}

	duration = time.Since(start)
	fmt.Printf("\nAll requests completed in %.2fs\n", duration.Seconds())

	// Example 4: Request with deadline
	fmt.Println("\n=== Example 4: Request with Deadline ===")
	deadline := time.Now().Add(3 * time.Second)
	ctx4, cancel4 := context.WithDeadline(context.Background(), deadline)
	defer cancel4()

	fmt.Printf("Request deadline: %s\n", deadline.Format("15:04:05"))

	response4, err := client.GetSalesTaxByAddress(ctx4, "410 Terry Avenue North, Seattle, WA 98109")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Request exceeded deadline")
		} else {
			fmt.Printf("Request failed: %v\n", err)
		}
	} else {
		fmt.Printf("✓ Request completed before deadline\n")
		fmt.Printf("  Address: %s\n", response4.AddressDetail.NormalizedAddress)
	}
}

// Helper function to extract tax rate from response
func getRate(response *models.V60Response) float64 {
	if len(response.TaxSummaries) > 0 {
		return response.TaxSummaries[0].Rate
	}
	return 0.0
}
