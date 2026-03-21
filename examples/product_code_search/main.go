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

	ctx := context.Background()

	// Example 1: Search for product codes (TICs)
	fmt.Println("=== Search Product Codes ===")
	searchResp, err := client.SearchProductCodes(ctx, "baked goods sold in plastic packaging")
	if err != nil {
		log.Fatalf("Failed to search product codes: %v", err)
	}

	fmt.Printf("Query: %s\n", searchResp.Query)
	fmt.Printf("Results: %d\n\n", len(searchResp.Results))
	for _, result := range searchResp.Results {
		fmt.Printf("  TIC %s: %s\n", result.TicID, result.Label)
		fmt.Printf("    Description: %s\n", result.Description)
		fmt.Printf("    Rank: %s, Score: %s\n\n", result.Rank, result.Score)
	}

	// Example 2: Get AI-powered product code recommendation
	fmt.Println("=== Recommend Product Code ===")
	recResp, err := client.RecommendProductCode(ctx, "baked goods sold in plastic packaging")
	if err != nil {
		log.Fatalf("Failed to recommend product code: %v", err)
	}

	for _, prediction := range recResp.Predictions {
		if prediction.Status == "success" {
			fmt.Printf("Recommended TIC: %s (%s)\n", prediction.TicID, prediction.Label)
			fmt.Printf("  TIC Description: %s\n", prediction.TicDescription)
			fmt.Printf("  Product Description: %s\n", prediction.ProductDescription)
		} else if prediction.Error != nil {
			fmt.Printf("Recommendation failed: %s\n", *prediction.Error)
		} else {
			fmt.Println("Recommendation failed with unknown error")
		}
	}

	// Example 3: Use the TIC with cart line items
	if len(searchResp.Results) > 0 {
		ticID := searchResp.Results[0].TicID
		fmt.Printf("\n=== Using TIC %s ===\n", ticID)
		fmt.Println("Use the TIC as the taxabilityCode in cart line items:")
		fmt.Printf("  CartLineItem{ItemID: \"item-1\", Price: 10.00, Quantity: 1, TaxabilityCode: &tic}\n")
		fmt.Printf("  (where tic is the int64 value parsed from TicID: %s)\n", ticID)
	}
}
