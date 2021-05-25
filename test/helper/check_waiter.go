package helper

import (
	"fmt"

	gosundheit "github.com/AppsFlyer/go-sundheit"
)

type CheckWaiter struct {
	completedChan chan string
}

func NewCheckWaiter() *CheckWaiter {
	return &CheckWaiter{
		completedChan: make(chan string),
	}
}

func (c *CheckWaiter) OnCheckRegistered(_ string, _ gosundheit.Result) {}

func (c *CheckWaiter) OnCheckStarted(_ string) {}

func (c *CheckWaiter) OnCheckCompleted(name string, _ gosundheit.Result) {
	c.completedChan <- name
}

func (c *CheckWaiter) AwaitChecksCompletion(checkNames ...string) error {
	if len(checkNames) == 0 {
		return nil
	}

	awaitingCompletion := make(map[string]int, len(checkNames))
	for _, c := range checkNames {
		_, ok := awaitingCompletion[c]
		if ok {
			awaitingCompletion[c]++
		} else {
			awaitingCompletion[c] = 1
		}
	}

	for chkName := range c.completedChan {
		fmt.Printf("check '%s' completed\n", chkName)
		remainingCount, ok := awaitingCompletion[chkName]
		if !ok {
			return fmt.Errorf("unexpected check completed: %s", chkName)
		}
		if remainingCount == 1 {
			delete(awaitingCompletion, chkName)
		} else {
			awaitingCompletion[chkName]--
		}

		if len(awaitingCompletion) == 0 {
			break
		}
	}

	return nil
}
