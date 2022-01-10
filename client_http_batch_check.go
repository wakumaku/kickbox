package kickbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// VeryfyBatchCheckResponse
// see: https://docs.kickbox.com/docs/batch-verification-api#example-responses
type VerifyBatchCheckResponse struct {
	ID      int    `json:"id"`      // 123,
	Status  string `json:"status"`  // "starting", "processing", "completed", "failed",
	Success bool   `json:"success"` // true,
	Message string `json:"message"` // null

	// when Status is "processing"
	Progress struct {
		Deliverable   int `json:"deliverable"`   // 1,
		Undeliverable int `json:"undeliverable"` // 0,
		Risky         int `json:"risky"`         // 0,
		Unknown       int `json:"unknown"`       // 0,
		Total         int `json:"total"`         // 3,
		Unprocessed   int `json:"unprocessed"`   // 2
	} `json:"progress"`

	// when Status is completed
	DownloadURL string `json:"download_url"` // "https://{{DOWNLOAD_URL_HERE}}",
	Stats       struct {
		Deliverable   int     `json:"deliverable"`   // 2,
		Undeliverable int     `json:"undeliverable"` // 1,
		Risky         int     `json:"risky"`         // 0,
		Unknown       int     `json:"unknown"`       // 0,
		Sendex        float64 `json:"sendex"`        // 0.35,
		Addresses     int     `json:"addresses"`     // 3
	} `json:"stats"`

	// when Status is completed OR failed
	Name      string `json:"name"`       // "Batch API Process - 05-12-2018-01-58-08",
	CreatedAt string `json:"created_at"` // "2018-05-12T18:58:08.000Z",
	Error     string `json:"error"`      //: null,
	Duration  int    `json:"duration"`   //: 0,
}

// VerifyBatchCheck Checking a Batch Verification Status
// see: https://docs.kickbox.com/docs/batch-verification-api#checking-a-batch-verification-status
func (c *ClientHTTP) VerifyBatchCheck(ctx context.Context, batchID string) (*VerifyBatchCheckResponse, error) {
	const verifyPath = "/v2/verify-batch/"

	if batchID == "" {
		return nil, errors.New("batch id is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Request building ...
	requestURL := c.baseURL + verifyPath + batchID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %v", err)
	}

	// Adds query params
	q := req.URL.Query()
	q.Add("apikey", c.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the the body response
	var body VerifyBatchCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %v", err)
	}

	return &body, nil

}
