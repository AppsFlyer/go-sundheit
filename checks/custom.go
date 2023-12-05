package checks

import (
	"context"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

// CustomCheck is a simple Check implementation if all you need is a functional check
type CustomCheck struct {
	// CheckName s the name of the check.
	CheckName string
	// CheckFunc is a function that runs a single time check, and returns an error when the check fails, and an optional details object.
	CheckFunc func(ctx context.Context) (details interface{}, err error)
}

var _ gosundheit.Check = (*CustomCheck)(nil)

// Name is the name of the check.
// Check names must be metric compatible.
func (check *CustomCheck) Name() string {
	return check.CheckName
}

// Execute runs the given Checkfunc, and return it's output.
func (check *CustomCheck) Execute(ctx context.Context) (details interface{}, err error) {
	if check.CheckFunc == nil {
		return "Unimplemented check", nil
	}

	return check.CheckFunc(ctx)
}
