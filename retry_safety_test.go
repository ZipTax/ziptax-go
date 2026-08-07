package ziptax

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
)

// Retry safety.
//
// The HTTP client retries on transport errors, 5xx, and 429. That is fine for
// reads and for writes keyed by a caller-supplied identifier, but not for
// operations where each request creates a new record server-side: retrying one
// of those after a timeout can duplicate a refund, a merchant, or an exemption
// certificate, and the caller cannot tell from the returned error that it
// happened. Those operations go through a single-attempt path instead.
//
// These tests pin that behavior by counting requests that actually reach the
// server while the client is configured to retry three times.

// countingServer counts requests and always fails with the given status, so a
// retrying client would keep coming back.
func countingServer(t *testing.T, status int, hits *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"title":"boom","status":500}`))
	}))
}

// retryingClient is configured to retry three times with negligible backoff.
func retryingClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient("test-api-key",
		WithBaseURL(serverURL),
		WithMaxRetries(3),
		WithRetryWait(time.Millisecond, 2*time.Millisecond),
	)
	require.NoError(t, err)
	return client
}

func TestNonIdempotentOperationsAreNotRetried(t *testing.T) {
	tests := []struct {
		name string
		why  string
		call func(context.Context, *Client) error
	}{
		{
			name: "CreateMerchantRefund",
			why:  "a duplicate submission records a duplicate refund",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateMerchantRefund(ctx, &models.MerchantCreateRefundRequest{
					MerchantID: testMerchantID,
					OrderID:    "my-order-1",
					Items:      []models.MerchantRefundRequestItem{{ItemID: "item-1", Quantity: 1}},
				})
				return err
			},
		},
		{
			name: "CreateMerchant",
			why:  "the server assigns the merchantId, so a resubmission creates a second merchant",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateMerchant(ctx, &models.CreateMerchantRequest{
					MerchantName: "Acme Supply Co",
				})
				return err
			},
		},
		{
			name: "CreateMerchantCertificate",
			why:  "the server assigns the certificateId, so a resubmission creates a second certificate",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateMerchantCertificate(ctx, validCertificateRequest())
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits int32
			server := countingServer(t, http.StatusInternalServerError, &hits)
			defer server.Close()

			err := tt.call(context.Background(), retryingClient(t, server.URL))
			require.Error(t, err, "the 500 should surface rather than being swallowed")

			assert.Equal(t, int32(1), atomic.LoadInt32(&hits),
				"%s must reach the server exactly once even with WithMaxRetries(3): %s", tt.name, tt.why)
		})
	}
}

// TestRefundOrderIsNotRetried covers the deprecated direct-TaxCloud refund,
// which hits the same underlying TaxCloud endpoint as CreateMerchantRefund.
func TestRefundOrderIsNotRetried(t *testing.T) {
	var hits int32
	server := countingServer(t, http.StatusInternalServerError, &hits)
	defer server.Close()

	client, err := NewClient("test-api-key",
		WithBaseURL(server.URL),
		WithMaxRetries(3),
		WithRetryWait(time.Millisecond, 2*time.Millisecond),
		//nolint:staticcheck // exercising the deprecated direct-TaxCloud path on purpose
		WithTaxCloudConnectionID("conn-1"),
		//nolint:staticcheck // exercising the deprecated direct-TaxCloud path on purpose
		WithTaxCloudAPIKey("tc-key"),
		//nolint:staticcheck // exercising the deprecated direct-TaxCloud path on purpose
		WithTaxCloudBaseURL(server.URL),
	)
	require.NoError(t, err)

	//nolint:staticcheck // exercising the deprecated direct-TaxCloud path on purpose
	_, err = client.RefundOrder(context.Background(), "order-1", &models.RefundTransactionRequest{})
	require.Error(t, err)

	assert.Equal(t, int32(1), atomic.LoadInt32(&hits),
		"RefundOrder must reach the server exactly once: a repeat records a duplicate refund")
}

// TestIdempotentOperationsStillRetry is the other half of the contract. Reads
// and calculations are explicitly documented as safe to retry, and order
// creation is keyed by a caller-supplied orderId, so a repeat conflicts rather
// than duplicating. Those must keep the configured retry behavior.
func TestIdempotentOperationsStillRetry(t *testing.T) {
	tests := []struct {
		name string
		call func(context.Context, *Client) error
	}{
		{
			name: "CalculateMerchantCart",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CalculateMerchantCart(ctx, validMerchantCartRequest())
				return err
			},
		},
		{
			name: "CreateMerchantOrder",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.CreateMerchantOrder(ctx, validMerchantOrderRequest())
				return err
			},
		},
		{
			name: "GetMerchantOrder",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.GetMerchantOrder(ctx, &models.MerchantGetOrderRequest{
					MerchantID: testMerchantID,
					OrderID:    "my-order-1",
				})
				return err
			},
		},
		{
			name: "GetHealth",
			call: func(ctx context.Context, c *Client) error {
				_, err := c.GetHealth(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hits int32
			server := countingServer(t, http.StatusInternalServerError, &hits)
			defer server.Close()

			err := tt.call(context.Background(), retryingClient(t, server.URL))
			require.Error(t, err)

			assert.Equal(t, int32(4), atomic.LoadInt32(&hits),
				"%s should retry: want 1 initial attempt + 3 retries", tt.name)
		})
	}
}

// TestRetryResendsTheFullBody guards the rewind fix at the SDK level: a retried
// write must carry its payload every time, not just on the first attempt.
func TestRetryResendsTheFullBody(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		n, _ := r.Body.Read(buf)
		bodies = append(bodies, string(buf[:n]))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"title":"boom","status":500}`))
	}))
	defer server.Close()

	_, err := retryingClient(t, server.URL).CalculateMerchantCart(
		context.Background(), validMerchantCartRequest())
	require.Error(t, err)

	require.Len(t, bodies, 4, "want 1 initial attempt + 3 retries")
	for i, body := range bodies {
		assert.Contains(t, body, testMerchantID,
			"attempt %d reached the server with a truncated or empty body: %q", i+1, body)
	}
}
