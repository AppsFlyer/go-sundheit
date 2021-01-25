package gosundheit

import (
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
)

// Config defines a health Check and it's scheduling timing requirements.
type Config struct {
	// Check is the health Check to be scheduled for execution.
	Check checks.Check
	// ExecutionPeriod is the period between successive executions.
	ExecutionPeriod time.Duration
	// InitialDelay is the time to delay first execution; defaults to zero.
	InitialDelay time.Duration
	// InitiallyPassing indicates when true, the check will be treated as passing before the first run; defaults to false
	InitiallyPassing bool
}
