package gosundheit

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/AppsFlyer/go-sundheit/checks"
	"github.com/AppsFlyer/go-sundheit/internal/metrics"
	"github.com/AppsFlyer/go-sundheit/internal/types"
)

const (
	successMsg = "success"
	failedMsg  = "failed"

	failingCheckName          = "failing.check"
	passingCheckName          = "passing.check"
	initiallyPassingCheckName = "initially.passing.check"
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
	defer h.DeregisterAll()

	assert.Error(t, err, "register bogus check should fail")
	assert.Contains(t, err.Error(), "misconfigured check", "fail message")
	assert.True(t, h.IsHealthy(), "health after bogus register")

	results, healthy := h.Results()
	assert.True(t, healthy, "results after bogus register")
	assert.Empty(t, results, "results after bogus register")
}

func TestRegisterDeregister(t *testing.T) {
	leaktest.Check(t)

	h := New()

	registerCheck(h, failingCheckName, false, false)
	registerCheck(h, passingCheckName, true, false)
	registerCheck(h, initiallyPassingCheckName, true, true)

	assert.False(t, h.IsHealthy(), "health after registration before first run")
	results, healthy := h.Results()
	assert.False(t, healthy, "health results after registration before first run")
	assert.NotEmpty(t, results, "health after registration before first run")

	passingCheck, ok1 := results[passingCheckName]
	failingCheck, ok2 := results[failingCheckName]
	initiallyPassingCheck, ok3 := results[initiallyPassingCheckName]
	assert.True(t, ok1, "check exists")
	assert.True(t, ok2, "check exists")
	assert.True(t, ok3, "check exists")
	assert.False(t, passingCheck.IsHealthy(), "check initially fails until first execution by default")
	assert.False(t, failingCheck.IsHealthy(), "check initially fails until first execution by default")
	assert.True(t, initiallyPassingCheck.IsHealthy(), "check should initially pass")
	assert.Contains(t, passingCheck.String(), "didn't run yet", "initial details")
	assert.Contains(t, failingCheck.String(), "didn't run yet", "initial details")
	assert.Contains(t, initiallyPassingCheck.String(), "didn't run yet", "initial details")

	// await first execution
	time.Sleep(50 * time.Millisecond)

	assert.False(t, h.IsHealthy(), "health after registration before first run with one failing check")
	results, healthy = h.Results()
	assert.False(t, healthy, "health results after registration before first run with one failing check")

	passingCheck, ok1 = results[passingCheckName]
	failingCheck, ok2 = results[failingCheckName]
	initiallyPassingCheck, ok3 = results[initiallyPassingCheckName]

	assert.True(t, ok1, "check exists")
	assert.True(t, ok2, "check exists")
	assert.True(t, ok3, "check exists")
	assert.True(t, passingCheck.IsHealthy(), "succeeding check should pass")
	assert.False(t, failingCheck.IsHealthy(), "failing check check should fail")
	assert.True(t, initiallyPassingCheck.IsHealthy(), "passing check check should pass")
	assert.NotContains(t, passingCheck.String(), "didn't run yet", "details after execution")
	assert.NotContains(t, failingCheck.String(), "didn't run yet", "details after execution")
	assert.NotContains(t, initiallyPassingCheck.String(), "didn't run yet", "details after execution")
	assert.Contains(t, passingCheck.String(), "success", "details after execution")
	assert.Contains(t, failingCheck.String(), "fail", "details after execution")
	assert.Contains(t, initiallyPassingCheck.String(), "success", "details after execution")

	h.Deregister(failingCheckName)
	// await check cleanup
	time.Sleep(50 * time.Millisecond)

	assert.True(t, h.IsHealthy(), "health after failing checks deregistration")

	results, healthy = h.Results()
	assert.True(t, healthy, "results of only passing checks should be healthy")
	assert.Equal(t, 2, len(results), "num results after deregistration")
	_, ok1 = results[passingCheckName]
	_, ok2 = results[failingCheckName]
	_, ok3 = results[initiallyPassingCheckName]
	assert.True(t, ok1, "check exists")
	assert.False(t, ok2, "check should have been removed")
	assert.True(t, ok3, "check exists")

	h.DeregisterAll()

	// await stop
	time.Sleep(50 * time.Millisecond)
	results, _ = h.Results()
	assert.Empty(t, results, "results after stop")
}

