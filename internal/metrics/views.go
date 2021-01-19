package metrics

import (
	"fmt"
	"strings"

	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// Views provides stats view as described by opencensus
type Views struct {
	classification string
	stats          *Stats

	// ViewCheckExecutionTime is the checks execution time aggregation tagged by check name
	ViewCheckExecutionTime *view.View

	// ViewCheckCountByNameAndStatus is the checks execution count aggregation grouped by check name, and check status
	ViewCheckCountByNameAndStatus *view.View

	// ViewCheckStatusByName is the checks status aggregation tagged by check name
	ViewCheckStatusByName *view.View

	// DefaultHealthViews are the default health check views provided by this package.
	DefaultViews []*view.View
}

func NewViews(classification string, stats *Stats) *Views {
	trimmed := strings.TrimSpace(classification)
	tagKeys := []tag.Key{keyCheck}
	if len(trimmed) > 0 {
		tagKeys = append(tagKeys, keyClassification)
	}
	viewCheckExecutionTime := &view.View{
		Measure:     stats.checkDuration,
		TagKeys:     tagKeys,
		Aggregation: view.Distribution(0, 1, 2, 3, 4, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 120, 160, 200, 250, 300, 500),
	}
	viewCheckCountByNameAndStatus := &view.View{
		Name:        fmt.Sprintf("%s/check_count_by_name_and_status", stats.prefix),
		Measure:     stats.checkStatus,
		TagKeys:     append(tagKeys, keyCheckPassing),
		Aggregation: view.Count(),
	}
	viewCheckStatusByName := &view.View{
		Name:        fmt.Sprintf("%s/check_status_by_name", stats.prefix),
		Measure:     stats.checkStatus,
		TagKeys:     tagKeys,
		Aggregation: view.LastValue(),
	}
	return &Views{
		classification:                classification,
		stats:                         stats,
		ViewCheckExecutionTime:        viewCheckExecutionTime,
		ViewCheckCountByNameAndStatus: viewCheckCountByNameAndStatus,
		ViewCheckStatusByName:         viewCheckStatusByName,
		DefaultViews: []*view.View{
			viewCheckExecutionTime,
			viewCheckCountByNameAndStatus,
			viewCheckStatusByName,
		},
	}
}
