package checks

// Must is a helper that wraps a call to a function returning (Check, error) and panics if the error is non-nil.
// It is intended for use in check initializations such as
//		c := checks.Must(checks.NewHTTPCheck(/*...*/))
func Must(check Check, err error) Check {
	if err != nil {
		panic(err)
	}

	return check
}
