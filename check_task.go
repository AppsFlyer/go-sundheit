package gosundheit

import (
	"context"
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
)

type checkTask struct {
	parentCtx context.Context
	timeout   time.Duration
	stopChan  chan bool
	ticker    *time.Ticker
	check     checks.Check
}

func (t *checkTask) stop() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

func (t *checkTask) execute() (details interface{}, duration time.Duration, err error) {
	startTime := time.Now()
	ctx, cancel := getContext(t.parentCtx, t.timeout)
	defer cancel()
	details, err = t.check.Execute(ctx)
	duration = time.Since(startTime)

	return
}

func getContext(parent context.Context, t time.Duration) (context.Context, context.CancelFunc) {
	if t == 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, t)
}
