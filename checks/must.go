package checks

import gosundheit "github.com/AppsFlyer/go-sundheit"

// Must is a helper that wraps a call to a function returning (gosundheit.Check, error) and panics if the error is non-nil.
// It is intended for use in check initializations such as
//
//	c := checks.Must(checks.NewHTTPCheck(/*...*/))
func Must(check gosundheit.Check, err error) gosundheit.Check {
	if err != nil {
		panic(err)
	}

	return check
}
