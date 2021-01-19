package gosundheit

import (
	"github.com/pkg/errors"
)

const (
	maxExpectedChecks = 16
	initialResultMsg  = "didn't run yet"
)

type marshalableError struct {
	Message string `json:"message,omitempty"`
	Cause   error  `json:"cause,omitempty"`
}

func newMarshalableError(err error) error {
	if err == nil {
		return nil
	}

	mr := &marshalableError{
		Message: err.Error(),
	}

	cause := errors.Cause(err)
	if cause != err {
		mr.Cause = newMarshalableError(cause)
	}

	return mr
}

func (e *marshalableError) Error() string {
	return e.Message
}
