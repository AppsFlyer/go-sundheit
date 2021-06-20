package gosundheit

import (
	"time"
)

// HealthOption configures a health checker using the functional options paradigm
// popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style, see:
// - https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
// - https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
// - https://sagikazarmark.hu/blog/functional-options-on-steroids/
type HealthOption interface {
	apply(*health)
}

type healthOptionFunc func(*health)

func (fn healthOptionFunc) apply(h *health) {
	fn(h)
}

// WithCheckListeners allows you to listen to check start/end events
func WithCheckListeners(listener ...CheckListener) HealthOption {
	return healthOptionFunc(func(h *health) {
		h.checksListener = listener
	})
}

// WithHealthListeners allows you to listen to overall results change
func WithHealthListeners(listener ...HealthListener) HealthOption {
	return healthOptionFunc(func(h *health) {
		h.healthListener = listener
	})
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// This is a simple placeholder for any future defaults
func WithDefaults() HealthOption {
	return healthOptionFunc(func(h *health) {})
}

// CheckOption configures a health check using the functional options paradigm
// popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style, see:
// - https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
// - https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
// - https://sagikazarmark.hu/blog/functional-options-on-steroids/
type CheckOption interface {
	applyCheck(*checkConfig)
}

// Option configures a health checker or a health check using the functional options paradigm
// popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style, see:
// - https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html
// - https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
// - https://sagikazarmark.hu/blog/functional-options-on-steroids/
type Option interface {
	HealthOption
	CheckOption
}

type executionPeriod time.Duration

func (o executionPeriod) apply(h *health) {
	h.defaultExecutionPeriod = time.Duration(o)
}

func (o executionPeriod) applyCheck(c *checkConfig) {
	c.executionPeriod = time.Duration(o)
}

// ExecutionPeriod is the period between successive executions.
func ExecutionPeriod(d time.Duration) Option {
	return executionPeriod(d)
}

type initialDelay time.Duration

func (o initialDelay) apply(h *health) {
	h.defaultInitialDelay = time.Duration(o)
}

func (o initialDelay) applyCheck(c *checkConfig) {
	c.initialDelay = time.Duration(o)
}

// InitialDelay is the time to delay first execution; defaults to zero.
func InitialDelay(d time.Duration) Option {
	return initialDelay(d)
}

type initiallyPassing bool

func (o initiallyPassing) apply(h *health) {
	h.defaultInitiallyPassing = bool(o)
}

func (o initiallyPassing) applyCheck(c *checkConfig) {
	c.initiallyPassing = bool(o)
}

// InitiallyPassing indicates when true, the check will be treated as passing before the first run; defaults to false
func InitiallyPassing(b bool) Option {
	return initiallyPassing(b)
}

type executionTimeout time.Duration

func (o executionTimeout) applyCheck(c *checkConfig) {
	c.executionTimeout = time.Duration(o)
}

// ExecutionTimeout sets the timeout of the check.
// It is up to the check to respect the timeout, which is provided via the Context argument of `Check.Execute` method.
// Defaults to no timeout
func ExecutionTimeout(d time.Duration) CheckOption {
	return executionTimeout(d)
}
