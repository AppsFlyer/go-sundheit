package gosundheit

type Option func(*health)

// WithCheckListener allows you to listen to check start/end events
func WithCheckListener(listener CheckListener) Option {
	return func(h *health) {
		if listener != nil {
			h.checksListener = listener
		}
	}
}

func withDefaultCheckListener() Option {
	return func(h *health) {
		if h.checksListener == nil {
			h.checksListener = noopCheckListener{}
		}
	}
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// Defaults are: no check listener
func WithDefaults() Option {
	return func(h *health) {
		for _, opt := range []Option{
			withDefaultCheckListener(),
		} {
			opt(h)
		}
	}
}
