package checks

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/pkg/errors"
)

// NewHostResolveCheck returns a Check that makes sure the provided host can resolve
// to at least `minRequiredResults` IP address within the specified timeout.
func NewHostResolveCheck(host string, timeout time.Duration, minRequiredResults int) Check {
	return NewResolveCheck(NewHostLookup(nil), host, timeout, minRequiredResults)
}

// LookupFunc is a function that is used for looking up something (in DNS) and return the resolved results count, and a possible error
type LookupFunc func(ctx context.Context, lookFor string) (resolvedCount int, err error)

// NewResolveCheck returns a Check that makes sure the `resolveThis` arg can be resolved using the `lookupFn`
// to at least `minRequiredResults` result within the specified timeout.
func NewResolveCheck(lookupFn LookupFunc, resolveThis string, timeout time.Duration, minRequiredResults int) Check {
	return &CustomCheck{
		CheckName: "resolve." + resolveThis,
		CheckFunc: func() (details interface{}, err error) {
			ctx, cancel := context.WithTimeout(context.TODO(), timeout)
			defer cancel()

			resolvedCount, err := lookupFn(ctx, resolveThis)
			details = fmt.Sprintf("[%d] ips were resolved", resolvedCount)
			if err != nil {
				return
			}
			if resolvedCount < minRequiredResults {
				err = errors.Errorf("[%s] lookup returned %d results, but requires at least %d", resolveThis, resolvedCount, minRequiredResults)
			}

			return
		},
	}
}

// NewHostLookup creates a LookupFunc that looks up host addresses
func NewHostLookup(resolver *net.Resolver) LookupFunc {
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	return func(ctx context.Context, host string) (resolvedCount int, err error) {
		addrs, err := resolver.LookupHost(ctx, host)
		resolvedCount = len(addrs)
		return
	}
}
