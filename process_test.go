package chagent

import (
	"testing"
	"time"
)

func Test_timeRunning(t *testing.T) {
	time.Sleep(1 * time.Second)
	if TimeRunning("make") < 0 {
		t.Errorf("timeRunning(iTerm2) should be positive")
	}
}
