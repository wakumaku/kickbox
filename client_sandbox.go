package kickbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// sandbox responses
// see: https://docs.kickbox.com/docs/sandbox-api
const (
	sandboxDeliverable string = `
	{
		"result": "deliverable",
		"reason": "accepted_email",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 1,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`
	sandboxUndeliverable string = `
	{
		"result": "undeliverable",
		"reason": "rejected_email",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`
	sandboxInvalidDomain string = `
	{
		"result": "undeliverable",
		"reason": "invalid_domain",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}	
`
	sandboxInvalidEmail string = `
	{
		"result": "undeliverable",
		"reason": "invalid_email",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxInvalidSMTP string = `
	{
		"result": "undeliverable",
		"reason": "invalid_smtp",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxLowQuality string = `
	{
		"result": "risky",
		"reason": "low_quality",
		"role": false,
		"free": true,
		"disposable": false,
		"accept_all": false,
		"sendex": 0.5,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxAcceptAll = `
	{
		"result": "risky",
		"reason": "low_deliverability",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": true,
		"sendex": 0.7,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxRole = `
	{
		"result": "risky",
		"reason": "low_quality",
		"role": true,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0.7,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`
	sandboxDisposable = `
	{
		"result": "risky",
		"reason": "low_quality",
		"role": false,
		"free": false,
		"disposable": true,
		"accept_all": true,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxTimeout = `
	{
		"result": "unknown",
		"reason": "timeout",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxUnexpectedError = `
	{
		"result": "unknown",
		"reason": "unexpected_error",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxNoConnect = `
	{
		"result": "unknown",
		"reason": "no_connect",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxUnavailableSMTP = `
	{
		"result": "unknown",
		"reason": "unavailable_smtp",
		"role": false,
		"free": false,
		"disposable": false,
		"accept_all": false,
		"sendex": 0,
		"email": "%s",
		"user": "%s",
		"domain": "%s",
		"success": true,
		"message": null
	}
`

	sandboxInsufficientBalance = `
	{
		"success": false,
		"message": "Insufficient balance"
	}
`
)

// ClientSandbox is a client for testing without doing external calls
type ClientSandbox struct {
	matchList map[*regexp.Regexp]string
}

// Ensure Verifier implementation
var _ Verifier = (*ClientSandbox)(nil)

// NewSandbox creates a new sandbox client
func NewSandbox() *ClientSandbox {
	return &ClientSandbox{
		matchList: map[*regexp.Regexp]string{
			regexp.MustCompile(`^deliverable@.+|.+\+deliverable@.+`):                     sandboxDeliverable,
			regexp.MustCompile(`^undeliverable@.+|.+\+undeliverable@.+`):                 sandboxUndeliverable,
			regexp.MustCompile(`^invalid\-domain@.+|.+\+invalid\-domain@.+`):             sandboxInvalidDomain,
			regexp.MustCompile(`^invalid\-email@.+|.+\+invalid\-email@.+`):               sandboxInvalidEmail,
			regexp.MustCompile(`^invalid\-smtp@.+|.+\+invalid\-smtp@.+`):                 sandboxInvalidSMTP,
			regexp.MustCompile(`^low\-quality@.+|.+\+low\-quality@.+`):                   sandboxLowQuality,
			regexp.MustCompile(`^accept\-all@.+|.+\+accept\-all@.+`):                     sandboxAcceptAll,
			regexp.MustCompile(`^role@.+|.+\+role@.+`):                                   sandboxRole,
			regexp.MustCompile(`^disposable@.+|.+\+disposable@.+`):                       sandboxDisposable,
			regexp.MustCompile(`^unexpected\-error@.+|.+\+unexpected\-error@.+`):         sandboxUnexpectedError,
			regexp.MustCompile(`^timeout@.+|.+\+timeout@.+`):                             sandboxTimeout,
			regexp.MustCompile(`^no\-connect@.+|.+\+no\-connect@.+`):                     sandboxNoConnect,
			regexp.MustCompile(`^unavailable\-smtp@.+|.+\+unavailable\-smtp@.+`):         sandboxUnavailableSMTP,
			regexp.MustCompile(`^insufficient\-balance@.+|.+\+insufficient\-balance@.+`): sandboxInsufficientBalance,
		},
	}
}

// Verify returns a response depending on the email pattern to be verified.
// this implementation won't call the kickbox api, it's a local sandbox
// see: https://docs.kickbox.com/docs/sandbox-api
func (c *ClientSandbox) Verify(_ context.Context, email string, _ ...VerifyOption) (*ResponseVerifyHeaders, *ResponseVerify, error) {
	body := sandboxDeliverable // default
	for r, b := range c.matchList {
		if r.MatchString(email) {
			body = b
			break
		}
	}

	resp := ResponseVerify{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, nil, fmt.Errorf("(sandbox) unmarshaling body: %v", err)
	}

	resp.Email = strings.ToLower(email)
	nameDomain := strings.Split(resp.Email, "@")
	if len(nameDomain) == 2 {
		resp.User = nameDomain[0]
		resp.Domain = nameDomain[1]
	}

	headers := ResponseVerifyHeaders{
		Balance:      1,
		ResponseTime: 1,
		HTTPStatus:   http.StatusOK,
	}

	return &headers, &resp, nil
}

// VerifyBatch always returns the same response
func (c *ClientSandbox) VerifyBatch(_ context.Context, _ io.ReadCloser, opts ...VerifyBatchOption) (*ResponseVerifyBatch, error) {
	const ID = 123456
	return &ResponseVerifyBatch{
		ID:      ID,
		Success: true,
	}, nil
}

func (c *ClientSandbox) VerifyBatchCheck(_ context.Context, _ string) (*VerifyBatchCheckResponse, error) {
	return nil, errors.New("not implemented")
}
