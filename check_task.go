package gosundheit

import (
	"context"
	"time"
)

type checkTask struct {
	stopChan chan bool
	ticker   *time.Ticker
	check    Check
	timeout  time.Duration
}

func (t *checkTask) stop() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

func (t *checkTask) execute(ctx context.Context) (details interface{}, duration time.Duration, err error) {
	timeoutCtx, cancel := contextWithTimeout(ctx, t.timeout)
	defer cancel()
	startTime := time.Now()
	details, err = t.check.Execute(timeoutCtx)
	duration = time.Since(startTime)

	return
}

func contextWithTimeout(parent context.Context, t time.Duration) (context.Context, context.CancelFunc) {
	if t <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, t)
}
