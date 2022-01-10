package kickbox_test

import (
	"errors"
	"net/http"
	"testing"
	"time"
	"wakumaku/kickbox"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestClientOptions(t *testing.T) {
	optionCases := []struct {
		optFnc     kickbox.ClientHTTPOption
		returnsErr bool
		expected   string
	}{
		{
			optFnc:     kickbox.OverrideBaseURL(""),
			returnsErr: true,
			expected:   "baseURL is empty",
		},
		{
			optFnc:     kickbox.OverrideBaseURL("http://baseURL"),
			returnsErr: false,
		},
		{
			optFnc:     kickbox.MaxConcurrentConnections(0),
			returnsErr: true,
			expected:   "max concurrent connection must be greater than zero",
		},
		{
			optFnc:     kickbox.MaxConcurrentConnections(5),
			returnsErr: false,
		},
		{
			optFnc:     kickbox.CustomRateLimiter(nil),
			returnsErr: true,
			expected:   "rate limiter is nil",
		},
		{
			optFnc:     kickbox.CustomRateLimiter(rate.NewLimiter(rate.Every(time.Second), 1)),
			returnsErr: false,
		},
		{
			optFnc:     kickbox.CustomHTTPClient(nil),
			returnsErr: true,
			expected:   "client is nil",
		},
		{
			optFnc:     kickbox.CustomHTTPClient(&http.Client{}),
			returnsErr: false,
		},
	}

	options := kickbox.ClientHTTPOptions{}
	for _, oc := range optionCases {
		err := oc.optFnc(&options)
		if oc.returnsErr {
			assert.EqualError(t, err, oc.expected)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestNewClient(t *testing.T) {
	// Create a client with an empty API Key
	client, err := kickbox.New("")
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "apikey is empty")

	// Create a client with a failing option
	failingOption := func(o *kickbox.ClientHTTPOptions) error {
		return errors.New("expected error")
	}
	client, err = kickbox.New("api_key", failingOption)
	assert.Nil(t, client)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "applying optional settings: expected error")
}
