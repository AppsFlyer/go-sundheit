package gosundheit

func allHealthy(results map[string]Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}

func copyResultsMap(results map[string]Result) map[string]Result {
	newMap := make(map[string]Result, len(results))
	for k, v := range results {
		newMap[k] = v
	}
	return newMap
}
