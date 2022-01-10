package kickbox

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSandboxResponses(t *testing.T) {
	tests := []struct {
		email        string
		expectedBody string
	}{
		{
			email:        "deliverable@example.com",
			expectedBody: sandboxDeliverable,
		},
		{
			email:        "user+deliverable@example.com",
			expectedBody: sandboxDeliverable,
		},
		{
			email:        "undeliverable@example.com",
			expectedBody: sandboxUndeliverable,
		},
		{
			email:        "user+undeliverable@example.com",
			expectedBody: sandboxUndeliverable,
		},
		{
			email:        "invalid-domain@example.com",
			expectedBody: sandboxInvalidDomain,
		},
		{
			email:        "user+invalid-domain@example.com",
			expectedBody: sandboxInvalidDomain,
		},
		{
			email:        "invalid-email@example.com",
			expectedBody: sandboxInvalidEmail,
		},
		{
			email:        "user+invalid-email@example.com",
			expectedBody: sandboxInvalidEmail,
		},
		{
			email:        "invalid-smtp@example.com",
			expectedBody: sandboxInvalidSMTP,
		},
		{
			email:        "user+invalid-smtp@example.com",
			expectedBody: sandboxInvalidSMTP,
		},
		{
			email:        "low-quality@example.com",
			expectedBody: sandboxLowQuality,
		},
		{
			email:        "user+low-quality@example.com",
			expectedBody: sandboxLowQuality,
		},
		{
			email:        "accept-all@example.com",
			expectedBody: sandboxAcceptAll,
		},
		{
			email:        "user+accept-all@example.com",
			expectedBody: sandboxAcceptAll,
		},
		{
			email:        "role@example.com",
			expectedBody: sandboxRole,
		},
		{
			email:        "user+role@example.com",
			expectedBody: sandboxRole,
		},
		{
			email:        "disposable@example.com",
			expectedBody: sandboxDisposable,
		},
		{
			email:        "user+disposable@example.com",
			expectedBody: sandboxDisposable,
		},
		{
			email:        "timeout@example.com",
			expectedBody: sandboxTimeout,
		},
		{
			email:        "user+timeout@example.com",
			expectedBody: sandboxTimeout,
		},
		{
			email:        "unexpected-error@example.com",
			expectedBody: sandboxUnexpectedError,
		},
		{
			email:        "user+unexpected-error@example.com",
			expectedBody: sandboxUnexpectedError,
		},
		{
			email:        "no-connect@example.com",
			expectedBody: sandboxNoConnect,
		},
		{
			email:        "user+no-connect@example.com",
			expectedBody: sandboxNoConnect,
		},
		{
			email:        "unavailable-smtp@example.com",
			expectedBody: sandboxUnavailableSMTP,
		},
		{
			email:        "user+unavailable-smtp@example.com",
			expectedBody: sandboxUnavailableSMTP,
		},
		{
			email:        "insufficient-balance@example.com",
			expectedBody: sandboxInsufficientBalance,
		},
		{
			email:        "user+insufficient-balance@example.com",
			expectedBody: sandboxInsufficientBalance,
		},
	}

	c := NewSandbox()

	for _, ucase := range tests {
		_, resp, err := c.Verify(context.TODO(), ucase.email)
		assert.Nil(t, err, "unexpected error")

		expectedResp := ResponseVerify{}
		_ = json.Unmarshal([]byte(ucase.expectedBody), &expectedResp)
		expectedResp.Email = ucase.email
		userDomain := strings.Split(ucase.email, "@")
		expectedResp.User = userDomain[0]
		expectedResp.Domain = userDomain[1]

		assert.Equal(t, expectedResp, *resp, "responses must match")
	}
}
