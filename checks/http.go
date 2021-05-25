package checks

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/pkg/errors"
)

// HTTPCheckConfig configures a check for the response from a given URL.
// The only required field is `URL`, which must be a valid URL.
type HTTPCheckConfig struct {
	// CheckName is the health check name - must be a valid metric name.
	// CheckName is required
	CheckName string
	// URL is required valid URL, to be called by the check
	URL string
	// Method is the HTTP method to use for this check.
	// Method is optional and defaults to `GET` if undefined.
	Method string
	// Body is an optional request body to be posted to the target URL.
	Body BodyProvider
	// ExpectedStatus is the expected response status code, defaults to `200`.
	ExpectedStatus int
	// ExpectedBody is optional; if defined, operates as a basic "body should contain <string>".
	ExpectedBody string
	// Client is optional; if undefined, a new client will be created using "Timeout".
	Client *http.Client
	// Timeout is the timeout used for the HTTP request, defaults to "1s".
	Timeout time.Duration
	// Options allow you to configure the HTTP request with arbitrary settings, e.g. add request headers, etc.
	Options []RequestOption
}

// RequestOption configures the request with arbitrary settings, e.g. add request headers, etc.
type RequestOption func(r *http.Request)

type httpCheck struct {
	config         *HTTPCheckConfig
	successDetails string
}

// BodyProvider allows the users to provide a body to the HTTP checks. For example for posting a payload as a check.
type BodyProvider func() io.Reader

// NewHTTPCheck creates a new http check defined by the given config
func NewHTTPCheck(config HTTPCheckConfig) (check gosundheit.Check, err error) {
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

	if config.ExpectedStatus == 0 {
		config.ExpectedStatus = http.StatusOK
	}
	if config.Method == "" {
		config.Method = http.MethodGet
	}
	if config.Body == nil {
		config.Body = func() io.Reader { return http.NoBody }
	}
	if config.Timeout == 0 {
		config.Timeout = time.Second
	}
	if config.Client == nil {
		config.Client = &http.Client{}
	}
	config.Client.Timeout = config.Timeout

	check = &httpCheck{
		config:         &config,
		successDetails: fmt.Sprintf("URL [%s] is accessible", config.URL),
	}
	return check, nil
}

func (check *httpCheck) Name() string {
	return check.config.CheckName
}

func (check *httpCheck) Execute(ctx context.Context) (details interface{}, err error) {
	details = check.config.URL
	resp, err := check.fetchURL(ctx)
	if err != nil {
		return details, err
	}
	defer func() { _ = resp.Body.Close() }()

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

// fetchURL executes the HTTP request to the target URL, and returns a `http.Response`, error.
// It is the callers responsibility to close the response body
func (check *httpCheck) fetchURL(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, check.config.Method, check.config.URL, check.config.Body())
	if err != nil {
		return nil, errors.Errorf("unable to create check HTTP request: %v", err)
	}

	configureHTTPOptions(req, check.config.Options)

	resp, err := check.config.Client.Do(req)
	if err != nil {
		return nil, errors.Errorf("fail to execute '%v' request: %v", check.config.Method, err)
	}

	return resp, nil
}

func configureHTTPOptions(req *http.Request, options []RequestOption) {
	for _, opt := range options {
		opt(req)
	}
}
