package gosundheit

import (
	"time"

	"github.com/AppsFlyer/go-sundheit/checks"
)

type checkTask struct {
	stopChan chan bool
	ticker   *time.Ticker
	check    checks.Check
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
