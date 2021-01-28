package opencencus

type Option func(*MetricsListener)

// WithClassification set custom classification for metrics
func WithClassification(classification string) Option {
	return func(listener *MetricsListener) {
		listener.classification = classification
	}
}

// WithLivenessClassification sets the classification to "liveness"
func WithLivenessClassification() Option {
	return func(listener *MetricsListener) {
		listener.classification = "liveness"
	}
}

// WithReadinessClassification sets the classification to "readiness"
func WithReadinessClassification() Option {
	return func(listener *MetricsListener) {
		listener.classification = "readiness"
	}
}

// WithStartupClassification sets the classification to "startup"
func WithStartupClassification() Option {
	return func(listener *MetricsListener) {
		listener.classification = "startup"
	}
}

func WithDefaults() Option {
	return func(listener *MetricsListener) {
		for _, opt := range []Option{} {
			opt(listener)
		}
	}
}
