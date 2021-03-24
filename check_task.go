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
	ctx, cancel := context.WithTimeout(t.parentCtx, t.timeout)
	defer cancel()
	details, err = t.check.Execute(ctx)
	duration = time.Since(startTime)

	return
}
