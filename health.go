package health

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/AppsFlyer/go-sundheit/checks"
)

const (
	maxExpectedChecks = 16
	initialResultMsg  = "didn't run yet"
	// ValAllChecks is the value used for the check tags when tagging all tests
	ValAllChecks = "all_checks"
)

var (
	keyCheck, _        = tag.NewKey("check")
	keyCheckPassing, _ = tag.NewKey("check_passing")

	mCheckStatus   = stats.Int64("health/status", "An health status (0/1 for fail/pass)", "pass/fail")
	mCheckDuration = stats.Float64("health/execute_time", "The time it took to execute a checks in ms", "ms")

	// ViewCheckExecutionTime is the checks execution time aggregation tagged by check name
	ViewCheckExecutionTime = &view.View{
		Measure:     mCheckDuration,
		TagKeys:     []tag.Key{keyCheck},
		Aggregation: view.Distribution(0, 1, 2, 3, 4, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 120, 160, 200, 250, 300, 500),
	}

	// ViewCheckCountByNameAndStatus is the checks execution count aggregation grouped by check name, and check status
	ViewCheckCountByNameAndStatus = &view.View{
		Name:        "health/check_count_by_name_and_status",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck, keyCheckPassing},
		Aggregation: view.Count(),
	}

	// ViewCheckStatusByName is the checks status aggregation tagged by check name
	ViewCheckStatusByName = &view.View{
		Name:        "health/check_status_by_name",
		Measure:     mCheckStatus,
		TagKeys:     []tag.Key{keyCheck},
		Aggregation: view.LastValue(),
	}

	// DefaultHealthViews are the default health check views provided by this package.
	DefaultHealthViews = []*view.View{
		ViewCheckCountByNameAndStatus,
		ViewCheckStatusByName,
		ViewCheckExecutionTime,
	}
)

// Health is the API for registering / deregistering health checks, and for fetching the health checks results.
type Health interface {
	// RegisterCheck registers a health check according to the given configuration.
	// Once RegisterCheck() is called, the check is scheduled to run in it's own goroutine.
	// Callers must make sure the checks complete at a reasonable time frame, or the next execution will delay.
	RegisterCheck(cfg *Config) error
	// Deregister removes a health check from this instance, and stops it's next executions.
	// If the check is running while Deregister() is called, the check may complete it's current execution.
	// Once a check is removed, it's results are no longer returned.
	Deregister(name string)
	// Results returns a snapshot of the health checks execution results at the time of calling, and the current health.
	// A system is considered healthy iff all checks are passing
	Results() (results map[string]Result, healthy bool)
	// IsHealthy returns the current health of the system.
	// A system is considered healthy iff all checks are passing.
	IsHealthy() bool
	// DeregisterAll Deregister removes all health checks from this instance, and stops their next executions.
	// It is equivalent of calling Deregister() for each currently registered check.
	DeregisterAll()
}

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

// Config defines a health Check and it's scheduling timing requirements.
type Config struct {
	// Check is the health Check to be scheduled for execution.
	Check checks.Check
	// Listener allows you to listen to check start/end events.
	Listener CheckListener
	// ExecutionPeriod is the period between successive executions.
	ExecutionPeriod time.Duration
	// InitialDelay is the time to delay first execution; defaults to zero.
	InitialDelay time.Duration
	// InitiallyPassing indicates when true, the check will be treated as passing before the first run; defaults to false
	InitiallyPassing bool
}

// New returns a new Health instance.
func New() Health {
	return &health{
		results:    make(map[string]Result, maxExpectedChecks),
		checkTasks: make(map[string]checkTask, maxExpectedChecks),
		lock:       sync.RWMutex{},
	}
}

type health struct {
	results    map[string]Result
	checkTasks map[string]checkTask
	lock       sync.RWMutex
}

func (h *health) RegisterCheck(cfg *Config) error {
	if cfg.Check == nil || cfg.Check.Name() == "" {
		return errors.Errorf("misconfigured check %v", cfg.Check)
	}

	if cfg.Listener == nil {
		cfg.Listener = noopCheckListener{}
	}

	// checks are initially failing by default, but we allow overrides...
	var initialErr error
	if !cfg.InitiallyPassing {
		initialErr = fmt.Errorf(initialResultMsg)
	}

	h.updateResult(cfg.Check.Name(), initialResultMsg, 0, initialErr, time.Now())
	h.scheduleCheck(h.createCheckTask(cfg), cfg)
	return nil
}

func (h *health) createCheckTask(cfg *Config) *checkTask {
	h.lock.Lock()
	defer h.lock.Unlock()

	task := checkTask{
		stopChan: make(chan bool, 1),
		check:    cfg.Check,
		listener: cfg.Listener,
	}
	h.checkTasks[cfg.Check.Name()] = task

	return &task
}

type checkTask struct {
	stopChan chan bool
	ticker   *time.Ticker
	check    checks.Check
	listener CheckListener
}

func (t *checkTask) stop() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

func (t *checkTask) execute() (details interface{}, duration time.Duration, err error) {
	startTime := time.Now()
	details, err = t.check.Execute()
	duration = time.Since(startTime)

	return
}

func (h *health) stopCheckTask(name string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	task := h.checkTasks[name]

	task.stop()

	delete(h.results, name)
	delete(h.checkTasks, name)
}

