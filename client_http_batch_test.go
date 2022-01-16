package kickbox_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wakumaku/kickbox"

	"github.com/stretchr/testify/assert"
)

func TestVerifyBatch(t *testing.T) {
	const csvFilePath = "./testdata/sample.csv"

	expectedResponse := []byte(`{
		"id":123,
		"success":true,
		"message":null
	  }`)

	expectedFileNameHeader := "myfilename.response"
	expectedCallbackHeader := "http://call.me.maybe"

	expectedFileContent, err := ioutil.ReadFile(csvFilePath)
	assert.Nil(t, err)

	handler := func(rw http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Method != http.MethodPut {
			t.Logf("incorrect http verb, expected PUT, got: %s", r.Method)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.Header.Get("X-Kickbox-Filename") != expectedFileNameHeader {
			t.Log("unexpected filename header content")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		if r.Header.Get("X-Kickbox-Callback") != expectedCallbackHeader {
			t.Log("unexpected callback header content")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Check body content
		content, errRead := ioutil.ReadAll(r.Body)
		if errRead != nil {
			t.Logf("cannot read body content: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		if !bytes.Equal(content, expectedFileContent) {
			t.Log("content and expected content is different")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(expectedResponse)
	}

	svr := httptest.NewServer(http.HandlerFunc(handler))
	defer svr.Close()

	c, err := kickbox.New("myapikey", kickbox.OverrideBaseURL(svr.URL))
	assert.Nil(t, err)

	emailsFile, err := os.Open(csvFilePath)
	assert.Nil(t, err)
	defer emailsFile.Close()

	resp, err := c.VerifyBatch(
		context.TODO(),
		emailsFile, kickbox.Filename(expectedFileNameHeader),
		kickbox.Callback(expectedCallbackHeader),
	)
	assert.Nil(t, err)

	var expectedResp kickbox.ResponseVerifyBatch
	err = json.Unmarshal(expectedResponse, &expectedResp)
	assert.Nil(t, err)

	assert.Equal(t, &expectedResp, resp)
}
