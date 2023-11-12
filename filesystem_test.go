package chagent

import (
	"fmt"
	"testing"
)

func TestReadableSize(t *testing.T) {
	type args struct {
		size     uint64
		expected string
	}

	mb := uint64(1024 * 1024)
	gb := 1024 * mb
	tb := 1024 * gb
	pb := 1024 * tb
	eb := 1024 * pb

	tests := []args{
		{0, "0 B"},
		{1, "1 B"},
		{1024, "1.0 kB"},
		{mb, "1.0 MB"},
		{gb, "1.0 GB"},
		{tb, "1.0 TB"},
		{pb, "1.0 PB"},
		{eb, "1.0 EB"},
		{8500, fmt.Sprintf("%.1f kB", 8500.0/1024.0)},
		{15000, fmt.Sprintf("%.0f kB", 15000.0/1024.0)},
		{15000000, fmt.Sprintf("%.0f MB", 15000000.0/1024.0/1024.0)},
	}

	for _, tt := range tests {
		if got := ReadableSize(tt.size); got != tt.expected {
			t.Errorf("ReadableSize(%d) = %s, want %s", tt.size, got, tt.expected)
		}
	}
}
