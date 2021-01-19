package metrics

import "go.opencensus.io/tag"

const (
	// ValAllChecks is the value used for the check tags when tagging all tests
	ValAllChecks = "all_checks"
)

var (
	keyCheck, _          = tag.NewKey("check")
	keyCheckPassing, _   = tag.NewKey("check_passing")
	keyClassification, _ = tag.NewKey("classification")
)

type status bool

func (s status) asInt64() int64 {
	if s {
		return 1
	}
	return 0
}
