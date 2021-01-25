package gosundheit

import (
	"context"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"log"
	"strconv"
)

var (
	keyCheck, _        = tag.NewKey("check")
	keyCheckPassing, _ = tag.NewKey("check_passing")

	mCheckStatus   = stats.Int64("health/status", "An health status (0/1 for fail/pass)", "pass/fail")
	mCheckDuration = stats.Float64("health/execute_time", "The time it took to execute a checks in ms", "ms")

	// ViewCheckExecutionTime is the checks execution time aggregation tagged by check name
	ViewCheckExecutionTime = &view.View{
		Measure:     mCheckDuration,
		TagKeys:     []tag.Key{keyCheck},
		Aggregation: view.Distribution(0, 1, 2, 3, 4, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 120, 160, 200, 250, 300, 500),
	}

	// ViewCheckCountByNameAndStatus is the checks execution count aggregation grouped by check name, and check status
	ViewCheckCountByNameAndStatus = &view.View{
		Name:        "health/check_count_by_name_and_status",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck, keyCheckPassing},
		Aggregation: view.Count(),
	}

	// ViewCheckStatusByName is the checks status aggregation tagged by check name
	ViewCheckStatusByName = &view.View{
		Name:        "health/check_status_by_name",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck},
		Aggregation: view.LastValue(),
	}

	// DefaultHealthViews are the default health check views provided by this package.
	DefaultHealthViews = []*view.View{
		ViewCheckCountByNameAndStatus,
		ViewCheckStatusByName,
		ViewCheckExecutionTime,
	}
)

func createMonitoringCtx(checkName string, isPassing bool) (ctx context.Context) {
	ctx, err := tag.New(context.Background(), tag.Insert(keyCheck, checkName), tag.Insert(keyCheckPassing, strconv.FormatBool(isPassing)))
	if err != nil {
		// When this happens it's a programming error caused by the line above
		log.Println("[Error] context creation failed for check ", checkName)
	}

	return
}