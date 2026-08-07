// Command taxcloud_order demonstrates the deprecated direct TaxCloud integration,
// which calls api.v3.taxcloud.com with a connection ID and TaxCloud API key held on
// the client.
//
// New integrations should use Merchant Management instead: see
// examples/merchant_management and the Migration section of the README. This example
// is kept so existing users of the deprecated path still have working reference code.
//
// The staticcheck exemptions below are deliberate: this file exercises the deprecated
// API on purpose, so SA1019 would otherwise fire on every call.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ziptax/ziptax-go"
	"github.com/ziptax/ziptax-go/models"
)

func main() {
	// Get API keys from environment variables
	ziptaxAPIKey := os.Getenv("ZIPTAX_API_KEY")
	if ziptaxAPIKey == "" {
		log.Fatal("ZIPTAX_API_KEY environment variable is required")
	}

	taxCloudConnectionID := os.Getenv("TAXCLOUD_CONNECTION_ID")
	taxCloudAPIKey := os.Getenv("TAXCLOUD_API_KEY")

	if taxCloudConnectionID == "" || taxCloudAPIKey == "" {
		log.Fatal("TAXCLOUD_CONNECTION_ID and TAXCLOUD_API_KEY environment variables are required for order operations")
	}

	// Create client with TaxCloud credentials
	client, err := ziptax.NewClient(
		ziptaxAPIKey,
		//nolint:staticcheck // demonstrates the deprecated direct-TaxCloud path by design
		ziptax.WithTaxCloudConnectionID(taxCloudConnectionID),
		//nolint:staticcheck // demonstrates the deprecated direct-TaxCloud path by design
		ziptax.WithTaxCloudAPIKey(taxCloudAPIKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Create an order request
	orderReq := &models.CreateOrderRequest{
		OrderID:         "example-order-001",
		CustomerID:      "customer-456",
		TransactionDate: "2026-02-13T09:45:00Z",
		CompletedDate:   "2026-02-13T09:50:00Z",
		Origin: models.TaxCloudAddress{
			Line1: "323 Washington Ave N",
			City:  "Minneapolis",
			State: "MN",
			Zip:   "55401-2427",
		},
		Destination: models.TaxCloudAddress{
			Line1: "200 Spectrum Center Drive Suite 300",
			City:  "Irvine",
			State: "CA",
			Zip:   "92618",
		},
		LineItems: []models.CartItemWithTax{
			{
				Index:    0,
				ItemID:   "item-001",
				Price:    15,
				Quantity: 1,
				Tax: models.Tax{
					Amount: 1.1625,
					Rate:   0.0775,
				},
			},
		},
		Currency: &models.Currency{},
	}

	// Create the order
	fmt.Println("Creating order in TaxCloud...")
	//nolint:staticcheck // demonstrates the deprecated direct-TaxCloud path by design
	response, err := client.CreateOrder(ctx, orderReq)
	if err != nil {
		log.Fatalf("Failed to create order: %v", err)
	}

	// Print the response
	fmt.Printf("✓ Order created successfully!\n")
	fmt.Printf("  Order ID: %s\n", response.OrderID)
	fmt.Printf("  Customer ID: %s\n", response.CustomerID)
	fmt.Printf("  Connection ID: %s\n", response.ConnectionID)
	fmt.Printf("  Transaction Date: %s\n", response.TransactionDate)
	fmt.Printf("  Line Items: %d\n", len(response.LineItems))

	for i, item := range response.LineItems {
		fmt.Printf("    Item %d: %s (Qty: %.2f, Price: $%.2f, Tax: $%.2f)\n",
			i+1, item.ItemID, item.Quantity, item.Price, item.Tax.Amount)
	}
}
