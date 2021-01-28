package gosundheit

type Option func(*health)

// WithCheckListeners allows you to listen to check start/end events
func WithCheckListeners(listener ...CheckListener) Option {
	return func(h *health) {
		h.checksListener = listener
	}
}

// WithHealthListeners allows you to listen to overall results change
func WithHealthListeners(listener ...HealthListener) Option {
	return func(h *health) {
		h.healthListener = listener
	}
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// This is a simple placeholder for any future defaults
func WithDefaults() Option {
	return func(h *health) {}
}
