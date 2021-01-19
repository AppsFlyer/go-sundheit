package metrics

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/AppsFlyer/go-sundheit/internal/types"
)

// Stats records measurements for status and duration per check
type Stats struct {
	prefix        string
	checkDuration *stats.Float64Measure
	checkStatus   *stats.Int64Measure
}

func NewStats(prefix string) *Stats {
	trimmed := strings.TrimSpace(prefix)
	if len(trimmed) == 0 {
		trimmed = "health"
	}
	return &Stats{
		prefix: trimmed,
		checkStatus: stats.Int64(
			fmt.Sprintf("%s/status", trimmed),
			"An health status (0/1 for fail/pass)",
			"pass/fail"),
		checkDuration: stats.Float64(
			fmt.Sprintf("%s/execute_time", trimmed),
			"The time it took to execute a checks in ms",
			"ms"),
	}
}

// RecordStats record the duration and status of specific check and also boolean to record status of all results
func (s *Stats) RecordStats(checkName string, result types.Result, allResults bool) {
	thisCheckCtx := s.createMonitoringCtx(checkName, result.IsHealthy())
	stats.Record(thisCheckCtx, s.checkDuration.M(float64(result.Duration)/float64(time.Millisecond)))
	stats.Record(thisCheckCtx, s.checkStatus.M(status(result.IsHealthy()).asInt64()))

	allChecksCtx := s.createMonitoringCtx(ValAllChecks, allResults)
	stats.Record(allChecksCtx, s.checkStatus.M(status(allResults).asInt64()))
}

func (s *Stats) createMonitoringCtx(checkName string, isPassing bool) (ctx context.Context) {
	ctx, err := tag.New(context.Background(), tag.Insert(keyCheck, checkName), tag.Insert(keyCheckPassing, strconv.FormatBool(isPassing)))
	if err != nil {
		// When this happens it's a programming error caused by the line above
		log.Println("[Error] context creation failed for check ", checkName)
	}

	return
}
