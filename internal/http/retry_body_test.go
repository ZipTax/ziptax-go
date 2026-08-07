package http

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// fastPolicy keeps backoff negligible so retry tests stay quick.
func fastPolicy(maxRetries int) *RetryPolicy {
	return DefaultRetryPolicy(maxRetries, time.Millisecond, 2*time.Millisecond)
}

// recordingServer captures the body of every request it receives and always
// replies with the given status.
func recordingServer(t *testing.T, status int, bodies *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		*bodies = append(*bodies, string(b))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{}`))
	}))
}

// TestPostReplaysBodyOnRetry covers a regression: DoWithRetry clones the request
// but Clone copies Body by reference, so once the first attempt drained it every
// later attempt sent 0 bytes while ContentLength still claimed the original
// size. net/http rejected that locally with "ContentLength=N with Body length 0",
// so the retry never reached the server and the real API error was replaced by a
// confusing transport error.
func TestPostReplaysBodyOnRetry(t *testing.T) {
	var bodies []string
	srv := recordingServer(t, http.StatusInternalServerError, &bodies)
	defer srv.Close()

	c := NewClient(srv.Client(), srv.URL, "k", "ua", fastPolicy(3), nil)

	err := c.Post(context.Background(), srv.URL, "/x", nil, map[string]string{"hello": "world"}, nil)
	if err == nil {
		t.Fatal("expected the 500 to surface as an error")
	}

	// The API error must survive, not be replaced by a local body error.
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("want the server's 500 surfaced as *APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}

	if len(bodies) != 4 {
		t.Fatalf("server saw %d attempts, want 4 (1 initial + 3 retries)", len(bodies))
	}
	const want = `{"hello":"world"}`
	for i, got := range bodies {
		if got != want {
			t.Errorf("attempt %d body = %q, want %q", i+1, got, want)
		}
	}
}

// TestPatchReplaysBodyOnRetry covers the same rewind path for PATCH.
func TestPatchReplaysBodyOnRetry(t *testing.T) {
	var bodies []string
	srv := recordingServer(t, http.StatusInternalServerError, &bodies)
	defer srv.Close()

	c := NewClient(srv.Client(), srv.URL, "k", "ua", fastPolicy(2), nil)
	_ = c.Patch(context.Background(), srv.URL, "/x", nil, map[string]string{"a": "b"}, nil)

	if len(bodies) != 3 {
		t.Fatalf("server saw %d attempts, want 3", len(bodies))
	}
	for i, got := range bodies {
		if got != `{"a":"b"}` {
			t.Errorf("attempt %d body = %q, want %q", i+1, got, `{"a":"b"}`)
		}
	}
}

// TestPostOnceDoesNotRetry is the guarantee non-idempotent operations rely on:
// exactly one request reaches the server no matter what the policy says.
func TestPostOnceDoesNotRetry(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"server error", http.StatusInternalServerError},
		{"bad gateway", http.StatusBadGateway},
		{"rate limited", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodies []string
			srv := recordingServer(t, tt.status, &bodies)
			defer srv.Close()

			// A policy that would otherwise retry three times.
			c := NewClient(srv.Client(), srv.URL, "k", "ua", fastPolicy(3), nil)

			err := c.PostOnce(context.Background(), srv.URL, "/x", nil, map[string]string{"a": "b"}, nil)
			if err == nil {
				t.Fatal("expected an error")
			}
			if len(bodies) != 1 {
				t.Fatalf("server saw %d requests, want exactly 1: a duplicate record may have been created", len(bodies))
			}
		})
	}
}

// TestPostOnceStillSendsBody guards against the single-attempt path regressing
// into sending an empty payload.
func TestPostOnceStillSendsBody(t *testing.T) {
	var bodies []string
	srv := recordingServer(t, http.StatusOK, &bodies)
	defer srv.Close()

	c := NewClient(srv.Client(), srv.URL, "k", "ua", fastPolicy(3), nil)
	if err := c.PostOnce(context.Background(), srv.URL, "/x", nil, map[string]string{"a": "b"}, nil); err != nil {
		t.Fatalf("PostOnce: %v", err)
	}
	if len(bodies) != 1 || bodies[0] != `{"a":"b"}` {
		t.Fatalf("bodies = %q, want one %q", bodies, `{"a":"b"}`)
	}
}

func TestSingleAttempt(t *testing.T) {
	t.Run("zeroes MaxRetries without mutating the original", func(t *testing.T) {
		base := fastPolicy(5)
		once := SingleAttempt(base)

		if once.MaxRetries != 0 {
			t.Errorf("MaxRetries = %d, want 0", once.MaxRetries)
		}
		if base.MaxRetries != 5 {
			t.Errorf("the source policy was mutated: MaxRetries = %d, want 5", base.MaxRetries)
		}
		if once.MinWait != base.MinWait || once.MaxWait != base.MaxWait {
			t.Error("wait settings should carry over")
		}
	})

	t.Run("nil policy", func(t *testing.T) {
		once := SingleAttempt(nil)
		if once == nil {
			t.Fatal("want a usable policy, got nil")
		}
		if once.MaxRetries != 0 {
			t.Errorf("MaxRetries = %d, want 0", once.MaxRetries)
		}
		if once.CheckRetry == nil || once.Backoff == nil {
			t.Error("want CheckRetry and Backoff populated so DoWithRetry can run")
		}
	})
}

// asAPIError is a local errors.As helper kept small to avoid another import
// in this file's assertions.
func asAPIError(err error, target **APIError) bool {
	for err != nil {
		if e, ok := err.(*APIError); ok {
			*target = e
			return true
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}
