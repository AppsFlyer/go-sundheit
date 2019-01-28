package health

import (
	"testing"
	"time"
	"errors"
	"fmt"

	"github.com/stretchr/testify/assert"

	"gitlab.appsflyer.com/Architecture/af-go-health/checks"
	"github.com/fortytw2/leaktest"
)

const (
	successMsg = "success"
	failedMsg  = "failed"
)

func TestHealthWithEmptySetup(t *testing.T) {
	h := New()

	assert.True(t, h.IsHealthy(), "empty health")

	results, healthy := h.Results()
	assert.True(t, healthy, "results of empty setup")
	assert.Empty(t, results, "results of empty setup")

	h.DeregisterAll()
}

func TestHealthWithBogusCheck(t *testing.T) {
	h := New()

	err := h.RegisterCheck(&Config{
		ExecutionPeriod: 1,
		InitialDelay:    1,
	})

	assert.Error(t, err, "register bogus check should fail")
	assert.Contains(t, err.Error(), "misconfigured check", "fail message")
	assert.True(t, h.IsHealthy(), "health after bogus register")

	results, healthy := h.Results()
	assert.True(t, healthy, "results after bogus register")
	assert.Empty(t, results, "results after bogus register")

	h.DeregisterAll()
}

func TestRegisterDeregister(t *testing.T) {
	leaktest.Check(t)

	h := New()

	const (
		failingCheckName = "failing.check"
		passingCheckName = "passing.check"
	)
	registerCheck(h, failingCheckName, false)
	registerCheck(h, passingCheckName, true)

	assert.False(t, h.IsHealthy(), "health after registration before first run")
	results, healthy := h.Results()
	assert.False(t, healthy, "health results after registration before first run")
	assert.NotEmpty(t, results, "health after registration before first run")

	passingCheck, ok1 := results[passingCheckName]
	failingCheck, ok2 := results[failingCheckName]
	assert.True(t, ok1, "check exists")
	assert.True(t, ok2, "check exists")
	assert.False(t, passingCheck.IsHealthy(), "check initially fails until first execution")
	assert.False(t, failingCheck.IsHealthy(), "check initially fails until first execution")
	assert.Contains(t, passingCheck.String(), "didn't run yet", "initial details")
	assert.Contains(t, failingCheck.String(), "didn't run yet", "initial details")

	// await first execution
	time.Sleep(50 * time.Millisecond)

	assert.False(t, h.IsHealthy(), "health after registration before first run with one failing check")
	results, healthy = h.Results()
	assert.False(t, healthy, "health results after registration before first run with one failing check")

	passingCheck, ok1 = results[passingCheckName]
	failingCheck, ok2 = results[failingCheckName]
	passingCheck.IsHealthy()

	assert.True(t, ok1, "check exists")
	assert.True(t, ok2, "check exists")
	assert.True(t, passingCheck.IsHealthy(), "succeeding check should pass")
	assert.False(t, failingCheck.IsHealthy(), "failing check check should fail")
	assert.NotContains(t, passingCheck.String(), "didn't run yet", "details after execution")
	assert.NotContains(t, failingCheck.String(), "didn't run yet", "details after execution")
	assert.Contains(t, passingCheck.String(), "success", "details after execution")
	assert.Contains(t, failingCheck.String(), "fail", "details after execution")

	h.Deregister(failingCheckName)
	// await check cleanup
	time.Sleep(50 * time.Millisecond)

	assert.True(t, h.IsHealthy(), "health after failing test deregistration")

	results, healthy = h.Results()
	assert.True(t, healthy, "results of empty setup")
	assert.Equal(t, 1, len(results), "num results after deregistration")
	_, ok1 = results[passingCheckName]
	_, ok2 = results[failingCheckName]
	assert.True(t, ok1, "check exists")
	assert.False(t, ok2, "check should have been removed")

	h.DeregisterAll()

	// await stop
	time.Sleep(50 * time.Millisecond)
	results, _ = h.Results()
	assert.Empty(t, results, "results after stop")
}

func registerCheck(h Health, name string, passing bool) {
	i := 0
	checkFunc := func() (details interface{}, err error) {
		i++

		if passing {
			return fmt.Sprintf("%s; i=%d", successMsg, i), nil
		}

		return fmt.Sprintf("%s; i=%d", failedMsg, i), errors.New(failedMsg)
	}

	h.RegisterCheck(&Config{
		Check: &checks.CustomCheck{
			CheckName: name,
			CheckFunc: checkFunc,
		},
		InitialDelay:    20 * time.Millisecond,
		ExecutionPeriod: 20 * time.Millisecond,
	})
}
