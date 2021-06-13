package checks

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHostResolveCheck(t *testing.T) {
	check := NewHostResolveCheck("127.0.0.1", 1)

	assert.Equal(t, "resolve.127.0.0.1", check.Name(), "check name")

	details, err := check.Execute(context.Background())
	assert.NoError(t, err, "check execution should succeed")
	assert.Equal(t, "[1] results were resolved", details)
}

func TestNewHostResolveCheck_noSuchHost(t *testing.T) {
	check := NewHostResolveCheck("I-hope-there-is.no.such.host.com", 1)

	assert.Equal(t, "resolve.I-hope-there-is.no.such.host.com", check.Name(), "check name")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	details, err := check.Execute(ctx)

	assert.Error(t, err, "check execution should fail")
	assert.Contains(t, err.Error(), "no such host")
	assert.Equal(t, "[0] results were resolved", details)
}

func TestNewHostResolveCheck_timeout(t *testing.T) {
	check := NewHostResolveCheck("I-hope-there-is.no.such.host.com", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()
	details, err := check.Execute(ctx)

	assert.Error(t, err, "check execution should fail")
	assert.Contains(t, err.Error(), "i/o timeout")
	assert.Equal(t, "[0] results were resolved", details)
}

const (
	ExpectedError = "fail-fail-fail"
	ExpectedCount = 666
)

func TestNewResolveCheck_lookupError(t *testing.T) {
	check := NewResolveCheck(creteMockLookupFunc(ExpectedCount, errors.New(ExpectedError)), "whatever", 1)

	assert.Equal(t, "resolve.whatever", check.Name(), "check name")
	details, err := check.Execute(context.Background())
	assert.EqualErrorf(t, err, ExpectedError, "error message")
	assert.Equal(t, fmt.Sprintf("[%d] results were resolved", ExpectedCount), details)
}

func TestNewResolveCheck_expectedCount(t *testing.T) {
	check := NewResolveCheck(creteMockLookupFunc(0, nil), "whatever", ExpectedCount)

	details, err := check.Execute(context.Background())
	assert.EqualErrorf(t, err, fmt.Sprintf("[whatever] lookup returned 0 results, but requires at least %d", ExpectedCount), "error message")
	assert.Equal(t, "[0] results were resolved", details)
}

func creteMockLookupFunc(resultCount int, err error) LookupFunc {
	return func(ctx context.Context, host string) (int, error) {
		return resultCount, err
	}
}
