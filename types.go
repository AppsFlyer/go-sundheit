package gosundheit

import (
	"errors"
	"fmt"
	"time"
)

const (
	maxExpectedChecks = 16
)

var (
	ErrNotRunYet = newMarshalableError(errors.New("didn't run yet"))
)

// Result represents the output of a health check execution.
type Result struct {
	// the details of task Result - may be nil
	Details interface{} `json:"message,omitempty"`
	// the error returned from a failed health check - nil when successful
	Error error `json:"error,omitempty"`
	// the time of the last health check
	Timestamp time.Time `json:"timestamp"`
	// the execution duration of the last check
	Duration time.Duration `json:"duration,omitempty"`
	// the number of failures that occurred in a row
	ContiguousFailures int64 `json:"contiguousFailures"`
	// the time of the initial transitional failure
	TimeOfFirstFailure *time.Time `json:"timeOfFirstFailure"`
}

// IsHealthy returns true iff the check result snapshot was a success
func (r Result) IsHealthy() bool {
	return r.Error == nil
}

func (r Result) String() string {
	return fmt.Sprintf("Result{details: %s, err: %s, time: %s, contiguousFailures: %d, timeOfFirstFailure:%s}",
		r.Details, r.Error, r.Timestamp, r.ContiguousFailures, r.TimeOfFirstFailure)
}

type marshalableError struct {
	Message string `json:"message,omitempty"`
	Cause   error  `json:"cause,omitempty"`
}

func newMarshalableError(err error) error {
	if err == nil {
		return nil
	}

	mr := marshalableError{
		Message: err.Error(),
	}
	cause := errors.Unwrap(err)
	if !errors.Is(cause, err) {
		mr.Cause = newMarshalableError(cause)
	}

	return mr
}

func (e marshalableError) Error() string {
	return e.Message
}
