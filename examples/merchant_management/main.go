// Command merchant_management walks the full Merchant Management lifecycle:
// create a merchant, connect its TaxCloud credentials, calculate a cart, capture
// the cart as an order, mark it shipped, and refund a line item.
//
// Merchant Management is a Private Preview feature. Contact support@zip.tax for access.
//
// Usage:
//
//	export ZIPTAX_API_KEY=your-ziptax-key
//	# Optional: connect an existing TaxCloud account to the new merchant.
//	# Without these, the merchant is created with an invite and the transaction
//	# steps are skipped until the merchant accepts it.
//	export TAXCLOUD_CONNECTION_ID=...
//	export TAXCLOUD_API_KEY=...
//	go run ./examples/merchant_management
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
	apiKey := os.Getenv("ZIPTAX_API_KEY")
	if apiKey == "" {
		log.Fatal("ZIPTAX_API_KEY environment variable is required")
	}

	// Merchant endpoints authenticate with the ZipTax API key alone. A merchant's
	// own TaxCloud credentials are stored server-side, not on the client.
	client, err := ziptax.NewClient(apiKey, ziptax.WithTimeout(60*time.Second))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// These two failures happen before any merchant exists, so exiting here leaves
	// nothing behind. Everything after this point runs inside run, which returns
	// errors instead of exiting.
	if err := run(context.Background(), client); err != nil {
		log.Fatal(err)
	}
}

