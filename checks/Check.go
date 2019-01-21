package checks

// Check is the API for defining health checks.
// A valid check has a non empty Name() and a check (Execute()) function.
type Check interface {
	// Name is the name of the check.
	// Check names must be metric compatible.
	Name() string
	// Execute runs a single time check, and returns an error when the check fails, and an optional details object.
	Execute() (details interface{}, err error)
}
