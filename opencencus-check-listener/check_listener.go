package opencencus

import (
	"time"

	"go.opencensus.io/stats"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

type CheckListener struct {
	classification string
}

func NewCheckListener(opts ...Option) *CheckListener {
	listener := &CheckListener{}

	for _, opt := range append(opts, WithDefaults()) {
		opt(listener)
	}

	return listener
}

func (c *CheckListener) OnCheckRegistered(name string, result gosundheit.Result) {
	c.recordCheck(name, result)
}

func (c *CheckListener) OnCheckStarted(_ string) {
}

func (c *CheckListener) OnCheckCompleted(name string, result gosundheit.Result) {
	c.recordCheck(name, result)
}

func (c *CheckListener) OnResultsUpdated(results map[string]gosundheit.Result) {
	allHealthy := allHealthy(results)
	allChecksCtx := createMonitoringCtx(c.classification, ValAllChecks, allHealthy)
	stats.Record(allChecksCtx, mCheckStatus.M(status(allHealthy).asInt64()))
}

func (c *CheckListener) recordCheck(name string, result gosundheit.Result) {
	thisCheckCtx := createMonitoringCtx(c.classification, name, result.IsHealthy())
	stats.Record(thisCheckCtx, mCheckDuration.M(float64(result.Duration)/float64(time.Millisecond)))
	stats.Record(thisCheckCtx, mCheckStatus.M(status(result.IsHealthy()).asInt64()))
}
