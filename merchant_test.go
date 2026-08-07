package ziptax

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziptax/ziptax-go/models"
)

const testMerchantID = "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e"

// newMerchantTestServer returns a server that asserts the request path and API key,
// captures the decoded request body, and replies with the supplied JSON.
func newMerchantTestServer(t *testing.T, wantPath, responseJSON string, status int, captured *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, wantPath, r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		if captured != nil && r.Body != nil {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				*captured = body
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(responseJSON))
	}))
}

func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient("test-api-key", WithBaseURL(serverURL))
	require.NoError(t, err)
	return client
}

const merchantOperationJSON = `{
	"merchantId": "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
	"message": "Merchant created",
	"status": "success"
}`

func TestClient_CreateMerchant(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/create", merchantOperationJSON, http.StatusCreated, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).CreateMerchant(context.Background(), &models.CreateMerchantRequest{
			MerchantName: "Acme Supply Co",
			ContactEmail: "ops@acme.example",
			MerchantType: models.MerchantTypeSelfManaged,
		})
		require.NoError(t, err)
		assert.Equal(t, testMerchantID, resp.MerchantID)
		assert.Equal(t, "success", resp.Status)

		// merchant_type is snake_case on the wire, unlike its camelCase siblings.
		assert.Equal(t, "Acme Supply Co", body["merchantName"])
		assert.Equal(t, "self-managed", body["merchant_type"])
	})

	t.Run("sendTaxcloudInvite uses the API spelling", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/create", merchantOperationJSON, http.StatusCreated, &body)
		defer server.Close()

		invite := true
		_, err := newTestClient(t, server.URL).CreateMerchant(context.Background(), &models.CreateMerchantRequest{
			MerchantName:       "Acme Supply Co",
			SendTaxCloudInvite: &invite,
		})
		require.NoError(t, err)
		assert.Equal(t, true, body["sendTaxcloudInvite"])
	})

	t.Run("nil request", func(t *testing.T) {
		client := newTestClient(t, "https://example.invalid")
		_, err := client.CreateMerchant(context.Background(), nil)
		assert.Error(t, err)
		var validationErr *ValidationError
		assert.ErrorAs(t, err, &validationErr)
	})

	t.Run("missing merchant name", func(t *testing.T) {
		client := newTestClient(t, "https://example.invalid")
		_, err := client.CreateMerchant(context.Background(), &models.CreateMerchantRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantName is required")
	})

	t.Run("invalid merchant type", func(t *testing.T) {
		client := newTestClient(t, "https://example.invalid")
		_, err := client.CreateMerchant(context.Background(), &models.CreateMerchantRequest{
			MerchantName: "Acme Supply Co",
			MerchantType: "not-a-model",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchant_type must be")
	})

	t.Run("API error is surfaced as APIError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"title":"Conflict","detail":"merchant already exists","status":409}`))
		}))
		defer server.Close()

		_, err := newTestClient(t, server.URL).CreateMerchant(context.Background(), &models.CreateMerchantRequest{
			MerchantName: "Acme Supply Co",
		})
		assert.Error(t, err)
		var apiErr *APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, http.StatusConflict, apiErr.StatusCode)
	})
}

func TestClient_UpdateMerchant(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/update", merchantOperationJSON, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).UpdateMerchant(context.Background(), &models.UpdateMerchantRequest{
			MerchantID: testMerchantID,
			Update: models.MerchantUpdate{
				MerchantName: "Acme Supply Co",
				ContactEmail: "billing@acme.example",
			},
		})
		require.NoError(t, err)
		assert.Equal(t, testMerchantID, resp.MerchantID)

		update, ok := body["update"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "billing@acme.example", update["contactEmail"])
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").UpdateMerchant(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("missing merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").UpdateMerchant(context.Background(), &models.UpdateMerchantRequest{
			Update: models.MerchantUpdate{MerchantName: "Acme"},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantId is required")
	})

	t.Run("missing nested merchant name", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").UpdateMerchant(context.Background(), &models.UpdateMerchantRequest{
			MerchantID: testMerchantID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantName is required")
	})
}

