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
}

// Config defines a health Check and it's scheduling timing requirements.
type Config struct {
	// Check is the health Check to be scheduled for execution.
	Check Check
	// ExecutionPeriod is the period between successive executions.
	ExecutionPeriod time.Duration
	// InitialDelay is the time to delay first execution; defaults to zero.
	InitialDelay time.Duration
	// InitiallyPassing indicates when true, the check will be treated as passing before the first run; defaults to false
	InitiallyPassing bool
}
