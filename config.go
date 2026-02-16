package ziptax

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the default base URL for the ZipTax API.
	DefaultBaseURL = "https://api.zip-tax.com"

	// DefaultTaxCloudBaseURL is the default base URL for the TaxCloud API.
	DefaultTaxCloudBaseURL = "https://api.v3.taxcloud.com"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default maximum number of retry attempts.
	DefaultMaxRetries = 3

	// DefaultRetryWaitMin is the default minimum wait time between retries.
	DefaultRetryWaitMin = 1 * time.Second

	// DefaultRetryWaitMax is the default maximum wait time between retries.
	DefaultRetryWaitMax = 30 * time.Second

	// APIKeyHeader is the HTTP header name for the API key.
	APIKeyHeader = "X-API-Key"
)

// Config holds the configuration for the ZipTax client.
type Config struct {
	// APIKey is the ZipTax API key for authentication.
	APIKey string

	// BaseURL is the base URL for the ZipTax API.
	// Defaults to DefaultBaseURL if not specified.
	BaseURL string

	// HTTPClient is the HTTP client to use for requests.
	// If nil, a default client will be created.
	HTTPClient *http.Client

	// Timeout is the timeout for HTTP requests.
	// Defaults to DefaultTimeout if not specified.
	Timeout time.Duration

	// MaxRetries is the maximum number of retry attempts for failed requests.
	// Set to 0 to disable retries. Defaults to DefaultMaxRetries.
	MaxRetries int

	// RetryWaitMin is the minimum wait time between retries.
	// Defaults to DefaultRetryWaitMin.
	RetryWaitMin time.Duration

	// RetryWaitMax is the maximum wait time between retries.
	// Defaults to DefaultRetryWaitMax.
	RetryWaitMax time.Duration

	// Logger is an optional logger for request/response logging.
	// If nil, logging is disabled.
	Logger Logger

	// UserAgent is the User-Agent header to send with requests.
	// Defaults to "ziptax-go/{version}".
	UserAgent string

	// TaxCloud configuration (optional - required for order management features)

	// TaxCloudConnectionID is the TaxCloud Connection ID for order operations.
	// If not provided, TaxCloud order features will not be available.
	TaxCloudConnectionID string

	// TaxCloudAPIKey is the TaxCloud API key for authentication.
	// If not provided, TaxCloud order features will not be available.
	TaxCloudAPIKey string

	// TaxCloudBaseURL is the base URL for the TaxCloud API.
	// Defaults to DefaultTaxCloudBaseURL if not specified.
	TaxCloudBaseURL string
}

// Logger is an interface for logging.
type Logger interface {
	// Printf logs a formatted message.
	Printf(format string, v ...interface{})
}

// validate checks that the configuration is valid.
func (c *Config) validate() error {
	if c.APIKey == "" {
		return ErrInvalidAPIKey
	}
	return nil
}

// applyDefaults applies default values to the configuration.
func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}

	if c.Timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	if c.MaxRetries < 0 {
		c.MaxRetries = DefaultMaxRetries
	}

	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = DefaultRetryWaitMin
	}

	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = DefaultRetryWaitMax
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
		}
	}

	if c.UserAgent == "" {
		c.UserAgent = "ziptax-go/" + Version
	}

	// Apply TaxCloud defaults if TaxCloud credentials are provided
	if c.TaxCloudConnectionID != "" && c.TaxCloudAPIKey != "" {
		if c.TaxCloudBaseURL == "" {
			c.TaxCloudBaseURL = DefaultTaxCloudBaseURL
		}
	}
}

// HasTaxCloudCredentials returns true if TaxCloud credentials are configured.
func (c *Config) HasTaxCloudCredentials() bool {
	return c.TaxCloudConnectionID != "" && c.TaxCloudAPIKey != ""
}
