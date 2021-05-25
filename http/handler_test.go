package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/AppsFlyer/go-sundheit/test/helper"

	"github.com/stretchr/testify/assert"
)

const (
	chkName = "check1"
)

func TestHandleHealthJSON_longFormatNoChecks(t *testing.T) {
	h := gosundheit.New()
	resp := execReq(h, true)
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "status when no checks are registered")
	assert.Equal(t, "{}\n", string(body), "body when no checks are registered")
}

func TestHandleHealthJSON_shortFormatNoChecks(t *testing.T) {
	h := gosundheit.New()
	resp := execReq(h, false)
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "status when no checks are registered")
	assert.Equal(t, "{}\n", string(body), "body when no checks are registered")
}

func TestHandleHealthJSON_longFormatPassingCheck(t *testing.T) {
	checkWaiter := helper.NewCheckWaiter()
	h := gosundheit.New(gosundheit.WithCheckListeners(checkWaiter))

	err := h.RegisterCheck(
		createCheck(chkName, true),
		createCheckOptions(10*time.Millisecond)...,
	)
	if err != nil {
		t.Error("Failed to register check: ", err)
	}
	defer h.DeregisterAll()

	resp := execReq(h, true)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "status before first run")

	var respMsg = unmarshalLongFormat(resp.Body)
	const freshCheckMsg = "didn't run yet"
	expectedResponse := response{
		Check1: checkResult{
			Message: freshCheckMsg,
			Error: Err{
				Message: freshCheckMsg,
			},
			ContiguousFailures: 1,
		},
	}
	assert.Equal(t, &expectedResponse, respMsg, "body when no checks are registered")

	assert.NoError(t, checkWaiter.AwaitChecksCompletion(chkName))

	resp = execReq(h, true)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status before first run")

	respMsg = unmarshalLongFormat(resp.Body)
	expectedResponse = response{
		Check1: checkResult{
			Message:            "pass",
			ContiguousFailures: 0,
		},
	}
	assert.Equal(t, &expectedResponse, respMsg, "body after first run")
}

func TestHandleHealthJSON_shortFormatPassingCheck(t *testing.T) {
	checkWaiter := helper.NewCheckWaiter()
	h := gosundheit.New(gosundheit.WithCheckListeners(checkWaiter))

	err := h.RegisterCheck(
		createCheck(chkName, true),
		createCheckOptions(10*time.Millisecond)...,
	)
	if err != nil {
		t.Error("Failed to register check: ", err)
	}
	defer h.DeregisterAll()

	resp := execReq(h, false)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "status before first run")

	var respMsg = unmarshalShortFormat(resp.Body)
	expectedResponse := map[string]string{"check1": "FAIL"}
	assert.Equal(t, expectedResponse, respMsg, "body when no checks are registered")

	assert.NoError(t, checkWaiter.AwaitChecksCompletion(chkName))
	resp = execReq(h, false)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status before first run")

	respMsg = unmarshalShortFormat(resp.Body)
	expectedResponse = map[string]string{"check1": "PASS"}
	assert.Equal(t, expectedResponse, respMsg, "body after first run")
}

func unmarshalShortFormat(r io.Reader) map[string]string {
	respMsg := make(map[string]string)
	_ = json.NewDecoder(r).Decode(&respMsg)
	return respMsg
}

func unmarshalLongFormat(r io.Reader) *response {
	var respMsg response
	_ = json.NewDecoder(r).Decode(&respMsg)
	return &respMsg
}

func createCheck(name string, passing bool) gosundheit.Check {
	return &checks.CustomCheck{
		CheckName: name,
		CheckFunc: func(ctx context.Context) (details interface{}, err error) {
			if passing {
				return "pass", nil
			}
			return "failing", fmt.Errorf("failing")
		},
	}
}

func createCheckOptions(delay time.Duration) []gosundheit.CheckOption {
	return []gosundheit.CheckOption{
		gosundheit.InitialDelay(delay),
		gosundheit.ExecutionPeriod(delay),
	}
}

func execReq(h gosundheit.Health, longFormat bool) *http.Response {
	var path = "/meh"
	if !longFormat {
		path = fmt.Sprintf("%s?type=%s", path, ReportTypeShort)
	}

	handler := HandleHealthJSON(h)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	return w.Result()
}

type response struct {
	Check1 checkResult `json:"check1"`
}

type checkResult struct {
	Message            string `json:"message"`
	Error              Err    `json:"error"`
	ContiguousFailures int64  `json:"contiguousFailures"`
}

type Err struct {
	Message string `json:"message"`
}
