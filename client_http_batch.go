package kickbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResponseVerifyBatch is the structure returned when batch verifying emails
// see: https://docs.kickbox.com/docs/batch-verification-api#example-response
type ResponseVerifyBatch struct {
	ID      int    `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// VerifyRequestOptions holds the optional parameters for the Verify Batch request
type VerifyBatchRequestOptions struct {
	filename string
	callback string
	timeout  time.Duration
}

// VerifyBatchOption option type
type VerifyBatchOption func(*VerifyBatchRequestOptions)

// Filename Allows you to specify a name for your job. Kickbox will name your file this name upon download
// see: https://docs.kickbox.com/docs/batch-verification-api#headers
func Filename(s string) VerifyBatchOption {
	return func(o *VerifyBatchRequestOptions) {
		o.filename = s
	}
}

// Callback URL where Kickbox will POST information when the job is complete.
// see: https://docs.kickbox.com/docs/batch-verification-api#the-batch-verification-callback
func Callback(s string) VerifyBatchOption {
	return func(o *VerifyBatchRequestOptions) {
		o.callback = s
	}
}

// RequestTimeout specify the http client timeout
func RequestTimeout(d time.Duration) VerifyBatchOption {
	return func(o *VerifyBatchRequestOptions) {
		o.timeout = d
	}
}

// VerifyBatch Verify batches of up to 1 million email addresses asynchronously from a single request
// see: https://docs.kickbox.com/docs/batch-verification-api
func (c *ClientHTTP) VerifyBatch(ctx context.Context, file io.ReadCloser, opts ...VerifyBatchOption) (*ResponseVerifyBatch, error) {
	const verifyBatchPath = "/v2/verify-batch"

	// Default options
	const defaultRequestTimeout = 30 * time.Second
	options := VerifyBatchRequestOptions{
		timeout: defaultRequestTimeout,
	}
	for _, apply := range opts {
		apply(&options)
	}

	ctx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()

	requestURL := c.baseURL + verifyBatchPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, file)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	// Adds query params
	q := req.URL.Query()
	q.Add("apikey", c.apiKey)
	req.URL.RawQuery = q.Encode()

	// Adds optional headers
	if options.filename != "" {
		req.Header.Add("X-Kickbox-Filename", options.filename)
	}
	if options.callback != "" {
		req.Header.Add("X-Kickbox-Callback", options.callback)
	}

	req.Header.Add("Content-Type", "text/csv")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %v", err)
	}
	defer resp.Body.Close()

	var response ResponseVerifyBatch
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %v", err)
	}

	return &response, nil
}
