package gosundheit

import (
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

func (t *checkTask) execute() (details interface{}, duration time.Duration, err error) {
	startTime := time.Now()
	details, err = t.check.Execute()
	duration = time.Since(startTime)

	return
}
