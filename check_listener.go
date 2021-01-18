package gosundheit

// CheckListener can be used to gain check stats or log check transitions.
// Implementations of this interface **must not block!**
// If an implementation blocks, it may result in delayed execution of other health checks down the line.
// It's OK to log in the implementation and it's OK to add metrics, but it's not OK to run anything that
// takes long time to complete such as network IO etc.
type CheckListener interface {
	// OnCheckStarted is called when a check with the specified name has started
	OnCheckStarted(name string)

	// OnCheckCompleted is called when the check with the specified name has completed it's execution.
	// The results are passed as an argument
	OnCheckCompleted(name string, result Result)
}

type noopCheckListener struct{}

func (noop noopCheckListener) OnCheckStarted(_ string) {}

func (noop noopCheckListener) OnCheckCompleted(_ string, _ Result) {}

// make sure noopCheckListener implements the CheckListener interface
var _ CheckListener = noopCheckListener{}
