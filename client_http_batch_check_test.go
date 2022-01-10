package kickbox_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wakumaku/kickbox"

	"github.com/stretchr/testify/assert"
)

func TestClientHTTPBatchCheck(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected *kickbox.VerifyBatchCheckResponse
	}{
		"starting": {
			input: `{
				"id": 123,
				"status": "starting",
				"success": true,
				"message": null
			  }`,
			expected: &kickbox.VerifyBatchCheckResponse{
				ID:      123,
				Status:  "starting",
				Success: true,
				Message: "",
			},
		},
		"processing": {
			input: `{
				"id": 123,
				"status":"processing",
				"progress":{
				  "deliverable": 1,
				  "undeliverable": 0,
				  "risky": 0,
				  "unknown": 0,
				  "total": 3,
				  "unprocessed": 2
				},
				"success": true,
				"message": null
			  }`,
			expected: &kickbox.VerifyBatchCheckResponse{
				ID:      123,
				Status:  "processing",
				Success: true,
				Message: "",
			},
		},
		"completed": {
			input: `{
				"id": 123,
				"name": "Batch API Process - 05-12-2018-01-58-08",
				"download_url": "https://{{DOWNLOAD_URL_HERE}}",
				"created_at": "2018-05-12T18:58:08.000Z",
				"status": "completed",
				"stats": {
				  "deliverable": 2,
				  "undeliverable": 1,
				  "risky": 0,
				  "unknown": 0,
				  "sendex": 0.35,
				  "addresses": 3
				},
				"error": null,
				"duration": 0,
				"success": true,
				"message": null
			  }`,
			expected: &kickbox.VerifyBatchCheckResponse{
				ID:      123,
				Status:  "completed",
				Success: true,
				Message: "",
			},
		},
		"failed": {
			input: `{
				"id": 123,
				"name": "Batch API Process - 05-12-2018-01-58-08",
				"created_at": "2018-05-12T18:58:08.000Z",
				"status": "failed",
				"error": "Description of error here...",
				"duration": 42,
				"success": true,
				"message": null
			  }`,
			expected: &kickbox.VerifyBatchCheckResponse{
				ID:      123,
				Status:  "failed",
				Success: true,
				Message: "",
			},
		},
	}

	handler := func(rw http.ResponseWriter, r *http.Request) {
		jobID := strings.TrimPrefix(r.URL.Path, "/v2/verify-batch/")
		test, found := tests[jobID]
		if !found {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(test.input))
	}

	svr := httptest.NewServer(http.HandlerFunc(handler))
	defer svr.Close()

	client, err := kickbox.New("apikey", kickbox.OverrideBaseURL(svr.URL))
	assert.Nil(t, err, "unexpected error")

	for name, values := range tests {
		resp, err := client.VerifyBatchCheck(context.TODO(), name)
		assert.Nil(t, err)
		assert.Equal(t, resp.ID, values.expected.ID)
		assert.Equal(t, resp.Status, values.expected.Status)
	}

}
