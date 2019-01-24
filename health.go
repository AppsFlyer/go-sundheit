package health

import (
	"time"
	"sync"
	"fmt"
	"log"

	"github.com/pkg/errors"

	"gitlab.appsflyer.com/Architecture/af-go-health/checks"
)

const (
	maxExpectedChecks = 16
	initialResultMsg  = "didn't run yet"
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

// Config defines a health Check and it's scheduling timing requirements.
type Config struct {
	// Check is the health Check to be scheduled for execution.
	Check checks.Check
	// ExecutionPeriod is the period between successive executions.
	ExecutionPeriod time.Duration
	// InitialDelay is the time to delay first execution; defaults to zero.
	InitialDelay time.Duration
}

// Result represents the output of a health check execution.
type Result interface {
	fmt.Stringer
	// IsHealthy returns true iff the corresponding health check has passed
	IsHealthy() bool
}

// New returns a new Health instance.
func New() Health {
	return &health{
		results:    make(map[string]*result, maxExpectedChecks),
		checkTasks: make(map[string]checkTask, maxExpectedChecks),
		lock:       sync.RWMutex{},
	}
}

type health struct {
	results    map[string]*result
	checkTasks map[string]checkTask
	lock       sync.RWMutex
}

func (h *health) RegisterCheck(cfg *Config) error {
	if cfg.Check == nil || cfg.Check.Name() == "" {
		err := errors.Errorf("misconfigured check %s", cfg.Check)
		log.Println(err)
		return err
	}
	// checks are initially failing...
	h.updateResult(cfg.Check.Name(), initialResultMsg, fmt.Errorf(initialResultMsg), time.Now())
	h.scheduleCheck(h.createCheckTask(cfg), cfg)
	return nil
}

func (h *health) createCheckTask(cfg *Config) *checkTask {
	h.lock.Lock()
	defer h.lock.Unlock()

	task := checkTask{
		stopChan: make(chan bool, 1),
		check:    cfg.Check,
	}
	h.checkTasks[cfg.Check.Name()] = task

	return &task
}

type checkTask struct {
	stopChan chan bool
	ticker   *time.Ticker
	check    checks.Check
}

func (h *health) stopCheckTask(name string) {
	log.Println("Cleaning task ", name)

	h.lock.Lock()
	defer h.lock.Unlock()

	task := h.checkTasks[name]
	if task.ticker != nil {
		task.ticker.Stop()
	}
	delete(h.results, name)
	delete(h.checkTasks, name)
	log.Printf("Task '%s' Stopped\n", name)
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
		h.checkAndUpdateResult(task.check, t)
		return true
	}
}

func (h *health) checkAndUpdateResult(check checks.Check, time time.Time) {
	log.Printf("%s running task '%s'...\n", time, check.Name())
	details, err := check.Execute()
	if err != nil {
		log.Printf("Check '%s' failed: [%s]", check.Name(), err)
	}
	h.updateResult(check.Name(), details, err, time)
}

func (h *health) Deregister(name string) {
	log.Printf("Stopping task '%s'", name)

	h.lock.RLock()
	defer h.lock.RUnlock()

	task, ok := h.checkTasks[name]
	if ok {
		// actual cleanup happens in the task go routine
		task.stopChan <- true
	}
}

func (h *health) DeregisterAll() {
	log.Printf("Stopping health instance")

	h.lock.RLock()
	defer h.lock.RUnlock()

	for k := range h.checkTasks {
		h.Deregister(k)
	}
}

func (h *health) Results() (results map[string]Result, healthy bool) {
	results = make(map[string]Result)
	h.lock.RLock()
	defer h.lock.RUnlock()

	healthy = true
	for k, v := range h.results {
		results[k] = v
		healthy = healthy && v.IsHealthy()
	}

	return
}

func (h *health) IsHealthy() (healthy bool) {
	h.lock.Lock()
	defer h.lock.Unlock()

	healthy = true
	for _, v := range h.results {
		healthy = healthy && v.IsHealthy()
	}

	return
}

func (h *health) updateResult(name string, details interface{}, err error, t time.Time) {
	h.lock.Lock()
	defer h.lock.Unlock()

	prevResult, ok := h.results[name]
	result := &result{
		Details:            details,
		Error:              newMarshalableError(err),
		Timestamp:          t,
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
}

type result struct {
	// the details of task result - may be nil
	Details interface{} `json:"message,omitempty"`
	// the error returned from a failed health check - nil when successful
	Error error `json:"error,omitempty"`
	// the time of the last health check
	Timestamp time.Time `json:"timestamp"`
	// the number of failures that occurred in a row
	ContiguousFailures int64 `json:"contiguousFailures"`
	// the time of the initial transitional failure
	TimeOfFirstFailure *time.Time `json:"timeOfFirstFailure"`
}

func (r *result) IsHealthy() bool {
	return r.Error == nil
}

func (r *result) String() string {
	return fmt.Sprintf("Result{details: %s, err: %s, time: %s, contiguousFailures: %d, timeOfFirstFailure:%s}",
		r.Details, r.Error, r.Timestamp, r.ContiguousFailures, r.TimeOfFirstFailure)
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
