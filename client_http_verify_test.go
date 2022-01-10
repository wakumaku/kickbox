package kickbox_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/wakumaku/kickbox"

	"github.com/stretchr/testify/assert"
)

func TestVerifyMaxTimeout(t *testing.T) {

	client, err := kickbox.New("apikey")
	if err != nil {
		assert.NotNil(t, err, "unexpected error")
	}

	_, _, err = client.Verify(context.TODO(), "myemail@example.com", kickbox.Timeout(60*time.Second))
	assert.NotNil(t, err)
	assert.EqualError(t, err, "timeout not valid, must be less than 30 sec: 1m0s")

	_, _, err = client.Verify(context.TODO(), "myemail@example.com", kickbox.Timeout(0))
	assert.NotNil(t, err)
	assert.EqualError(t, err, "timeout not valid, must be less than 30 sec: 0s")
}

func TestVerifyRateLimitError(t *testing.T) {
	client, err := kickbox.New("apikey")
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel the context!
	_, _, err = client.Verify(ctx, "email@example.com")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "rate limiting requests: context canceled")
}

func TestVerifyBuildingRequestError(t *testing.T) {
	client, err := kickbox.New("apikey", kickbox.OverrideBaseURL(":::"))
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, err = client.Verify(ctx, "email@example.com")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "building request: parse \":::/v2/verify\": missing protocol scheme")
}

func TestVerifyRequestError(t *testing.T) {
	client, err := kickbox.New("apikey",
		kickbox.OverrideBaseURL("http://nonexistinghost.test.me"),
	)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, err = client.Verify(ctx, "email@example.com")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "doing request: Get \"http://nonexistinghost.test.me/v2/verify?apikey=apikey&email=email%40example.com&timeout=6000\": dial tcp: lookup nonexistinghost.test.me: no such host")
}

func TestVerifyRequestBodyBroken(t *testing.T) {
	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{broken json`))
	}

	svr := httptest.NewServer(http.HandlerFunc(handler))
	defer svr.Close()

	client, err := kickbox.New("apikey",
		kickbox.OverrideBaseURL(svr.URL),
	)
	assert.Nil(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, _, err = client.Verify(ctx, "email@example.com")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "decoding response: invalid character 'b' looking for beginning of object key string")
}

func TestVerifyMockResponse(t *testing.T) {

	body := []byte(`{
		"result":"undeliverable",
		"reason":"rejected_email",
		"role":false,
		"free":false,
		"disposable":false,
		"accept_all":false,
		"did_you_mean":"bill.lumbergh@gmail.com",
		"sendex":0.23,
		"email":"bill.lumbergh@gamil.com",
		"user":"bill.lumbergh",
		"domain":"gamil.com",
		"success":true,
		"message":null
  	}`)

	handler := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write(body)
	}

	svr := httptest.NewServer(http.HandlerFunc(handler))
	defer svr.Close()

	client, err := kickbox.New("apikey", kickbox.OverrideBaseURL(svr.URL))
	assert.Nil(t, err, "unexpected error")

	_, resp, err := client.Verify(context.TODO(), "email@example.com")
	assert.Nil(t, err, "unexpected error")

	var expectedResponse kickbox.ResponseVerify
	err = json.Unmarshal(body, &expectedResponse)
	assert.Nil(t, err)
	assert.Equal(t, &expectedResponse, resp, "must be equals")
}

func TestVerifyMaxConcurrentConnections(t *testing.T) {

	handler := func(rw http.ResponseWriter, r *http.Request) {
		// delay the response
		time.Sleep(3 * time.Second)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(`{}`))
	}

	svr := httptest.NewServer(http.HandlerFunc(handler))
	defer svr.Close()

	client, err := kickbox.New("apikey",
		kickbox.OverrideBaseURL(svr.URL),
		kickbox.MaxConcurrentConnections(25),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errCtrl := make(chan struct{})
	totalErrors := 0
	go func() {
		for range errCtrl {
			totalErrors++
		}
	}()

	wg := sync.WaitGroup{}
	// Launch 26 concurrent request, only 1 must fail
	for i := 0; i < 26; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := client.Verify(context.TODO(), "email@example.com")
			if err != nil {
				// on error send a signal to count the total number
				errCtrl <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(errCtrl)

	if totalErrors != 1 {
		t.Errorf("expecting only 1 error, got: %d", totalErrors)
	}
}
