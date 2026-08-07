package ziptax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
)

// validMerchantCartRequest returns a request that passes client-side validation,
// so each test can mutate the one field it is exercising.
func validMerchantCartRequest() *models.MerchantCalculateCartRequest {
	return &models.MerchantCalculateCartRequest{
		MerchantID: testMerchantID,
		Items: []models.MerchantCart{
			{
				CartID:     "my-cart-1",
				CustomerID: "customer-453",
				Currency:   models.Currency{},
				Origin: models.TaxCloudAddress{
					Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
				},
				Destination: models.TaxCloudAddress{
					Line1: "200 Spectrum Center Dr", City: "Irvine", State: "CA", Zip: "92618",
				},
				LineItems: []models.MerchantCartLineItem{
					{Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5},
				},
			},
		},
	}
}

const taxCloudCartResponseJSON = `{
	"connectionId": "25eb9b97-5acb-492d-b720-c03e79cf715a",
	"transactionDate": "2026-08-01T14:00:00Z",
	"items": [{
		"cartId": "my-cart-1",
		"customerId": "customer-453",
		"deliveredBySeller": false,
		"exemption": {"isExempt": false},
		"currency": {"currencyCode": "USD"},
		"origin": {"line1":"323 Washington Ave N","city":"Minneapolis","state":"MN","zip":"55401","countryCode":"US"},
		"destination": {"line1":"200 Spectrum Center Dr","city":"Irvine","state":"CA","zip":"92618","countryCode":"US"},
		"lineItems": [{
			"index": 0, "itemId": "item-1", "tic": 0,
			"price": 10.75, "originalPrice": 10.75, "quantity": 1.5,
			"tax": {"amount": 1.31, "rate": 0.08125}
		}]
	}]
}`

