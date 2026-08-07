package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client wraps an HTTP client with additional functionality.
type Client struct {
	HTTPClient  *http.Client
	BaseURL     string
	APIKey      string
	UserAgent   string
	RetryPolicy *RetryPolicy
	Logger      Logger
}

// Logger is an interface for logging.
type Logger interface {
	Printf(format string, v ...interface{})
}

// NewClient creates a new HTTP client wrapper.
func NewClient(httpClient *http.Client, baseURL, apiKey, userAgent string, retryPolicy *RetryPolicy, logger Logger) *Client {
	return &Client{
		HTTPClient:  httpClient,
		BaseURL:     baseURL,
		APIKey:      apiKey,
		UserAgent:   userAgent,
		RetryPolicy: retryPolicy,
		Logger:      logger,
	}
}

// Get performs a GET request with query parameters.
func (c *Client) Get(ctx context.Context, path string, queryParams map[string]string, result interface{}) error {
	// Build URL with query parameters
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	for key, value := range queryParams {
		if value != "" {
			q.Set(key, value)
		}
	}
	u.RawQuery = q.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	// Log request if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Request: %s %s", req.Method, req.URL.String())
	}

	// Execute request with retry
	resp, err := DoWithRetry(ctx, c.HTTPClient, req, c.RetryPolicy)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Log response if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Response: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse JSON response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// GetWithOptions performs a GET request with a custom base URL and headers.
// This is used for APIs other than the default (e.g., TaxCloud).
func (c *Client) GetWithOptions(ctx context.Context, baseURL, path string, headers map[string]string, result interface{}) error {
	// Build URL
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	// Set custom headers (including authentication)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Log request if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Request: %s %s", req.Method, req.URL.String())
	}

	// Execute request with retry
	resp, err := DoWithRetry(ctx, c.HTTPClient, req, c.RetryPolicy)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Log response if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Response: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp.StatusCode, body)
	}

	// Parse JSON response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// Post performs a POST request with JSON body.
//
// The request is retried according to the client's retry policy. Use PostOnce
// for operations where a repeated submission would create a duplicate record.
func (c *Client) Post(ctx context.Context, baseURL, path string, headers map[string]string, body, result interface{}) error {
	return c.doJSON(ctx, http.MethodPost, baseURL, path, headers, body, result, c.RetryPolicy)
}

// PostOnce performs a POST request with JSON body using exactly one attempt,
// regardless of the client's configured retry policy.
//
// This is for non-idempotent operations, where the server creates a new record
// per request rather than converging on the same state. Retrying those on a
// timeout or 5xx can silently create duplicates that the caller has no way to
// detect from the returned error, so a single attempt and a surfaced error is
// the safer failure mode: the caller can then reconcile by reading current
// state before deciding whether to resubmit.
func (c *Client) PostOnce(ctx context.Context, baseURL, path string, headers map[string]string, body, result interface{}) error {
	return c.doJSON(ctx, http.MethodPost, baseURL, path, headers, body, result, SingleAttempt(c.RetryPolicy))
}

// Patch performs a PATCH request with JSON body.
func (c *Client) Patch(ctx context.Context, baseURL, path string, headers map[string]string, body, result interface{}) error {
	return c.doJSON(ctx, http.MethodPatch, baseURL, path, headers, body, result, c.RetryPolicy)
}

// doJSON performs a request with a JSON body and decodes a JSON response,
// applying the supplied retry policy.
func (c *Client) doJSON(ctx context.Context, method, baseURL, path string, headers map[string]string, body, result interface{}, policy *RetryPolicy) error {
	// Marshal request body to JSON
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	// Build URL
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Create request with body. A *bytes.Reader lets net/http populate
	// Request.GetBody, which DoWithRetry needs to rewind the body per attempt.
	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Set custom headers (including authentication)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Log request if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Request: %s %s", req.Method, req.URL.String())
	}

	// Execute request with retry
	resp, err := DoWithRetry(ctx, c.HTTPClient, req, policy)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Log response if logger is available
	if c.Logger != nil {
		c.Logger.Printf("Response: %d %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code (201 Created is valid for POST requests)
	ok := resp.StatusCode == http.StatusOK ||
		(method == http.MethodPost && resp.StatusCode == http.StatusCreated)
	if !ok {
		return c.handleErrorResponse(resp.StatusCode, respBody)
	}

	// Parse JSON response if result is provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse parses and returns an error from an API error response.
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse as a ZipTax structured error response
	var errResp struct {
		Metadata struct {
			Response struct {
				Code    int    `json:"code"`
				Name    string `json:"name"`
				Message string `json:"message"`
			} `json:"response"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil &&
		errResp.Metadata.Response.Code != 0 {
		return &APIError{
			StatusCode: statusCode,
			Code:       errResp.Metadata.Response.Code,
			Name:       errResp.Metadata.Response.Name,
			Message:    errResp.Metadata.Response.Message,
		}
	}

	// Try to parse as a TaxCloud structured error response
	var tcErrResp struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}

	if err := json.Unmarshal(body, &tcErrResp); err == nil &&
		tcErrResp.Status != 0 {
		return &APIError{
			StatusCode: statusCode,
			Code:       tcErrResp.Status,
			Name:       tcErrResp.Title,
			Message:    tcErrResp.Detail,
		}
	}

	// Fallback to generic error
	return &APIError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
	}
}

// APIError represents an API error response.
type APIError struct {
	StatusCode int
	Code       int
	Name       string
	Message    string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("API error (status %d, code %d): %s - %s",
			e.StatusCode, e.Code, e.Name, e.Message)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}
