package gosundheit

import "github.com/AppsFlyer/go-sundheit/internal/types"

func allHealthy(results map[string]types.Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}
