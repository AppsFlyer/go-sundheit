package checks

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	longRequest       = "LongRequest"
	expectedContent   = "I'm healthy"
	testCookieKey     = "test-cookie"
	expectedCookieVal = "test-cookie-val"
	testHeaderKey     = "test-header"
	expectedHeaderVal = "test-value"
)

type receivedRequest struct {
	sync.RWMutex
	details map[string]string
}

func (req *receivedRequest) addDetail(key string, val string) {
	if val == "" {
		return
	}
	req.Lock()
	defer req.Unlock()
	req.details[key] = val
}

func (req *receivedRequest) getDetail(key string) string {
	req.RLock()
	defer req.RUnlock()
	return req.details[key]
}

func (req *receivedRequest) clear() {
	req.Lock()
	defer req.Unlock()
	req.details = make(map[string]string, 2)
}

func TestHTTPCheckName(t *testing.T) {
	name := "http-check"
	check, err := NewHTTPCheck(HTTPCheckConfig{
		CheckName: name,
		URL:       "http://example.org",
	})
	assert.NoError(t, err)
	assert.Equal(t, name, check.Name())
}

func TestNewHttpCheckRequiredFields(t *testing.T) {
	check, err := NewHTTPCheck(HTTPCheckConfig{
		CheckName: "meh",
	})
	assert.Nil(t, check, "nil URL should yield nil check")
	assert.Error(t, err, "nil URL should yield error")

	check, err = NewHTTPCheck(HTTPCheckConfig{
		URL: "http://example.org",
	})
	assert.Nil(t, check, "nil CheckName should yield nil check")
	assert.Error(t, err, "nil CheckName should yield error")

	check, err = NewHTTPCheck(HTTPCheckConfig{
		URL:       ":/invalid.url",
		CheckName: "invalid.url",
	})
	assert.Nil(t, check, "invalid url should yield nil check")
	assert.Error(t, err, "invalid url should yield error")
}

func TestNewHttpCheck(t *testing.T) {
	receivedDetails := receivedRequest{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), longRequest) {
			waitDuration, err := time.ParseDuration(req.URL.Query().Get("wait"))
			if err != nil {
				fmt.Println("Failed to parse sleep duration: ", err)
				waitDuration = 10 * time.Second
			}

			time.Sleep(waitDuration)
		}

		receivedDetails.clear()
		receivedDetails.addDetail(testHeaderKey, req.Header.Get(testHeaderKey))
		if cookie, err := req.Cookie(testCookieKey); err == nil {
			receivedDetails.addDetail(testCookieKey, cookie.Value)
		}

		rw.WriteHeader(200)

		reqBody, _ := ioutil.ReadAll(req.Body)
		responsePayload := expectedContent
		if len(reqBody) != 0 {
			responsePayload = string(reqBody)
		}

		_, err := rw.Write([]byte(responsePayload))
		if err != nil {
			t.Fatal("Failed to write response: ", err)
		}
	}))

	defer server.Close()

	t.Run("HttpCheck success call", testHTTPCheckSuccess(server.URL, server.Client()))
	t.Run("HttpCheck success call with expected body check", testHTTPCheckSuccessWithExpectedBody(server.URL, server.Client()))
	t.Run("HttpCheck success call with POST body payload", testHTTPCheckSuccessWithPostBodyPayload(server.URL, server.Client()))
	t.Run("HttpCheck success call with failing expected body check", testHTTPCheckFailWithUnexpectedBody(server.URL, server.Client()))
	t.Run("HttpCheck success call with options", testHTTPCheckSuccessWithOptions(server.URL, server.Client(), &receivedDetails))
	t.Run("HttpCheck fail on status code", testHTTPCheckFailStatusCode(server.URL, server.Client()))
	t.Run("HttpCheck fail on URL", testHTTPCheckFailURL(server.URL, server.Client()))
	t.Run("HttpCheck fail on timeout", testHTTPCheckFailTimeout(server.URL, server.Client()))
}

func testHTTPCheckSuccess(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName: "url.check",
			URL:       url,
			Client:    client,
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Nil(t, err, "check should pass")
		assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
	}
}

func testHTTPCheckSuccessWithExpectedBody(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName:    "url.check",
			URL:          url,
			Client:       client,
			ExpectedBody: expectedContent,
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Nil(t, err, "check should pass")
		assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
	}
}

func testHTTPCheckSuccessWithPostBodyPayload(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		const postPayload = "body-payload"

		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName:    "url.check",
			URL:          url,
			Client:       client,
			ExpectedBody: postPayload,
			Body:         func() io.Reader { return strings.NewReader(postPayload) },
			Method:       http.MethodPost,
		})
		assert.Nil(t, err)

		for i := 0; i < 5; i++ {
			details, err := check.Execute(context.Background())
			assert.Nil(t, err, "check should pass")
			assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
		}
	}
}

func testHTTPCheckFailWithUnexpectedBody(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName:    "url.check",
			URL:          url,
			Client:       client,
			ExpectedBody: "my body is a temple",
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Error(t, err, "check should fail")
		assert.Equal(t, "body does not contain expected content 'my body is a temple'", err.Error(), "check error message")
		assert.Equal(t, url, details, "check details when fail are the URL")
	}
}

func testHTTPCheckFailStatusCode(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName:      "url.check",
			URL:            url,
			Client:         client,
			ExpectedStatus: 300,
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Error(t, err, "check should fail")
		assert.Equal(t, "unexpected status code: '200' expected: '300'", err.Error(), "check error message")
		assert.Equal(t, url, details, "check details when fail are the URL")
	}
}

func testHTTPCheckSuccessWithOptions(url string, client *http.Client, rr *receivedRequest) func(t *testing.T) {

	return func(t *testing.T) {
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName: "url.check",
			URL:       url,
			Client:    client,
			Options: []RequestOption{
				func(r *http.Request) {
					r.Header.Add(testHeaderKey, expectedHeaderVal)
				},
				func(r *http.Request) {
					r.AddCookie(&http.Cookie{Name: testCookieKey, Value: expectedCookieVal})
				},
			},
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Nil(t, err, "check should pass")
		assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
		assert.Equal(t, expectedCookieVal, rr.getDetail(testCookieKey))
		assert.Equal(t, expectedHeaderVal, rr.getDetail(testHeaderKey))
	}
}

func testHTTPCheckFailURL(_ string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		bogusURL := "http://devil-dot-com:666"
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName: "url.check",
			URL:       bogusURL,
			Client:    client,
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Error(t, err, "check should fail")
		assert.Contains(t, err.Error(), "lookup", "check error message")
		assert.Equal(t, bogusURL, details, "check details when fail are the URL")
	}
}

func testHTTPCheckFailTimeout(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		waitURL := fmt.Sprintf("%s/%s?wait=%s", url, longRequest, 100*time.Millisecond)
		check, err := NewHTTPCheck(HTTPCheckConfig{
			CheckName: "url.check",
			URL:       waitURL,
			Client:    client,
			Timeout:   10 * time.Millisecond,
		})
		assert.Nil(t, err)

		details, err := check.Execute(context.Background())
		assert.Error(t, err, "check should fail")
		assert.Contains(t, err.Error(), "Client.Timeout exceeded", "check error message")
		assert.Equal(t, waitURL, details, "check details when fail are the URL")
	}
}
