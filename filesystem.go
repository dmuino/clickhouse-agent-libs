package chagent

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
)

func ReadableSize(size uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	if size < 10 {
		return fmt.Sprintf("%d B", size)
	}
	s := float64(size)
	idx := 0
	for s >= 1024.0 && idx < len(sizes) {
		s /= 1024.0
		idx++
	}
	f := "%.0f %s"
	if s < 10.0 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, s, sizes[idx])
}

func GetFreeDiskSpace(path string) uint64 {
	logger := GetLogger("GetFreeDiskSpace")
	usage, err := disk.Usage(path)
	logger.CheckErr(err)
	logger.Infof("Free disk space on %s: %s", path, ReadableSize(usage.Free))
	return usage.Free
}
