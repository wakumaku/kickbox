package kickbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ResponseVerify kickbox structure on success response (200OK)
// see: https://docs.kickbox.com/docs/single-verification-api#the-response
type ResponseVerify struct {
	Result     string  `json:"result"`       // "undeliverable",
	Reason     string  `json:"reason"`       // "rejected_email",
	Role       bool    `json:"role"`         // false,
	Free       bool    `json:"free"`         // false,
	Disposable bool    `json:"disposable"`   // false,
	AcceptAll  bool    `json:"accept_all"`   // false,
	DidYouMean string  `json:"did_you_mean"` // "bill.lumbergh@gmail.com",
	Sendex     float64 `json:"sendex"`       // 0.23,
	Email      string  `json:"email"`        // "bill.lumbergh@gamil.com",
	User       string  `json:"user"`         // "bill.lumbergh",
	Domain     string  `json:"domain"`       // "gamil.com",
	Success    bool    `json:"success"`      // true,
	Message    string  `json:"message"`      // null
}

// ResponseVerifyHeaders
// see: https://docs.kickbox.com/docs/using-the-api#response-headers
type ResponseVerifyHeaders struct {
	Balance      int // Your remaining verification credit balance
	ResponseTime int // The elapsed time (in milliseconds) it took Kickbox to process the request
	HTTPStatus   int // HTTP Status Response Code
}

// VerifyRequestOptions holds the optional parameters for the Verify request
type VerifyRequestOptions struct {
	timeout time.Duration
}

// VerifyOption option type
type VerifyOption func(*VerifyRequestOptions)

// Timeout Maximum time for the API to complete a verification request.
// Default: 6 seconds, Maximum: 30 seconds
// see: https://docs.kickbox.com/docs/single-verification-api#query-parameters
func Timeout(d time.Duration) VerifyOption {
	return func(o *VerifyRequestOptions) {
		o.timeout = d
	}
}

// Verify calls the verification endpoint
// Optionaly a timeout can be specified
func (c *ClientHTTP) Verify(ctx context.Context, email string, opts ...VerifyOption) (*ResponseVerifyHeaders, *ResponseVerify, error) {
	const verifyPath = "/v2/verify"

	// RateLimiter will block until it is permitted or the context is canceled
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, nil, fmt.Errorf("rate limiting requests: %v", err)
	}

	// MaxConcurrentConnections control
	select {
	case c.connPool <- struct{}{}:
	default:
		return nil, nil, fmt.Errorf("max connections oppened: %d", maxConcurrentConnections)
	}
	defer func() {
		<-c.connPool
	}()

	// Default options
	const defaultRequestTimeout = 6000 * time.Millisecond
	options := VerifyRequestOptions{
		timeout: defaultRequestTimeout,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if options.timeout > 30*time.Second || options.timeout == 0 {
		return nil, nil, fmt.Errorf("timeout not valid, must be less than 30 sec: %v", options.timeout)
	}

	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()

	// Request building ...
	requestURL := c.baseURL + verifyPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("building request: %v", err)
	}

	// Adds query params
	q := req.URL.Query()
	q.Add("email", email)
	q.Add("apikey", c.apiKey)
	q.Add("timeout", fmt.Sprintf("%v", options.timeout.Milliseconds()))
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("doing request: %v", err)
	}
	defer resp.Body.Close()

	// Retrieve extra information in headers
	balance, _ := strconv.Atoi(resp.Header.Get("X-Kickbox-Balance"))
	responseTime, _ := strconv.Atoi(resp.Header.Get("X-Kickbox-Response-Time"))

	header := ResponseVerifyHeaders{
		Balance:      balance,
		ResponseTime: responseTime,
		HTTPStatus:   resp.StatusCode,
	}

	// Parse the the body response
	var body ResponseVerify
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return &header, nil, fmt.Errorf("decoding response: %v", err)
	}

	return &header, &body, nil
}