func TestClient_DeleteMerchant(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/delete", merchantOperationJSON, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).DeleteMerchant(context.Background(), testMerchantID)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, testMerchantID, body["merchantId"])
	})

	t.Run("empty merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").DeleteMerchant(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantId is required")
	})
}

func TestClient_GetMerchant(t *testing.T) {
	const merchantJSON = `{
		"merchantId": "b1e0a5c2-0f3d-4a51-9f8e-3f2c1d4b5a6e",
		"merchantName": "Acme Supply Co",
		"contactEmail": "ops@acme.example",
		"referenceId": "seller-42",
		"status": "taxcloud_connected"
	}`

	t.Run("successful request", func(t *testing.T) {
		server := newMerchantTestServer(t, "/merchant/get", merchantJSON, http.StatusOK, nil)
		defer server.Close()

		merchant, err := newTestClient(t, server.URL).GetMerchant(context.Background(), testMerchantID)
		require.NoError(t, err)
		assert.Equal(t, "Acme Supply Co", merchant.MerchantName)
		assert.Equal(t, models.MerchantStatusTaxCloudConnected, merchant.Status)
		assert.False(t, merchant.IsSelfManaged())
	})

	t.Run("self-managed merchant", func(t *testing.T) {
		server := newMerchantTestServer(t, "/merchant/get",
			`{"merchantId":"m1","merchantName":"Solo","status":"external_compliance"}`, http.StatusOK, nil)
		defer server.Close()

		merchant, err := newTestClient(t, server.URL).GetMerchant(context.Background(), testMerchantID)
		require.NoError(t, err)
		assert.True(t, merchant.IsSelfManaged())
	})

	t.Run("empty merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").GetMerchant(context.Background(), "")
		assert.Error(t, err)
	})
}

func TestClient_ListMerchants(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/merchant/list", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"merchantId":"m1","merchantName":"Acme","status":"taxcloud_connected"},
				{"merchantId":"m2","merchantName":"Solo","status":"external_compliance"}
			]`))
		}))
		defer server.Close()

		merchants, err := newTestClient(t, server.URL).ListMerchants(context.Background())
		require.NoError(t, err)
		require.Len(t, merchants, 2)
		assert.Equal(t, "Acme", merchants[0].MerchantName)
		assert.True(t, merchants[1].IsSelfManaged())
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		}))
		defer server.Close()

		merchants, err := newTestClient(t, server.URL).ListMerchants(context.Background())
		require.NoError(t, err)
		assert.Empty(t, merchants)
	})
}

func TestClient_SetMerchantCredentials(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/credentials/set", merchantOperationJSON, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).SetMerchantCredentials(context.Background(), &models.SetMerchantCredentialsRequest{
			MerchantID:   testMerchantID,
			ConnectionID: "25eb9b97-5acb-492d-b720-c03e79cf715a",
			APIKey:       "tc-key",
		})
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "25eb9b97-5acb-492d-b720-c03e79cf715a", body["connectionId"])
	})

	t.Run("validation error does not leak the API key", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").SetMerchantCredentials(context.Background(),
			&models.SetMerchantCredentialsRequest{
				MerchantID: testMerchantID,
				APIKey:     "super-secret-key",
			})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connectionId is required")
		assert.NotContains(t, err.Error(), "super-secret-key")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").SetMerchantCredentials(context.Background(), nil)
		assert.Error(t, err)
	})

	t.Run("missing merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").SetMerchantCredentials(context.Background(),
			&models.SetMerchantCredentialsRequest{ConnectionID: "c", APIKey: "k"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "merchantId is required")
	})
}

func TestClient_DeleteMerchantCredentials(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		var body map[string]interface{}
		server := newMerchantTestServer(t, "/merchant/credentials/delete", merchantOperationJSON, http.StatusOK, &body)
		defer server.Close()

		resp, err := newTestClient(t, server.URL).DeleteMerchantCredentials(context.Background(), testMerchantID)
		require.NoError(t, err)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, testMerchantID, body["merchantId"])
	})

	t.Run("empty merchant ID", func(t *testing.T) {
		_, err := newTestClient(t, "https://example.invalid").DeleteMerchantCredentials(context.Background(), "")
		assert.Error(t, err)
	})
}
