package gosundheit

// CheckListener can be used to gain check stats or log check transitions.
// Implementations of this interface **must not block!**
// If an implementation blocks, it may result in delayed execution of other health checks down the line.
// It's OK to log in the implementation and it's OK to add metrics, but it's not OK to run anything that
// takes long time to complete such as network IO etc.
type CheckListener interface {
	// OnCheckRegistered is called when the check with the specified name has registered.
	// Result argument is for reporting the first run state of the check
	OnCheckRegistered(name string, result Result)

	// OnCheckStarted is called when a check with the specified name has started
	OnCheckStarted(name string)

	// OnCheckCompleted is called when the check with the specified name has completed it's execution.
	// The results are passed as an argument
	OnCheckCompleted(name string, result Result)
}

// CheckListeners is a slice of check listeners
type CheckListeners []CheckListener

// OnCheckRegistered is called when the check with the specified name has registered.
// Result argument is for reporting the first run state of the check
func (c CheckListeners) OnCheckRegistered(name string, result Result) {
	for _, listener := range c {
		listener.OnCheckRegistered(name, result)
	}
}

// OnCheckStarted is called when a check with the specified name has started
func (c CheckListeners) OnCheckStarted(name string) {
	for _, listener := range c {
		listener.OnCheckStarted(name)
	}
}

// OnCheckCompleted is called when the check with the specified name has completed it's execution.
// The results are passed as an argument
func (c CheckListeners) OnCheckCompleted(name string, result Result) {
	for _, listener := range c {
		listener.OnCheckCompleted(name, result)
	}
}
