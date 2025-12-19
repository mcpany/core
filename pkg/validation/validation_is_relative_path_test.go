package validation

import (
	"runtime"
	"testing"
)

func TestIsRelativePath(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"config.yaml", false},
		{"subdir/config.yaml", false},
		{"../config.yaml", true},         // Traversal
		{"/etc/passwd", true},            // Absolute
		{"C:\\Windows", true},            // Absolute (on Windows, but let's see)
		{"subdir/../config.yaml", false}, // Clean resolves this to config.yaml (safe)
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Skip Windows path test on non-Windows
			if tt.path == "C:\\Windows" && runtime.GOOS != "windows" {
				return
			}

			err := IsRelativePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsRelativePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
