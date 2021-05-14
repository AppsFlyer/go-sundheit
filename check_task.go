package gosundheit

import (
	"context"
	"time"
)

type checkTask struct {
	stopChan chan bool
	ticker   *time.Ticker
	check    Check
}

func (t *checkTask) stop() {
	if t.ticker != nil {
		t.ticker.Stop()
	}
}

func (t *checkTask) execute(ctx context.Context) (details interface{}, duration time.Duration, err error) {
	startTime := time.Now()
	//ctx, cancel := getContext(t.parentCtx, t.timeout)
	//defer cancel()
	details, err = t.check.Execute(ctx)
	duration = time.Since(startTime)

	return
}
