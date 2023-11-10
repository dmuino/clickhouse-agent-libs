package chagent

import (
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"syscall"
	"time"
)

func Uptime() time.Duration {
	bootTime, err := host.BootTime()
	logger.CheckErr(err)

	uptime, err := host.Uptime()
	logger.CheckErr(err)

	return time.Duration(uptime - bootTime)
}

func TimeRunning(processName string) int {
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
			logger.Fatalf("Error getting process: %v", err)
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
				logger.Fatalf("Error getting process start time: %v", err)
			}
			t /= 1000
			return int(time.Now().Unix() - t)
		}
	}
	return -1
}
