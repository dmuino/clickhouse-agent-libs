package chagent

import (
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"syscall"
	"time"
)

func Uptime() time.Duration {
	logger := GetLogger("Uptime")
	bootTime, err := host.BootTime()
	logger.CheckErr(err)
	now := time.Now().Unix()
	elapsed := time.Duration(now-int64(bootTime)) * time.Second
	return elapsed
}

func TimeRunning(processName string) int {
	logger := GetLogger("TimeRunning")
	pids, err := process.Pids()
	if err != nil {
		logger.Fatalf("Error getting pids: %v", err)
	}
	for _, pid := range pids {
		if pid <= 1 {
			continue
		}
		p, err := process.NewProcess(pid)
		if err != nil {
			logger.Warningf("Error getting process: %v", err)
			continue
		}
		name, err := p.Name()
		if err != nil {
			if err != syscall.EINVAL {
				logger.Warningf("Error getting process name for pid %d: %v", pid, err)
			}
			continue
		}
		if name == processName {
			logger.Debugf("Found process %s with pid %d", name, pid)
			t, err := p.CreateTime()
			if err != nil {
				logger.Errorf("Error getting process start time: %v", err)
				continue
			}
			t /= 1000
			return int(time.Now().Unix() - t)
		}
	}
	return -1
}
