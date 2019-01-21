package checks

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"net/http"
	"strings"
	"fmt"
	"time"
)

const (
	longRequest     = "LongRequest"
	expectedContent = "I'm healthy"
)

func TestNewHttpCheckRequiredFields(t *testing.T) {
	check, err := NewHttpCheck(nil)
	assert.Nil(t, check, "nil config should yield nil check")
	assert.Error(t, err, "nil config should yield error")

	check, err = NewHttpCheck(&HttpCheckConfig{
		CheckName: "meh",
	})
	assert.Nil(t, check, "nil URL should yield nil check")
	assert.Error(t, err, "nil URL should yield error")

	check, err = NewHttpCheck(&HttpCheckConfig{
		URL: "http://example.org",
	})
	assert.Nil(t, check, "nil CheckName should yield nil check")
	assert.Error(t, err, "nil CheckName should yield error")

	check, err = NewHttpCheck(&HttpCheckConfig{
		URL:       ":/invalid.url",
		CheckName: "invalid.url",
	})
	assert.Nil(t, check, "invalid url should yield nil check")
	assert.Error(t, err, "invalid url should yield error")
}

func TestNewHttpCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.Contains(req.URL.String(), longRequest) {
			waitDuration, err := time.ParseDuration(req.URL.Query().Get("wait"))
			if err != nil {
				fmt.Println("Failed to parse sleep duration: ", err)
				waitDuration = 10 * time.Second
			}

			time.Sleep(waitDuration)
		}

		rw.WriteHeader(200)
		rw.Write([]byte(expectedContent))
	}))

	defer server.Close()

	t.Run("HttpCheck success call", testHttpCheckSuccess(server.URL, server.Client()))
	t.Run("HttpCheck success call with body check", testHttpCheckSuccessWithExpectedBody(server.URL, server.Client()))
	t.Run("HttpCheck success call with failing body check", testHttpCheckFailWithUnexpectedBody(server.URL, server.Client()))
	t.Run("HttpCheck fail on status code", testHttpCheckFailStatusCode(server.URL, server.Client()))
	t.Run("HttpCheck fail on URL", testHttpCheckFailURL(server.URL, server.Client()))
	t.Run("HttpCheck fail on timeout", testHttpCheckFailTimeout(server.URL, server.Client()))
}

func testHttpCheckSuccess(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName: "url.check",
			URL:       url,
			Client:    client,
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Nil(t, err, "check should pass")
		assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
	}
}

func testHttpCheckSuccessWithExpectedBody(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName:    "url.check",
			URL:          url,
			Client:       client,
			ExpectedBody: expectedContent,
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Nil(t, err, "check should pass")
		assert.Equal(t, fmt.Sprintf("URL [%s] is accessible", url), details, "check should pass")
	}
}

func testHttpCheckFailWithUnexpectedBody(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName:    "url.check",
			URL:          url,
			Client:       client,
			ExpectedBody: "my body is a temple",
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Error(t, err, "check should fail")
		assert.Equal(t, "body does not contain expected content 'my body is a temple'", err.Error(), "check error message")
		assert.Equal(t, url, details, "check details when fail are the URL")
	}
}

func testHttpCheckFailStatusCode(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName:      "url.check",
			URL:            url,
			Client:         client,
			ExpectedStatus: 300,
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Error(t, err, "check should fail")
		assert.Equal(t, "unexpected status code: '200' expected: '300'", err.Error(), "check error message")
		assert.Equal(t, url, details, "check details when fail are the URL")
	}
}

func testHttpCheckFailURL(_ string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		bogusUrl := "http://devil.dot.com:666"
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName: "url.check",
			URL:       bogusUrl,
			Client:    client,
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Error(t, err, "check should fail")
		assert.Contains(t, err.Error(), "no such host", "check error message")
		assert.Equal(t, bogusUrl, details, "check details when fail are the URL")
	}
}

func testHttpCheckFailTimeout(url string, client *http.Client) func(t *testing.T) {
	return func(t *testing.T) {
		waitUrl := fmt.Sprintf("%s/%s?wait=%s", url, longRequest, 100*time.Millisecond)
		check, err := NewHttpCheck(&HttpCheckConfig{
			CheckName: "url.check",
			URL:       waitUrl,
			Client:    client,
			Timeout:   10 * time.Millisecond,
		})
		assert.Nil(t, err)

		details, err := check.Execute()
		assert.Error(t, err, "check should fail")
		assert.Contains(t, err.Error(), "Client.Timeout exceeded", "check error message")
		assert.Equal(t, waitUrl, details, "check details when fail are the URL")
	}
}