func registerCheck(h Health, name string, passing bool, initiallyPassing bool) {
	i := 0
	checkFunc := func() (details interface{}, err error) {
		i++

		if passing {
			return fmt.Sprintf("%s; i=%d", successMsg, i), nil
		}

		return fmt.Sprintf("%s; i=%d", failedMsg, i), errors.New(failedMsg)
	}

	_ = h.RegisterCheck(&Config{
		Check: &checks.CustomCheck{
			CheckName: name,
			CheckFunc: checkFunc,
		},
		InitialDelay:     20 * time.Millisecond,
		ExecutionPeriod:  20 * time.Millisecond,
		InitiallyPassing: initiallyPassing,
	})
}

func TestHealthMetrics(t *testing.T) {

	h := New()
	views := h.Views()
	_ = view.Register(views.DefaultViews...)
	registerCheck(h, failingCheckName, false, false)
	registerCheck(h, passingCheckName, true, false)
	defer h.DeregisterAll()

	// await first execution
	time.Sleep(21 * time.Millisecond)

	checksStatusData := simplifyRows(views.ViewCheckStatusByName.Name)
	assert.Equal(t, 3, len(checksStatusData), "num status rows")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[metrics.ValAllChecks], "all check status")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[failingCheckName], "failing check status")
	assert.Equal(t, &view.LastValueData{Value: 1}, checksStatusData[passingCheckName], "passing check status")

	checksCountData := simplifyRows(views.ViewCheckCountByNameAndStatus.Name)
	assert.Equal(t, 4, len(checksCountData), "num count rows")
	// at this stage there should have been 2 "executions" of each check, the initial state is always failing
	assert.Equal(t, &view.CountData{Value: 4}, checksCountData[metrics.ValAllChecks+".false"], "all checks fail count")
	assert.Equal(t, &view.CountData{Value: 2}, checksCountData[failingCheckName+".false"], "failing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".false"], "passing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".true"], "passing check pass count")

	checksTimeData := simplifyRows(views.ViewCheckExecutionTime.Name)
	assert.Equal(t, 2, len(checksTimeData), "num timing rows")
	assert.Equal(t, int64(2), checksTimeData[passingCheckName].(*view.DistributionData).Count, "passing check timing measurement count")
	assert.Equal(t, int64(2), checksTimeData[failingCheckName].(*view.DistributionData).Count, "failing check timing measurement count")
}

func TestCheckListener(t *testing.T) {

	listenerMock := &checkListenerMock{}
	listenerMock.On("OnCheckStarted", failingCheckName).Return()
	listenerMock.On("OnCheckStarted", passingCheckName).Return()
	listenerMock.On("OnCheckCompleted", failingCheckName, mock.AnythingOfType("Result")).Return()
	listenerMock.On("OnCheckCompleted", passingCheckName, mock.AnythingOfType("Result")).Return()
	h := New(WithCheckListener(listenerMock))

	registerCheck(h, failingCheckName, false, false)
	registerCheck(h, passingCheckName, true, false)
	defer h.DeregisterAll()

	// await first execution
	time.Sleep(30 * time.Millisecond)

	listenerMock.AssertExpectations(t)

	completedChecks := listenerMock.getCompletedChecks()
	assert.Equal(t, 2, len(completedChecks), "num completed checks")

	for _, c := range completedChecks {
		if c.name == failingCheckName {
			assert.False(t, c.res.IsHealthy())
			assert.Error(t, c.res.Error)
			assert.Equal(t, "failed; i=1", c.res.Details)
		} else {
			assert.True(t, c.res.IsHealthy())
			assert.NoError(t, c.res.Error)
			assert.Equal(t, "success; i=1", c.res.Details)
		}
	}
}

func (l *checkListenerMock) getCompletedChecks() []completedCheck {
	l.lock.RLock()
	defer l.lock.RUnlock()

	return l.completed
}

type checkListenerMock struct {
	mock.Mock
	completed []completedCheck
	lock      sync.RWMutex
}

type completedCheck struct {
	name string
	res  types.Result
}

func (l *checkListenerMock) OnCheckStarted(name string) {
	l.Called(name)
}

func (l *checkListenerMock) OnCheckCompleted(name string, res types.Result) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.Called(name, res)
	l.completed = append(l.completed, completedCheck{name, res})
}

func simplifyRows(viewName string) (check2data map[string]view.AggregationData) {
	rows, err := view.RetrieveData(viewName)
	if err != nil {
		return nil
	}

	check2data = make(map[string]view.AggregationData)
	for _, r := range rows {
		check2data[tagsString(r.Tags)] = r.Data
	}

	return check2data
}

func tagsString(tags []tag.Tag) string {
	var values []string
	for _, t := range tags {
		values = append(values, t.Value)
	}

	return strings.Join(values, ".")
}
