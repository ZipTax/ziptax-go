package http

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryPolicy defines the retry configuration.
type RetryPolicy struct {
	MaxRetries int
	MinWait    time.Duration
	MaxWait    time.Duration
	CheckRetry func(*http.Response, error) bool
	Backoff    func(attempt int, min, max time.Duration) time.Duration
}

// DefaultRetryPolicy returns the default retry policy.
func DefaultRetryPolicy(maxRetries int, minWait, maxWait time.Duration) *RetryPolicy {
	return &RetryPolicy{
		MaxRetries: maxRetries,
		MinWait:    minWait,
		MaxWait:    maxWait,
		CheckRetry: DefaultCheckRetry,
		Backoff:    ExponentialBackoff,
	}
}

// DefaultCheckRetry is the default retry check function.
// It retries on network errors and 5xx status codes.
func DefaultCheckRetry(resp *http.Response, err error) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on 5xx errors and 429 (Too Many Requests)
	if resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode >= 500 {
		return true
	}

	return false
}

// ExponentialBackoff calculates an exponential backoff with jitter.
func ExponentialBackoff(attempt int, min, max time.Duration) time.Duration {
	mult := math.Pow(2, float64(attempt))
	wait := time.Duration(mult) * min

	// Add jitter
	jitter := time.Duration(rand.Int63n(int64(wait / 4)))
	wait += jitter

	if wait > max {
		wait = max
	}

	return wait
}

// DoWithRetry executes an HTTP request with retry logic.
func DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, policy *RetryPolicy) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// Check if context is already canceled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Clone the request for retry (in case body needs to be re-read)
		reqClone := req.Clone(ctx)

		resp, err = client.Do(reqClone)

		// Check if we should retry
		if attempt < policy.MaxRetries && policy.CheckRetry(resp, err) {
			// Close response body if present
			if resp != nil {
				_ = resp.Body.Close()
			}

			// Calculate backoff
			wait := policy.Backoff(attempt, policy.MinWait, policy.MaxWait)

			// Wait with context cancellation support
			select {
			case <-time.After(wait):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			continue
		}

		// Either succeeded or shouldn't retry
		break
	}

	return resp, err
}
