package http

import (
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
	defer resp.Body.Close()

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

// handleErrorResponse parses and returns an error from an API error response.
func (c *Client) handleErrorResponse(statusCode int, body []byte) error {
	// Try to parse as a structured error response
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