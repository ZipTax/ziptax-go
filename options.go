package ziptax

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client.
type Option func(*Config)

// WithBaseURL sets the base URL for the API.
//
// Example:
//
//	client := ziptax.NewClient(apiKey, ziptax.WithBaseURL("https://custom-api.example.com"))
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client.
//
// Example:
//
//	httpClient := &http.Client{Timeout: 60 * time.Second}
//	client := ziptax.NewClient(apiKey, ziptax.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = httpClient
	}
}

// WithTimeout sets the HTTP request timeout.
//
// Example:
//
//	client := ziptax.NewClient(apiKey, ziptax.WithTimeout(60*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.Timeout = timeout
		if c.HTTPClient != nil {
			c.HTTPClient.Timeout = timeout
		}
	}
}

// WithMaxRetries sets the maximum number of retry attempts for failed requests.
// Set to 0 to disable retries.
//
// Example:
//
//	client := ziptax.NewClient(apiKey, ziptax.WithMaxRetries(5))
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithRetryWait sets the minimum and maximum wait times between retries.
//
// Example:
//
//	client := ziptax.NewClient(apiKey,
//		ziptax.WithRetryWait(2*time.Second, 60*time.Second))
func WithRetryWait(min, max time.Duration) Option {
	return func(c *Config) {
		c.RetryWaitMin = min
		c.RetryWaitMax = max
	}
}

// WithLogger sets a logger for request/response logging.
//
// Example:
//
//	logger := log.New(os.Stdout, "[ziptax] ", log.LstdFlags)
//	client := ziptax.NewClient(apiKey, ziptax.WithLogger(logger))
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithUserAgent sets a custom User-Agent header.
//
// Example:
//
//	client := ziptax.NewClient(apiKey, ziptax.WithUserAgent("my-app/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

// WithTaxCloudConnectionID sets the TaxCloud Connection ID for order management features.
// This must be provided along with WithTaxCloudAPIKey to enable TaxCloud order operations.
//
// Deprecated: the direct TaxCloud integration is superseded by Merchant Management,
// which reaches TaxCloud through the ZipTax API using only the ZipTax API key. Store a
// merchant's TaxCloud credentials server-side with Client.SetMerchantCredentials and
// address them by merchant ID instead. See the Migration section of the README.
//
// Example:
//
//	client := ziptax.NewClient(apiKey,
//		ziptax.WithTaxCloudConnectionID("25eb9b97-5acb-492d-b720-c03e79cf715a"),
//		ziptax.WithTaxCloudAPIKey("your-taxcloud-api-key"))
func WithTaxCloudConnectionID(connectionID string) Option {
	return func(c *Config) {
		c.TaxCloudConnectionID = connectionID
	}
}

// WithTaxCloudAPIKey sets the TaxCloud API key for order management features.
// This must be provided along with WithTaxCloudConnectionID to enable TaxCloud order operations.
//
// Deprecated: the direct TaxCloud integration is superseded by Merchant Management,
// which reaches TaxCloud through the ZipTax API using only the ZipTax API key. Store a
// merchant's TaxCloud credentials server-side with Client.SetMerchantCredentials and
// address them by merchant ID instead. See the Migration section of the README.
//
// Example:
//
//	client := ziptax.NewClient(apiKey,
//		ziptax.WithTaxCloudConnectionID("25eb9b97-5acb-492d-b720-c03e79cf715a"),
//		ziptax.WithTaxCloudAPIKey("your-taxcloud-api-key"))
func WithTaxCloudAPIKey(apiKey string) Option {
	return func(c *Config) {
		c.TaxCloudAPIKey = apiKey
	}
}

// WithTaxCloudBaseURL sets a custom base URL for the TaxCloud API.
// Only needed if using a custom TaxCloud endpoint.
//
// Deprecated: the direct TaxCloud integration is superseded by Merchant Management,
// which reaches TaxCloud through the ZipTax API. Merchant endpoints are served from
// the ZipTax base URL, which WithBaseURL sets. See the Migration section of the README.
//
// Example:
//
//	client := ziptax.NewClient(apiKey,
//		ziptax.WithTaxCloudBaseURL("https://custom-taxcloud-api.example.com"))
func WithTaxCloudBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.TaxCloudBaseURL = baseURL
	}
}
