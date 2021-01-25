package gosundheit

func allHealthy(results map[string]Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}