func TestClient_CalculateMerchantCart(t *testing.T) {
	t.Run("TaxCloud-connected merchant", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/cart/calculate", taxCloudCartResponseJSON, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).CalculateMerchantCart(context.Background(), validMerchantCartRequest())
		require.NoError(t, err)
		assert.False(t, resp.IsSelfManaged())
		assert.Equal(t, "25eb9b97-5acb-492d-b720-c03e79cf715a", resp.ConnectionID)
		require.Len(t, resp.Items, 1)
		assert.Equal(t, 1.31, resp.Items[0].LineItems[0].Tax.Amount)
		require.NotNil(t, resp.Items[0].DeliveredBySeller)
		assert.False(t, *resp.Items[0].DeliveredBySeller)
		assert.Equal(t, testMerchantID, body["merchantId"])
	})

	t.Run("self-managed merchant omits connectionId", func(t *testing.T) {
		const selfManagedJSON = `{
			"items": [{
				"cartId": "my-cart-1",
				"customerId": "customer-453",
				"currency": {"currencyCode": "USD"},
				"origin": {"line1":"323 Washington Ave N","city":"Minneapolis","state":"MN","zip":"55401","countryCode":"US"},
				"destination": {"line1":"200 Spectrum Center Dr","city":"Irvine","state":"CA","zip":"92618","countryCode":"US"},
				"lineItems": [{
					"index": 0, "itemId": "item-1", "tic": null,
					"price": 10.75, "originalPrice": 10.75, "quantity": 1.5,
					"tax": {"amount": 1.31, "rate": 0.08125}
				}]
			}]
		}`
		server := newMerchantTestServer(t, "/merchant/cart/calculate", selfManagedJSON, http.StatusOK, nil)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).CalculateMerchantCart(context.Background(), validMerchantCartRequest())
		require.NoError(t, err)
		assert.True(t, resp.IsSelfManaged())
		assert.Empty(t, resp.ConnectionID)
		require.Len(t, resp.Items, 1)
		assert.Nil(t, resp.Items[0].DeliveredBySeller)
		assert.Nil(t, resp.Items[0].Exemption)
		assert.Nil(t, resp.Items[0].LineItems[0].TIC)
	})

	t.Run("self-managed merchant rejects unsupported fields with 403", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"title":"Forbidden","detail":"self-managed merchant","status":403}`))
		}))
		defer server.Close()

		_, err := newTestClient(t, server.URL).CalculateMerchantCart(context.Background(), validMerchantCartRequest())
		assert.Error(t, err)
		var apiErr *APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(*models.MerchantCalculateCartRequest)
			wantErr string
		}{
			{
				name:    "missing merchant ID",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.MerchantID = "" },
				wantErr: "merchantId is required",
			},
			{
				name:    "no carts",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items = nil },
				wantErr: "at least 1 cart",
			},
			{
				name:    "missing customer ID",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items[0].CustomerID = "" },
				wantErr: "items[0].customerId is required",
			},
			{
				name:    "missing destination city",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items[0].Destination.City = "" },
				wantErr: "items[0].destination.city is required",
			},
			{
				name:    "missing origin zip",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items[0].Origin.Zip = "" },
				wantErr: "items[0].origin.zip is required",
			},
			{
				name:    "no line items",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items[0].LineItems = nil },
				wantErr: "items[0].lineItems must contain at least 1 item",
			},
			{
				name:    "zero quantity",
				mutate:  func(r *models.MerchantCalculateCartRequest) { r.Items[0].LineItems[0].Quantity = 0 },
				wantErr: "quantity must be greater than 0",
			},
		}

		client := newTestClient(t, "https://example.invalid")
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := validMerchantCartRequest()
				tt.mutate(req)
				_, err := client.CalculateMerchantCart(context.Background(), req)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("too many carts", func(t *testing.T) {
		req := validMerchantCartRequest()
		cart := req.Items[0]
		for len(req.Items) <= 100 {
			req.Items = append(req.Items, cart)
		}
		_, err := newTestClient(t, "https://example.invalid").CalculateMerchantCart(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must not exceed 100 carts")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CalculateMerchantCart(context.Background(), nil)
		assert.Error(t, err)
	})
}

const merchantOrderJSON = `{
	"orderId": "my-order-1",
	"customerId": "customer-453",
	"connectionId": "25eb9b97-5acb-492d-b720-c03e79cf715a",
	"transactionDate": "2026-08-01T14:00:00Z",
	"completedDate": "2026-08-02T09:15:00Z",
	"kind": "order",
	"channel": null,
	"deliveredBySeller": false,
	"excludeFromFiling": false,
	"currency": {"currencyCode": "USD"},
	"origin": {"line1":"323 Washington Ave N","city":"Minneapolis","state":"MN","zip":"55401","countryCode":"US"},
	"destination": {"line1":"200 Spectrum Center Dr","city":"Irvine","state":"CA","zip":"92618","countryCode":"US"},
	"lineItems": [{
		"index": 0, "itemId": "item-1", "tic": 0,
		"price": 10.75, "originalPrice": 10.75, "quantity": 1.5,
		"tax": {"amount": 1.31, "rate": 0.08125}
	}]
}`

func TestClient_CreateMerchantOrderFromCart(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/order/create-from-cart", merchantOrderJSON, http.StatusOK, &body)
		defer server.Close()

		order, err := newTestClient(t, server.URL).CreateMerchantOrderFromCart(context.Background(),
			&models.MerchantCreateOrderFromCartRequest{
				MerchantID: testMerchantID,
				CartID:     "my-cart-1",
				OrderID:    "my-order-1",
			})
		require.NoError(t, err)
		assert.Equal(t, "my-order-1", order.OrderID)
		assert.Equal(t, models.OrderKindOrder, order.Kind)
		assert.Nil(t, order.Channel)
		assert.Equal(t, "my-cart-1", body["cartId"])
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *models.MerchantCreateOrderFromCartRequest
			wantErr string
		}{
			{"missing merchant ID", &models.MerchantCreateOrderFromCartRequest{CartID: "c", OrderID: "o"}, "merchantId is required"},
			{"missing cart ID", &models.MerchantCreateOrderFromCartRequest{MerchantID: testMerchantID, OrderID: "o"}, "cartId is required"},
			{"missing order ID", &models.MerchantCreateOrderFromCartRequest{MerchantID: testMerchantID, CartID: "c"}, "orderId is required"},
		}

		client := newTestClient(t, "https://example.invalid")
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := client.CreateMerchantOrderFromCart(context.Background(), tt.request)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CreateMerchantOrderFromCart(context.Background(), nil)
		assert.Error(t, err)
	})
}

func validMerchantOrderRequest() *models.MerchantCreateOrderRequest {
	return &models.MerchantCreateOrderRequest{
		MerchantID:      testMerchantID,
		OrderID:         "my-order-1",
		CustomerID:      "customer-453",
		TransactionDate: "2026-08-01T14:00:00Z",
		CompletedDate:   "2026-08-02T09:15:00Z",
		Currency:        models.Currency{},
		Origin: models.TaxCloudAddress{
			Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
		},
		Destination: models.TaxCloudAddress{
			Line1: "200 Spectrum Center Dr", City: "Irvine", State: "CA", Zip: "92618",
		},
		LineItems: []models.MerchantOrderLineItem{
			{Index: 0, ItemID: "item-1", Price: 10.75, Quantity: 1.5, Tax: models.Tax{Amount: 1.31, Rate: 0.08125}},
		},
	}
}

func TestClient_CreateMerchantOrder(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/order/create", merchantOrderJSON, http.StatusOK, &body)
		defer server.Close()

		order, err := newTestClient(t, server.URL).CreateMerchantOrder(context.Background(), validMerchantOrderRequest())
		require.NoError(t, err)
		assert.Equal(t, "my-order-1", order.OrderID)
		require.Len(t, order.LineItems, 1)
		assert.Equal(t, 0.08125, order.LineItems[0].Tax.Rate)
		assert.Equal(t, "customer-453", body["customerId"])
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(*models.MerchantCreateOrderRequest)
			wantErr string
		}{
			{"missing merchant ID", func(r *models.MerchantCreateOrderRequest) { r.MerchantID = "" }, "merchantId is required"},
			{"missing order ID", func(r *models.MerchantCreateOrderRequest) { r.OrderID = "" }, "orderId is required"},
			{"missing customer ID", func(r *models.MerchantCreateOrderRequest) { r.CustomerID = "" }, "customerId is required"},
			{"missing transaction date", func(r *models.MerchantCreateOrderRequest) { r.TransactionDate = "" }, "transactionDate is required"},
			{"missing completed date", func(r *models.MerchantCreateOrderRequest) { r.CompletedDate = "" }, "completedDate is required"},
			{"no line items", func(r *models.MerchantCreateOrderRequest) { r.LineItems = nil }, "lineItems must contain at least 1 item"},
			{"missing item ID", func(r *models.MerchantCreateOrderRequest) { r.LineItems[0].ItemID = "" }, "lineItems[0].itemId is required"},
		}

		client := newTestClient(t, "https://example.invalid")
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := validMerchantOrderRequest()
				tt.mutate(req)
				_, err := client.CreateMerchantOrder(context.Background(), req)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CreateMerchantOrder(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestClient_GetMerchantOrder(t *testing.T) {
	t.Run("successful request with expand", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/order/get", merchantOrderJSON, http.StatusOK, &body)
		defer server.Close()

		order, err := newTestClient(t, server.URL).GetMerchantOrder(context.Background(), &models.MerchantGetOrderRequest{
			MerchantID: testMerchantID,
			OrderID:    "my-order-1",
			Expand:     models.ExpandRefunds,
		})
		require.NoError(t, err)
		assert.Equal(t, "my-order-1", order.OrderID)
		assert.Equal(t, "refunds", body["expand"])
	})

	t.Run("expand omitted when empty", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/order/get", merchantOrderJSON, http.StatusOK, &body)
		defer server.Close()

		_, err := newTestClient(t, server.URL).GetMerchantOrder(context.Background(), &models.MerchantGetOrderRequest{
			MerchantID: testMerchantID,
			OrderID:    "my-order-1",
		})
		require.NoError(t, err)
		_, present := body["expand"]
		assert.False(t, present)
	})

	t.Run("missing order ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").GetMerchantOrder(context.Background(),
			&models.MerchantGetOrderRequest{MerchantID: testMerchantID})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "orderId is required")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").GetMerchantOrder(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestClient_UpdateMerchantOrder(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/order/update", merchantOrderJSON, http.StatusOK, &body)
		defer server.Close()

		order, err := newTestClient(t, server.URL).UpdateMerchantOrder(context.Background(),
			&models.MerchantUpdateOrderRequest{
				MerchantID:    testMerchantID,
				OrderID:       "my-order-1",
				CompletedDate: "2026-08-03T10:00:00Z",
			})
		require.NoError(t, err)
		assert.Equal(t, "2026-08-02T09:15:00Z", order.CompletedDate)
		assert.Equal(t, "2026-08-03T10:00:00Z", body["completedDate"])
	})

	t.Run("missing merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").UpdateMerchantOrder(context.Background(),
			&models.MerchantUpdateOrderRequest{OrderID: "my-order-1"})
		assert.Error(t, err)
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").UpdateMerchantOrder(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestClient_CreateMerchantRefund(t *testing.T) {
	const refundJSON = `{
		"connectionId": "25eb9b97-5acb-492d-b720-c03e79cf715a",
		"createdDate": "2026-08-05T12:00:00Z",
		"items": [{
			"index": 0, "itemId": "item-1", "price": 10.75, "quantity": 1,
			"tax": {"amount": 0.87}, "tic": 0
		}]
	}`

	t.Run("partial refund", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/refund/create", refundJSON, http.StatusOK, &body)
		defer server.Close()

		refund, err := newTestClient(t, server.URL).CreateMerchantRefund(context.Background(),
			&models.MerchantCreateRefundRequest{
				MerchantID: testMerchantID,
				OrderID:    "my-order-1",
				Items: []models.MerchantRefundRequestItem{
					{ItemID: "item-1", Quantity: 1},
				},
			})
		require.NoError(t, err)
		require.Len(t, refund.Items, 1)
		assert.Equal(t, 0.87, refund.Items[0].Tax.Amount)

		items, ok := body["items"].([]interface{})
		require.True(t, ok)
		assert.Len(t, items, 1)
	})

	t.Run("full refund omits items", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/refund/create", refundJSON, http.StatusOK, &body)
		defer server.Close()

		_, err := newTestClient(t, server.URL).CreateMerchantRefund(context.Background(),
			&models.MerchantCreateRefundRequest{MerchantID: testMerchantID, OrderID: "my-order-1"})
		require.NoError(t, err)
		_, present := body["items"]
		assert.False(t, present)
	})

	t.Run("missing order ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CreateMerchantRefund(context.Background(),
			&models.MerchantCreateRefundRequest{MerchantID: testMerchantID})
		assert.Error(t, err)
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CreateMerchantRefund(context.Background(), nil)
		assert.Error(t, err)
	})
}

const certificateJSON = `{
	"certificateId": "cert-123",
	"connectionId": "25eb9b97-5acb-492d-b720-c03e79cf715a",
	"accountId": 4242,
	"customerId": "customer-453",
	"customerName": "Acme Reseller LLC",
	"customerBusinessType": "RetailTrade",
	"reason": "Resale",
	"reasonDescription": "Resale",
	"createdDate": "2026-08-01T14:00:00Z",
	"disabledAt": null,
	"singlePurchase": false,
	"address": {"line1":"323 Washington Ave N","city":"Minneapolis","state":"MN","zip":"55401","countryCode":"US"},
	"states": [{"abbreviation": "MN"}]
}`

func validCertificateRequest() *models.MerchantCreateCertificateRequest {
	return &models.MerchantCreateCertificateRequest{
		MerchantID:           testMerchantID,
		CustomerID:           "customer-453",
		CustomerName:         "Acme Reseller LLC",
		CustomerBusinessType: models.BusinessTypeRetailTrade,
		Reason:               models.ExemptionReasonResale,
		ReasonDescription:    "Resale",
		Address: models.TaxCloudAddress{
			Line1: "323 Washington Ave N", City: "Minneapolis", State: "MN", Zip: "55401",
		},
		States: []models.CertificateState{{Abbreviation: "MN"}},
	}
}

func TestClient_CreateMerchantCertificate(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/cert/create", certificateJSON, http.StatusOK, &body)
		defer server.Close()

		cert, err := newTestClient(t, server.URL).CreateMerchantCertificate(context.Background(), validCertificateRequest())
		require.NoError(t, err)
		assert.Equal(t, "cert-123", cert.CertificateID)
		assert.True(t, cert.IsActive())
		assert.Equal(t, int64(4242), cert.AccountID)
		assert.Equal(t, "RetailTrade", body["customerBusinessType"])
	})

	t.Run("disabled certificate is not active", func(t *testing.T) {
		disabled := `{"certificateId":"cert-123","disabledAt":"2026-08-04T00:00:00Z"}`
		server := newMerchantTestServer(t, "/merchant/cert/create", disabled, http.StatusOK, nil)
		defer server.Close()

		cert, err := newTestClient(t, server.URL).CreateMerchantCertificate(context.Background(), validCertificateRequest())
		require.NoError(t, err)
		assert.False(t, cert.IsActive())
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			mutate  func(*models.MerchantCreateCertificateRequest)
			wantErr string
		}{
			{"missing merchant ID", func(r *models.MerchantCreateCertificateRequest) { r.MerchantID = "" }, "merchantId is required"},
			{"missing customer ID", func(r *models.MerchantCreateCertificateRequest) { r.CustomerID = "" }, "customerId is required"},
			{"missing customer name", func(r *models.MerchantCreateCertificateRequest) { r.CustomerName = "" }, "customerName is required"},
			{"missing reason", func(r *models.MerchantCreateCertificateRequest) { r.Reason = "" }, "reason is required"},
			{"missing reason description", func(r *models.MerchantCreateCertificateRequest) { r.ReasonDescription = "" }, "reasonDescription is required"},
			{
				name: "reason description too long",
				mutate: func(r *models.MerchantCreateCertificateRequest) {
					r.ReasonDescription = "this description is well over twenty characters"
				},
				wantErr: "must not exceed 20 characters",
			},
			{"missing business type", func(r *models.MerchantCreateCertificateRequest) { r.CustomerBusinessType = "" }, "customerBusinessType is required"},
			{"missing address state", func(r *models.MerchantCreateCertificateRequest) { r.Address.State = "" }, "address.state is required"},
			{"no states", func(r *models.MerchantCreateCertificateRequest) { r.States = nil }, "states must contain at least 1 state"},
			{
				name:    "blank state abbreviation",
				mutate:  func(r *models.MerchantCreateCertificateRequest) { r.States = []models.CertificateState{{}} },
				wantErr: "states[0].abbreviation is required",
			},
		}

		client := newTestClient(t, "https://example.invalid")
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := validCertificateRequest()
				tt.mutate(req)
				_, err := client.CreateMerchantCertificate(context.Background(), req)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").CreateMerchantCertificate(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestClient_GetMerchantCertificate(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/cert/get", certificateJSON, http.StatusOK, &body)
		defer server.Close()

		cert, err := newTestClient(t, server.URL).GetMerchantCertificate(context.Background(), testMerchantID, "cert-123")
		require.NoError(t, err)
		assert.Equal(t, "Acme Reseller LLC", cert.CustomerName)
		assert.Equal(t, "cert-123", body["certificateId"])
	})

	t.Run("missing certificate ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").GetMerchantCertificate(context.Background(), testMerchantID, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "certificateId is required")
	})

	t.Run("missing merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").GetMerchantCertificate(context.Background(), "", "cert-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantId is required")
	})
}

func TestClient_ListMerchantCertificates(t *testing.T) {
	const listJSON = `{
		"items": [` + certificateJSON + `],
		"limit": 20,
		"nextCursor": "cursor-2"
	}`

	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/cert/list", listJSON, http.StatusOK, &body)
		defer server.Close()

		page, err := newTestClient(t, server.URL).ListMerchantCertificates(context.Background(),
			&models.MerchantListCertificatesRequest{
				MerchantID: testMerchantID,
				Limit:      20,
				SortBy:     models.CertificateSortByCreatedDate,
			})
		require.NoError(t, err)
		require.Len(t, page.Items, 1)
		assert.Equal(t, int64(20), page.Limit)
		assert.Equal(t, "cursor-2", page.NextCursor)
		assert.Equal(t, "createdDate", body["sortBy"])
	})

	t.Run("validation errors", func(t *testing.T) {
		tests := []struct {
			name    string
			request *models.MerchantListCertificatesRequest
			wantErr string
		}{
			{"missing merchant ID", &models.MerchantListCertificatesRequest{}, "merchantId is required"},
			{"limit too large", &models.MerchantListCertificatesRequest{MerchantID: testMerchantID, Limit: 500}, "must not exceed 100"},
			{"negative limit", &models.MerchantListCertificatesRequest{MerchantID: testMerchantID, Limit: -1}, "must not be negative"},
			{"bad sort field", &models.MerchantListCertificatesRequest{MerchantID: testMerchantID, SortBy: "name"}, "sortBy must be"},
		}

		client := newTestClient(t, "https://example.invalid")
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := client.ListMerchantCertificates(context.Background(), tt.request)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").ListMerchantCertificates(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestClient_DeleteMerchantCertificate(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/cert/delete", `{"deleted": true}`, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).DeleteMerchantCertificate(context.Background(), testMerchantID, "cert-123")
		require.NoError(t, err)
		assert.Equal(t, true, resp["deleted"])
		assert.Equal(t, "cert-123", body["certificateId"])
	})

	t.Run("missing certificate ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").DeleteMerchantCertificate(context.Background(), testMerchantID, "")
		assert.Error(t, err)
	})
}

// TestLineItemIndicesAreValidatedBeforeSending pins that invalid indices are
// rejected locally. Previously the validator ignored Index entirely, so these
// payloads were sent and only rejected by the API after a round trip.
func TestLineItemIndicesAreValidatedBeforeSending(t *testing.T) {
	cartWith := func(items []models.MerchantCartLineItem) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			req := validMerchantCartRequest()
			req.Items[0].LineItems = items
			_, err := c.CalculateMerchantCart(ctx, req)
			return err
		}
	}
	orderWith := func(items []models.MerchantOrderLineItem) func(context.Context, *Client) error {
		return func(ctx context.Context, c *Client) error {
			req := validMerchantOrderRequest()
			req.LineItems = items
			_, err := c.CreateMerchantOrder(ctx, req)
			return err
		}
	}

	tests := []struct {
		name    string
		call    func(context.Context, *Client) error
		wantErr string
	}{
		{
			name: "cart duplicate index",
			call: cartWith([]models.MerchantCartLineItem{
				{Index: 0, ItemID: "item-1", Price: 10, Quantity: 1},
				{Index: 0, ItemID: "item-2", Price: 20, Quantity: 1},
			}),
			wantErr: "each line item must have a unique index",
		},
		{
			name: "cart negative index",
			call: cartWith([]models.MerchantCartLineItem{
				{Index: -1, ItemID: "item-1", Price: 10, Quantity: 1},
			}),
			wantErr: "index must be between 0 and 500",
		},
		{
			name: "cart index above the maximum",
			call: cartWith([]models.MerchantCartLineItem{
				{Index: 501, ItemID: "item-1", Price: 10, Quantity: 1},
			}),
			wantErr: "index must be between 0 and 500",
		},
		{
			name: "order duplicate index",
			call: orderWith([]models.MerchantOrderLineItem{
				{Index: 4, ItemID: "item-1", Price: 10, Quantity: 1, Tax: models.Tax{Amount: 1, Rate: 0.1}},
				{Index: 4, ItemID: "item-2", Price: 20, Quantity: 1, Tax: models.Tax{Amount: 2, Rate: 0.1}},
			}),
			wantErr: "each line item must have a unique index",
		},
		{
			name: "order negative index",
			call: orderWith([]models.MerchantOrderLineItem{
				{Index: -3, ItemID: "item-1", Price: 10, Quantity: 1, Tax: models.Tax{Amount: 1, Rate: 0.1}},
			}),
			wantErr: "index must be between 0 and 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits int32
			server := countingServer(t, http.StatusOK, &hits)
			defer server.Close()

			err := tt.call(context.Background(), newTestClient(t, server.URL))

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)

			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr, "want a local ValidationError, not an API error")
			assert.Zero(t, atomic.LoadInt32(&hits),
				"the request should be rejected locally, without reaching the server")
		})
	}
}

// TestNonContiguousIndicesAreAccepted guards against over-strict validation:
// the API requires uniqueness within a cart, not a gapless 0..n-1 run.
func TestNonContiguousIndicesAreAccepted(t *testing.T) {
	var hits int32
	server := countingServer(t, http.StatusOK, &hits)
	defer server.Close()

	req := validMerchantCartRequest()
	req.Items[0].LineItems = []models.MerchantCartLineItem{
		{Index: 0, ItemID: "item-1", Price: 10, Quantity: 1},
		{Index: 9, ItemID: "item-2", Price: 20, Quantity: 1},
		{Index: 250, ItemID: "item-3", Price: 30, Quantity: 1},
	}

	// countingServer replies 200 with an error-shaped body, so decoding yields an
	// empty response rather than an error. What matters is that validation let it
	// through to the transport at all.
	_, _ = newTestClient(t, server.URL).CalculateMerchantCart(context.Background(), req)

	assert.Equal(t, int32(1), atomic.LoadInt32(&hits),
		"non-contiguous but unique indices should pass validation and be sent")
}
