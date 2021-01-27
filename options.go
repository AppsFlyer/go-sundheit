package gosundheit

type Option func(*health)

// WithCheckListener allows you to listen to check start/end events
func WithCheckListener(listener ...CheckListener) Option {
	return func(h *health) {
		h.checksListener = listener
	}
}

func WithHealthListener(listener ...HealthListener) Option {
	return func(h *health) {
		h.healthListener = listener
	}
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// This is a simple placeholder for any future defaults
func WithDefaults() Option {
	return func(h *health) {}
}
