package checks

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

// NewHostResolveCheck returns a gosundheit.Check that makes sure the provided host can resolve
// to at least `minRequiredResults` IP address within the timeout specified by the provided context..
func NewHostResolveCheck(host string, minRequiredResults int) gosundheit.Check {
	return NewResolveCheck(NewHostLookup(nil), host, minRequiredResults)
}

// LookupFunc is a function that is used for looking up something (in DNS) and return the resolved results count, and a possible error
type LookupFunc func(ctx context.Context, lookFor string) (resolvedCount int, err error)

// NewResolveCheck returns a gosundheit.Check that makes sure the `resolveThis` arg can be resolved using the `lookupFn`
// to at least `minRequiredResults` result, within the timeout specified by the provided context.
func NewResolveCheck(lookupFn LookupFunc, resolveThis string, minRequiredResults int) gosundheit.Check {
	return &CustomCheck{
		CheckName: "resolve." + resolveThis,
		CheckFunc: func(ctx context.Context) (details interface{}, err error) {
			resolvedCount, err := lookupFn(ctx, resolveThis)
			details = fmt.Sprintf("[%d] results were resolved", resolvedCount)
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
