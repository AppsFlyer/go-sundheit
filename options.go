package gosundheit

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
