package opencencus

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	gosundheit "github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
)

const (
	successMsg = "success"
	failedMsg  = "failed"

	failingCheckName = "failing.check"
	passingCheckName = "passing.check"
)

func TestHealthMetrics(t *testing.T) {
	_ = view.Register(DefaultHealthViews...)

	listener := NewCheckListener()
	h := gosundheit.New(gosundheit.WithCheckListeners(listener), gosundheit.WithHealthListeners(listener))
	registerCheck(h, failingCheckName, false, false)
	registerCheck(h, passingCheckName, true, false)
	defer h.DeregisterAll()

	// await first execution
	time.Sleep(25 * time.Millisecond)

	checksStatusData := simplifyRows(ViewCheckStatusByName.Name)
	assert.Equal(t, 3, len(checksStatusData), "num status rows")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[ValAllChecks], "all check status")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[failingCheckName], "failing check status")
	assert.Equal(t, &view.LastValueData{Value: 1}, checksStatusData[passingCheckName], "passing check status")

	checksCountData := simplifyRows(ViewCheckCountByNameAndStatus.Name)
	assert.Equal(t, 4, len(checksCountData), "num count rows")
	// at this stage there should have been 2 "executions" of each check, the initial state is always failing
	assert.Equal(t, &view.CountData{Value: 2}, checksCountData[ValAllChecks+".false"], "all checks fail count")
	assert.Equal(t, &view.CountData{Value: 2}, checksCountData[failingCheckName+".false"], "failing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".false"], "passing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".true"], "passing check pass count")

	checksTimeData := simplifyRows(ViewCheckExecutionTime.Name)
	assert.Equal(t, 2, len(checksTimeData), "num timing rows")
	assert.Equal(t, int64(2), checksTimeData[passingCheckName].(*view.DistributionData).Count, "passing check timing measurement count")
	assert.Equal(t, int64(2), checksTimeData[failingCheckName].(*view.DistributionData).Count, "failing check timing measurement count")

	view.Unregister(DefaultHealthViews...)
}

func runTestHealthMetricsWithClassification(t *testing.T, option Option, classification string) {
	_ = view.Register(DefaultHealthViews...)

	listener := NewCheckListener(option)
	h := gosundheit.New(gosundheit.WithCheckListeners(listener), gosundheit.WithHealthListeners(listener))
	registerCheck(h, failingCheckName, false, false)
	registerCheck(h, passingCheckName, true, false)
	defer h.DeregisterAll()

	// await first execution
	time.Sleep(25 * time.Millisecond)

	checksStatusData := simplifyRows(ViewCheckStatusByName.Name)
	assert.Equal(t, 3, len(checksStatusData), "num status rows")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[ValAllChecks+"."+classification], "all check status")
	assert.Equal(t, &view.LastValueData{Value: 0}, checksStatusData[failingCheckName+"."+classification], "failing check status")
	assert.Equal(t, &view.LastValueData{Value: 1}, checksStatusData[passingCheckName+"."+classification], "passing check status")

	checksCountData := simplifyRows(ViewCheckCountByNameAndStatus.Name)
	assert.Equal(t, 4, len(checksCountData), "num count rows")
	// at this stage there should have been 2 "executions" of each check, the initial state is always failing
	assert.Equal(t, &view.CountData{Value: 2}, checksCountData[ValAllChecks+".false"+"."+classification], "all checks fail count")
	assert.Equal(t, &view.CountData{Value: 2}, checksCountData[failingCheckName+".false"+"."+classification], "failing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".false"+"."+classification], "passing check fail count")
	assert.Equal(t, &view.CountData{Value: 1}, checksCountData[passingCheckName+".true"+"."+classification], "passing check pass count")

	checksTimeData := simplifyRows(ViewCheckExecutionTime.Name)
	assert.Equal(t, 2, len(checksTimeData), "num timing rows")
	assert.Equal(t, int64(2), checksTimeData[passingCheckName+"."+classification].(*view.DistributionData).Count, "passing check timing measurement count")
	assert.Equal(t, int64(2), checksTimeData[failingCheckName+"."+classification].(*view.DistributionData).Count, "failing check timing measurement count")

	view.Unregister(DefaultHealthViews...)
}

func TestHealthMetricsWithLivenessClassification(t *testing.T) {
	runTestHealthMetricsWithClassification(t, WithLivenessClassification(), "liveness")
}

func TestHealthMetricsWithStartupClassification(t *testing.T) {
	runTestHealthMetricsWithClassification(t, WithStartupClassification(), "startup")
}

func TestHealthMetricsWithReadinessClassification(t *testing.T) {
	runTestHealthMetricsWithClassification(t, WithReadinessClassification(), "readiness")
}

func TestHealthMetricsWithCustomClassification(t *testing.T) {
	runTestHealthMetricsWithClassification(t, WithClassification("demo"), "demo")
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

func registerCheck(h gosundheit.Health, name string, passing bool, initiallyPassing bool) {
	stub := checkStub{
		counts:  0,
		passing: passing,
	}

	_ = h.RegisterCheck(&gosundheit.Config{
		Check: &checks.CustomCheck{
			CheckName: name,
			CheckFunc: stub.run,
		},
		InitialDelay:     20 * time.Millisecond,
		ExecutionPeriod:  120 * time.Millisecond,
		InitiallyPassing: initiallyPassing,
	})
}

type checkStub struct {
	counts  int64
	passing bool
}

func (c *checkStub) run() (details interface{}, err error) {
	c.counts++
	if c.passing {
		return fmt.Sprintf("%s; i=%d", successMsg, c.counts), nil
	}

	return fmt.Sprintf("%s; i=%d", failedMsg, c.counts), errors.New(failedMsg)
}
