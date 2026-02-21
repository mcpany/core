package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitivePath_SystemDirectories(t *testing.T) {
	tests := []struct {
		path    string
		blocked bool
	}{
		{"/proc/self/environ", true},
		{"/proc/cpuinfo", true},
		{"/sys/kernel/debug", true},
		{"/dev/zero", true},
		{"/dev/null", true},
		{"/proc", true},
		{"/sys", true},
		{"/dev", true},
		{"proc/foo", false}, // Relative path starting with proc is fine unless it resolves to absolute /proc (which IsAllowedPath checks)
		{"/var/log", false},
		{"/tmp", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			err := IsSensitivePath(tt.path)
			if tt.blocked {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "access to system directory")
			} else {
				// It might fail for other reasons (e.g. sensitive file name), but here we check system dir block
				if err != nil {
					assert.NotContains(t, err.Error(), "access to system directory")
				}
			}
		})
	}
}
