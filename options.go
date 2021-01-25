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

// WithClassification allows you to create Health object for specific usage (liveness/readiness)
func WithClassification(classification string) Option {
	return func(h *health) {
		h.classification = classification
	}
}

// WithLivenessClassification sets the classification to "liveness"
func WithLivenessClassification() Option {
	return func(h *health) {
		h.classification = "liveness"
	}
}

// WithReadinessClassification sets the classification to "readiness"
func WithReadinessClassification() Option {
	return func(h *health) {
		h.classification = "readiness"
	}
}

// WithSetupClassification sets the classification to "setup"
func WithSetupClassification() Option {
	return func(h *health) {
		h.classification = "setup"
	}
}

func withDefaultClassification() Option {
	return func(h *health) {
		if h.classification == "" {
			h.classification = "none"
		}
	}
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// Defaults are: no check listener, classification set to 'none'
func WithDefaults() Option {
	return func(h *health) {
		for _, opt := range []Option{
			withDefaultCheckListener(),
			withDefaultClassification(),
		} {
			opt(h)
		}
	}
}
