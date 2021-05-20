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
	startTime := time.Now()
	details, err = t.check.Execute(ctx)
	duration = time.Since(startTime)

	return
}
