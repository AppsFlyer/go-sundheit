package checks

// A simple Check implementation if all you need is a functional check
type CustomCheck struct {
	// CheckName s the name of the check.
	CheckName string
	// CheckFunc is a function that runs a single time check, and returns an error when the check fails, and an optional details object.
	CheckFunc func() (details interface{}, err error)
}

var _ Check = (*CustomCheck)(nil)

func (check *CustomCheck) Name() string {
	return check.CheckName
}

func (check *CustomCheck) Execute() (details interface{}, err error) {
	if check.CheckFunc == nil {
		return "Unimplemented check", nil
	}

	return check.CheckFunc()
}
