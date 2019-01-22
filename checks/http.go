package checks

import (
	"net/url"
	"net/http"
	"time"
	"io"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

// HttpCheckConfig configures a check for the response from a given URL.
// The only required field is `URL`, which must be a valid URL.
type HttpCheckConfig struct {
	// CheckName is the health check name - must be a valid metric name.
	// CheckName is required
	CheckName string
	// URL is required valid URL, to be called by the check
	URL string
	// Method is the HTTP method to use for this check.
	// Method is optional and defaults to `GET` if undefined.
	Method string
	// Body is an optional request body to be posted to the target URL.
	Body io.Reader
	// ExpectedStatus is the expected response status code, defaults to `200`.
	ExpectedStatus int
	// ExpectedBody is optional; if defined, operates as a basic "body should contain <string>".
	ExpectedBody string
	// Client is optional; if undefined, a new client will be created using "Timeout".
	Client *http.Client
	// Timeout is the timeout used for the HTTP request, defaults to "1s".
	Timeout time.Duration
}

type httpCheck struct {
	config         *HttpCheckConfig
	successDetails string
}

// NewHttpCheck creates a new http check defined by the given config
func NewHttpCheck(config *HttpCheckConfig) (check Check, err error) {
	if config == nil {
		return nil, errors.Errorf("config must not be nil")
	}
	if config.URL == "" {
		return nil, errors.Errorf("URL must not be empty")
	}
	_, err = url.Parse(config.URL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if config.CheckName == "" {
		return nil, errors.Errorf("CheckName must not be empty")
	}

	fullConfig := *config
	if fullConfig.ExpectedStatus == 0 {
		fullConfig.ExpectedStatus = http.StatusOK
	}
	if fullConfig.Method == "" {
		fullConfig.Method = http.MethodGet
	}
	if fullConfig.Timeout == 0 {
		fullConfig.Timeout = time.Second
	}
	if fullConfig.Client == nil {
		fullConfig.Client = &http.Client{}
	}
	fullConfig.Client.Timeout = fullConfig.Timeout

	check = &httpCheck{
		config:         &fullConfig,
		successDetails: fmt.Sprintf("URL [%s] is accessible", config.URL),
	}
	return check, nil
}

func (check *httpCheck) Name() string {
	return check.config.CheckName
}

func (check *httpCheck) Execute() (details interface{}, err error) {
	details = check.config.URL
	resp, err := check.fetchUrl()
	if err != nil {
		return details, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != check.config.ExpectedStatus {
		return details, errors.Errorf("unexpected status code: '%v' expected: '%v'",
			resp.StatusCode, check.config.ExpectedStatus)
	}

	if check.config.ExpectedBody != "" {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return details, errors.Errorf("failed to read response body: %v", err)
		}

		if !strings.Contains(string(body), check.config.ExpectedBody) {
			return details, errors.Errorf("body does not contain expected content '%v'", check.config.ExpectedBody)
		}
	}

	return check.successDetails, nil

}

// fetchUrl executes the HTTP request to the target URL, and returns a `http.Response`, error.
// It is the callers responsibility to close the response body
func (check *httpCheck) fetchUrl() (*http.Response, error) {
	req, err := http.NewRequest(check.config.Method, check.config.URL, check.config.Body)
	if err != nil {
		return nil, errors.Errorf("unable to create check HTTP request: %v", err)
	}

	resp, err := check.config.Client.Do(req)
	if err != nil {
		return nil, errors.Errorf("fail to execute '%v' request: %v", check.config.Method, err)
	}

	return resp, nil
}