// run performs the merchant lifecycle and returns an error rather than exiting.
//
// This matters: the example creates a real merchant and registers a deferred
// DeleteMerchant to clean it up. log.Fatal calls os.Exit, which does not run
// deferred functions, so calling it from the lifecycle body would leave an
// orphaned merchant on the account every time a step failed.
func run(ctx context.Context, client *ziptax.Client) (err error) {
	// Step 1: Create a merchant.
	fmt.Println("=== Create Merchant ===")
	created, err := client.CreateMerchant(ctx, &models.CreateMerchantRequest{
		MerchantName: "Acme Supply Co",
		ContactFirst: "Dana",
		ContactLast:  "Reyes",
		ContactEmail: "ops@acme.example",
		ReferenceID:  "seller-42",
		MerchantType: models.MerchantTypeTaxCloud,
	})
	if err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}
	merchantID := created.MerchantID
	fmt.Printf("Merchant %s: %s\n\n", merchantID, created.Message)

	// Clean up the merchant this example created. Every path out of run below is
	// a return, so this always executes.
	defer func() {
		fmt.Println("\n=== Delete Merchant ===")
		if _, delErr := client.DeleteMerchant(ctx, merchantID); delErr != nil {
			// Report the cleanup failure without masking whatever went wrong
			// first: a merchant left behind needs manual attention.
			err = errors.Join(err, fmt.Errorf(
				"failed to delete merchant %s, delete it manually: %w", merchantID, delErr))
			return
		}
		fmt.Printf("Deleted merchant %s\n", merchantID)
	}()

	// Step 2: Read it back and list the account's merchants.
	fmt.Println("=== Get and List Merchants ===")
	merchant, err := client.GetMerchant(ctx, merchantID)
	if err != nil {
		return fmt.Errorf("failed to get merchant: %w", err)
	}
	fmt.Printf("%s (%s) status=%s self-managed=%t\n",
		merchant.MerchantName, merchant.ReferenceID, merchant.Status, merchant.IsSelfManaged())

	merchants, err := client.ListMerchants(ctx)
	if err != nil {
		return fmt.Errorf("failed to list merchants: %w", err)
	}
	fmt.Printf("Account has %d merchant(s)\n\n", len(merchants))

	// Step 3: Connect the merchant's TaxCloud account, if credentials were supplied.
	// A merchant created with an invite instead establishes credentials by accepting it.
	connectionID := os.Getenv("TAXCLOUD_CONNECTION_ID")
	taxCloudKey := os.Getenv("TAXCLOUD_API_KEY")
	if connectionID == "" || taxCloudKey == "" {
		fmt.Println("TAXCLOUD_CONNECTION_ID and TAXCLOUD_API_KEY not set.")
		fmt.Println("Skipping the transaction steps, which need a connected merchant.")
		return nil
	}

	fmt.Println("=== Set Merchant Credentials ===")
	if _, err := client.SetMerchantCredentials(ctx, &models.SetMerchantCredentialsRequest{
		MerchantID:   merchantID,
		ConnectionID: connectionID,
		APIKey:       taxCloudKey,
	}); err != nil {
		return fmt.Errorf("failed to set merchant credentials: %w", err)
	}
	fmt.Printf("Connected merchant %s to TaxCloud\n\n", merchantID)

	origin := models.TaxCloudAddress{
		Line1: "323 Washington Ave N",
		City:  "Minneapolis",
		State: "MN",
		Zip:   "55401",
	}
	destination := models.TaxCloudAddress{
		Line1: "200 Spectrum Center Dr",
		City:  "Irvine",
		State: "CA",
		Zip:   "92618",
	}

	// Step 4: Calculate cart tax. The same request works for both compliance models.
	fmt.Println("=== Calculate Cart Tax ===")
	cart, err := client.CalculateMerchantCart(ctx, &models.MerchantCalculateCartRequest{
		MerchantID: merchantID,
		Items: []models.MerchantCart{
			{
				CartID:      "example-cart-1",
				CustomerID:  "customer-453",
				Currency:    models.Currency{},
				Origin:      origin,
				Destination: destination,
				LineItems: []models.MerchantCartLineItem{
					{Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5},
					{Index: 1, ItemID: "item-2", Price: 24.00, Quantity: 1},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to calculate cart: %w", err)
	}

	for _, item := range cart.Items {
		fmt.Printf("Cart %s (customer %s)\n", item.CartID, item.CustomerID)
		for _, li := range item.LineItems {
			fmt.Printf("  %s: %.2f x %.2f -> tax %.2f at %.5f\n",
				li.ItemID, li.Price, li.Quantity, li.Tax.Amount, li.Tax.Rate)
		}
	}

	if cart.IsSelfManaged() {
		// Self-managed calculation is stateless: this cartId cannot become an order.
		fmt.Println("\nSelf-managed merchant: calculation only, stopping here.")
		return nil
	}
	fmt.Printf("Calculated under TaxCloud connection %s\n\n", cart.ConnectionID)

	// Step 5: Capture the calculated cart as a recorded order.
	fmt.Println("=== Create Order From Cart ===")
	orderID := fmt.Sprintf("example-order-%d", time.Now().Unix())
	order, err := client.CreateMerchantOrderFromCart(ctx, &models.MerchantCreateOrderFromCartRequest{
		MerchantID: merchantID,
		CartID:     cart.Items[0].CartID,
		OrderID:    orderID,
	})
	if err != nil {
		return fmt.Errorf("failed to create order from cart: %w", err)
	}
	fmt.Printf("Recorded order %s (kind %s)\n\n", order.OrderID, order.Kind)

	// Step 6: Mark it shipped. Setting the completed date creates the tax liability.
	fmt.Println("=== Update Order ===")
	updated, err := client.UpdateMerchantOrder(ctx, &models.MerchantUpdateOrderRequest{
		MerchantID:    merchantID,
		OrderID:       orderID,
		CompletedDate: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	fmt.Printf("Order %s completed on %s\n\n", updated.OrderID, updated.CompletedDate)

	// Step 7: Refund a single line item.
	fmt.Println("=== Create Refund ===")
	refund, err := client.CreateMerchantRefund(ctx, &models.MerchantCreateRefundRequest{
		MerchantID: merchantID,
		OrderID:    orderID,
		Items: []models.MerchantRefundRequestItem{
			{ItemID: "item-2", Quantity: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}
	for _, item := range refund.Items {
		fmt.Printf("Refunded %s: %.2f x %.2f, tax %.2f\n",
			item.ItemID, item.Price, item.Quantity, item.Tax.Amount)
	}

	// Step 8: Read the order back with its refunds attached.
	fmt.Println("\n=== Get Order With Refunds ===")
	final, err := client.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
		MerchantID: merchantID,
		OrderID:    orderID,
		Expand:     models.ExpandRefunds,
	})
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	fmt.Printf("Order %s has %d line item(s) and %d refund(s)\n",
		final.OrderID, len(final.LineItems), len(final.Refunds))

	return nil
}
