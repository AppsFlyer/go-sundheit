package gosundheit

func allHealthy(results map[string]Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}

func copyResultMap(result map[string]Result) map[string]Result {
	newMap := make(map[string]Result)
	for k, v := range result {
		newMap[k] = v
	}
	return newMap
}
