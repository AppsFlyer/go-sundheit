package gosundheit

import (
	"time"
)

// checkConfig configures a health Check and it's scheduling timing requirements.
type checkConfig struct {
	// executionPeriod is the period between successive executions.
	executionPeriod time.Duration

	// initialDelay is the time to delay first execution; defaults to zero.
	initialDelay time.Duration

	// initiallyPassing indicates when true, the check will be treated as passing before the first run; defaults to false
	initiallyPassing bool

	// executionTimeout is the maximum allowed execution time for a check. If this timeout is exceeded, the provided Context will be cancelled.
	// defaults to no timeout.
	executionTimeout time.Duration
}
