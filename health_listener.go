package gosundheit

type HealthListener interface {
	OnResultsUpdated(results map[string]Result)
}

type HealthListeners []HealthListener

func (h HealthListeners) OnResultsUpdated(results map[string]Result) {
	for _, listener := range h {
		listener.OnResultsUpdated(results)
	}
}
