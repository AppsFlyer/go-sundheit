package opencensus

import gosundheit "github.com/AppsFlyer/go-sundheit"

func allHealthy(results map[string]gosundheit.Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}
