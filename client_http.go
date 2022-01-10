package kickbox

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// ClientHTTP kickbox
type ClientHTTP struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	connPool   chan struct{}
	rateLimit  *rate.Limiter
}

// Ensure Verifier implementation
var _ Verifier = (*ClientHTTP)(nil)

// ClientHTTPOptions holds optional values to parametrize the client
type ClientHTTPOptions struct {
	baseURL                  string
	maxConcurrentConnections uint
	httpClient               *http.Client
}

// ClientHTTPOption signature
type ClientHTTPOption func(*ClientHTTPOptions) error

// OverrideBaseURL allows override the main endpoint to run tests agains mock servers
func OverrideBaseURL(baseURL string) ClientHTTPOption {
	return func(o *ClientHTTPOptions) error {
		if baseURL == "" {
			return errors.New("baseURL is empty")
		}
		o.baseURL = baseURL
		return nil
	}
}

// MaxConcurrentConnections sets the number of maximum simultaneous connections to the service
// see: https://docs.kickbox.com/docs/using-the-api#api-limits
func MaxConcurrentConnections(num uint) ClientHTTPOption {
	return func(o *ClientHTTPOptions) error {
		if num > maxConcurrentConnections || num <= 0 {
			return fmt.Errorf("max concurrent connections must be between 1 and %d", maxConcurrentConnections)
		}
		o.maxConcurrentConnections = num
		return nil
	}
}

// CustomHTTPClient allows to use a custom http client instead of the default one
func CustomHTTPClient(client *http.Client) ClientHTTPOption {
	return func(o *ClientHTTPOptions) error {
		if client == nil {
			return errors.New("client is nil")
		}
		o.httpClient = client
		return nil
	}
}

// New creates a new kickbox HTTP API client
func New(apiKey string, opts ...ClientHTTPOption) (*ClientHTTP, error) {
	if apiKey == "" {
		return nil, errors.New("apikey is empty")
	}

	// default option values
	options := ClientHTTPOptions{
		baseURL:                  BaseURL,
		maxConcurrentConnections: maxConcurrentConnections,
		httpClient:               &http.Client{Timeout: 30 * time.Second},
	}

	for _, o := range opts {
		if err := o(&options); err != nil {
			return nil, fmt.Errorf("applying optional settings: %v", err)
		}
	}

	return &ClientHTTP{
		apiKey:     apiKey,
		httpClient: options.httpClient,
		baseURL:    options.baseURL,
		connPool:   make(chan struct{}, options.maxConcurrentConnections),
		rateLimit:  rate.NewLimiter(rate.Every(maxRatePerMinute), 1),
	}, nil
}
