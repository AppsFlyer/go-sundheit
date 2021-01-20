package gosundheit

import "github.com/AppsFlyer/go-sundheit/internal/metrics"

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

// WithMetricClassification allows you to create Health object for specific usage (liveness/readiness)
func WithMetricClassification(classification string) Option {
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

// WithStartupClassification sets the classification to "startup"
func WithStartupClassification() Option {
	return func(h *health) {
		h.classification = "startup"
	}
}

// WithStatsPrefix sets the reported metrics (and views) prefix
func WithStatsPrefix(prefix string) Option {
	return func(h *health) {
		h.stats = metrics.NewStats(prefix)
	}
}

func withDefaultStatsPrefix() Option {
	return func(h *health) {
		if h.stats == nil {
			h.stats = metrics.NewStats("health")
		}
	}
}

func withDefaultViews() Option {
	return func(h *health) {
		if h.view == nil {
			h.view = metrics.NewViews(h.classification, h.stats)
		}
	}
}

// WithDefaults sets all the Health object settings. It's not required to use this as no options is always default
// Defaults are: no check listener, stats prefix is "health", and no classification
func WithDefaults() Option {
	return func(h *health) {
		for _, opt := range []Option{
			withDefaultCheckListener(),
			withDefaultStatsPrefix(),
			withDefaultViews(),
		} {
			opt(h)
		}
	}
}
