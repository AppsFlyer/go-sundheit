package opencencus

import (
	"context"
	"log"
	"strconv"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	// ValAllChecks is the value used for the check tags when tagging all tests
	ValAllChecks = "all_checks"
)

var (
	keyCheck, _          = tag.NewKey("check")
	keyCheckPassing, _   = tag.NewKey("check_passing")
	keyClassification, _ = tag.NewKey("classification")

	mCheckStatus   = stats.Int64("health/status", "An health status (0/1 for fail/pass)", "pass/fail")
	mCheckDuration = stats.Float64("health/execute_time", "The time it took to execute a checks in ms", "ms")

	// ViewCheckExecutionTime is the checks execution time aggregation tagged by check name
	ViewCheckExecutionTime = &view.View{
		Measure:     mCheckDuration,
		TagKeys:     []tag.Key{keyCheck, keyClassification},
		Aggregation: view.Distribution(0, 1, 2, 3, 4, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 120, 160, 200, 250, 300, 500),
	}

	// ViewCheckCountByNameAndStatus is the checks execution count aggregation grouped by check name, and check status
	ViewCheckCountByNameAndStatus = &view.View{
		Name:        "health/check_count_by_name_and_status",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck, keyCheckPassing, keyClassification},
		Aggregation: view.Count(),
	}

	// ViewCheckStatusByName is the checks status aggregation tagged by check name
	ViewCheckStatusByName = &view.View{
		Name:        "health/check_status_by_name",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck, keyClassification},
		Aggregation: view.LastValue(),
	}

	// DefaultHealthViews are the default health check views provided by this package.
	DefaultHealthViews = []*view.View{
		ViewCheckCountByNameAndStatus,
		ViewCheckStatusByName,
		ViewCheckExecutionTime,
	}
)

func createMonitoringCtx(classification, checkName string, isPassing bool) (ctx context.Context) {
	tags := []tag.Mutator{
		tag.Insert(keyCheck, checkName),
		tag.Insert(keyCheckPassing, strconv.FormatBool(isPassing)),
	}
	if classification != "" {
		tags = append(tags, tag.Insert(keyClassification, classification))
	}
	ctx, err := tag.New(context.Background(), tags...)
	if err != nil {
		// When this happens it's a programming error caused by the line above
		log.Println("[Error] context creation failed for check ", checkName)
	}

	return
}

type status bool

func (s status) asInt64() int64 {
	if s {
		return 1
	}
	return 0
}
