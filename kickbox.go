package kickbox

import (
	"context"
	"io"
)

const (
	// see: https://docs.kickbox.com/docs/using-the-api#calling-the-kickbox-api
	BaseURL = "https://api.kickbox.com"

	/*
		If you are utilizing a "EU Only" account (sign in from app.eu.kickbox.com rather than the
		standard URL) then your API calls will need to be directed to the matching endpoint: api.eu.kickbox.com
	*/
	BaseURLEU = "https://api.eu.kickbox.com"

	// see: https://docs.kickbox.com/docs/using-the-api#api-limits
	maxConcurrentConnections = 25
	maxRatePerMinute         = 8000 / 60
)

// Verifier
type Verifier interface {
	Verify(ctx context.Context, email string, ops ...VerifyOption) (*ResponseVerifyHeaders, *ResponseVerify, error)
	VerifyBatch(ctx context.Context, file io.ReadCloser, opts ...VerifyBatchOption) (*ResponseVerifyBatch, error)
	VerifyBatchCheck(ctx context.Context, batchID string) (*VerifyBatchCheckResponse, error)
}