func (h *health) scheduleCheck(task *checkTask, cfg *Config) {
	go func() {
		// initial execution
		if !h.runCheckOrStop(task, time.After(cfg.InitialDelay)) {
			return
		}

		// scheduled recurring execution
		task.ticker = time.NewTicker(cfg.ExecutionPeriod)
		for {
			if !h.runCheckOrStop(task, task.ticker.C) {
				return
			}
		}
	}()
}

func (h *health) runCheckOrStop(task *checkTask, timerChan <-chan time.Time) bool {
	select {
	case <-task.stopChan:
		h.stopCheckTask(task.check.Name())
		return false
	case t := <-timerChan:
		h.checkAndUpdateResult(task, t)
		return true
	}
}

func (h *health) checkAndUpdateResult(task *checkTask, checkTime time.Time) {
	task.listener.OnCheckStarted(task.check.Name())
	details, duration, err := task.execute()
	result := h.updateResult(task.check.Name(), details, duration, err, checkTime)
	task.listener.OnCheckCompleted(task.check.Name(), result)
}

func (h *health) Deregister(name string) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	task, ok := h.checkTasks[name]
	if ok {
		// actual cleanup happens in the task go routine
		task.stopChan <- true
	}
}

func (h *health) DeregisterAll() {
	h.lock.RLock()
	defer h.lock.RUnlock()

	for k := range h.checkTasks {
		h.Deregister(k)
	}
}

func (h *health) Results() (results map[string]Result, healthy bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	results = make(map[string]Result, len(h.results))

	healthy = true
	for k, v := range h.results {
		results[k] = v
		healthy = healthy && v.IsHealthy()
	}

	return
}

func (h *health) IsHealthy() (healthy bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	return allHealthy(h.results)
}

func allHealthy(results map[string]Result) (healthy bool) {
	for _, v := range results {
		if !v.IsHealthy() {
			return false
		}
	}

	return true
}

func (h *health) updateResult(
	name string, details interface{}, checkDuration time.Duration, err error, t time.Time) (result Result) {

	h.lock.Lock()
	defer h.lock.Unlock()

	prevResult, ok := h.results[name]
	result = Result{
		Details:            details,
		Error:              newMarshalableError(err),
		Timestamp:          t,
		Duration:           checkDuration,
		TimeOfFirstFailure: nil,
	}

	if !result.IsHealthy() {
		if ok {
			result.ContiguousFailures = prevResult.ContiguousFailures + 1
			if prevResult.IsHealthy() {
				result.TimeOfFirstFailure = &t
			} else {
				result.TimeOfFirstFailure = prevResult.TimeOfFirstFailure
			}
		} else {
			result.ContiguousFailures = 1
			result.TimeOfFirstFailure = &t
		}
	}

	h.results[name] = result
	h.recordStats(name, result)

	return result
}

func (h *health) recordStats(checkName string, result Result) {
	thisCheckCtx := h.createMonitoringCtx(checkName, result.IsHealthy())
	stats.Record(thisCheckCtx, mCheckDuration.M(float64(result.Duration)/float64(time.Millisecond)))
	stats.Record(thisCheckCtx, mCheckStatus.M(status(result.IsHealthy()).asInt64()))

	allHealthy := allHealthy(h.results)
	allChecksCtx := h.createMonitoringCtx(ValAllChecks, allHealthy)
	stats.Record(allChecksCtx, mCheckStatus.M(status(allHealthy).asInt64()))
}

func (h *health) createMonitoringCtx(checkName string, isPassing bool) (ctx context.Context) {
	ctx, err := tag.New(context.Background(), tag.Insert(keyCheck, checkName), tag.Insert(keyCheckPassing, strconv.FormatBool(isPassing)))
	if err != nil {
		// When this happens it's a programming error caused by the line above
		log.Println("[Error] context creation failed for check ", checkName)
	}

	return
}

// Result represents the output of a health check execution.
type Result struct {
	// the details of task Result - may be nil
	Details interface{} `json:"message,omitempty"`
	// the error returned from a failed health check - nil when successful
	Error error `json:"error,omitempty"`
	// the time of the last health check
	Timestamp time.Time `json:"timestamp"`
	// the execution duration of the last check
	Duration time.Duration `json:"duration,omitempty"`
	// the number of failures that occurred in a row
	ContiguousFailures int64 `json:"contiguousFailures"`
	// the time of the initial transitional failure
	TimeOfFirstFailure *time.Time `json:"timeOfFirstFailure"`
}

func (r Result) IsHealthy() bool {
	return r.Error == nil
}

func (r Result) String() string {
	return fmt.Sprintf("Result{details: %s, err: %s, time: %s, contiguousFailures: %d, timeOfFirstFailure:%s}",
		r.Details, r.Error, r.Timestamp, r.ContiguousFailures, r.TimeOfFirstFailure)
}

type status bool

func (s status) asInt64() int64 {
	if s {
		return 1
	}
	return 0
}

type marshalableError struct {
	Message string `json:"message,omitempty"`
	Cause   error  `json:"cause,omitempty"`
}

func newMarshalableError(err error) error {
	if err == nil {
		return nil
	}

	mr := &marshalableError{
		Message: err.Error(),
	}

	cause := errors.Cause(err)
	if cause != err {
		mr.Cause = newMarshalableError(cause)
	}

	return mr
}

func (e *marshalableError) Error() string {
	return e.Message
}

type noopCheckListener struct{}

func (noop noopCheckListener) OnCheckStarted(name string) {}

func (noop noopCheckListener) OnCheckCompleted(name string, res Result) {}

// make sure noopCheckListener implements the CheckListener interface
var _ CheckListener = noopCheckListener{}
