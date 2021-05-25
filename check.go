package gosundheit

import "context"

// Check is the API for defining health checks.
// A valid check has a non empty Name() and a check (Execute()) function.
type Check interface {
	// Name is the name of the check.
	// Check names must be metric compatible.
	Name() string
	// Execute runs a single time check, and returns an error when the check fails, and an optional details object.
	// The function is expected to exit as soon as the provided Context is Done.
	Execute(ctx context.Context) (details interface{}, err error)
}
