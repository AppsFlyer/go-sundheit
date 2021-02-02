package opencensus

import (
	"time"

	"go.opencensus.io/stats"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

// MetricsListener reports metrics on each check registration, start and completion event (as gosundheit.CheckListener)
// This listener all reports metrics for the entire service health (as gosundheit.HealthListener)
type MetricsListener struct {
	classification string
}

func NewMetricsListener(opts ...Option) *MetricsListener {
	listener := &MetricsListener{}

	for _, opt := range append(opts, WithDefaults()) {
		opt(listener)
	}

	return listener
}

func (c *MetricsListener) OnCheckRegistered(name string, result gosundheit.Result) {
	c.recordCheck(name, result)
}

func (c *MetricsListener) OnCheckStarted(_ string) {
}

func (c *MetricsListener) OnCheckCompleted(name string, result gosundheit.Result) {
	c.recordCheck(name, result)
}

func (c *MetricsListener) OnResultsUpdated(results map[string]gosundheit.Result) {
	allHealthy := allHealthy(results)
	allChecksCtx := createMonitoringCtx(c.classification, ValAllChecks, allHealthy)
	stats.Record(allChecksCtx, mCheckStatus.M(status(allHealthy).asInt64()))
}

func (c *MetricsListener) recordCheck(name string, result gosundheit.Result) {
	thisCheckCtx := createMonitoringCtx(c.classification, name, result.IsHealthy())
	stats.Record(thisCheckCtx, mCheckDuration.M(float64(result.Duration)/float64(time.Millisecond)))
	stats.Record(thisCheckCtx, mCheckStatus.M(status(result.IsHealthy()).asInt64()))
}
