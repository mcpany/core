// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"runtime"
	"testing"
)

func TestIsRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		enforced bool
		wantErr  bool
	}{
		// Enforced = true
		{"Enforced_Relative", "config.yaml", true, false},
		{"Enforced_Subdir", "subdir/config.yaml", true, false},
		{"Enforced_Traversal", "../config.yaml", true, true},
		{"Enforced_Absolute", "/etc/passwd", true, true},
		{"Enforced_AbsoluteWin", "C:\\Windows", true, true},
		{"Enforced_CleanRelative", "subdir/../config.yaml", true, false},

		// Enforced = false (default)
		{"Default_Relative", "config.yaml", false, false},
		{"Default_Subdir", "subdir/config.yaml", false, false},
		{"Default_Traversal", "../config.yaml", false, true}, // Still checks IsSecurePath
		{"Default_Absolute", "/etc/passwd", false, false},    // Allowed!
		{"Default_AbsoluteWin", "C:\\Windows", false, false}, // Allowed!
		{"Default_CleanRelative", "subdir/../config.yaml", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Windows path test on non-Windows
			if tt.path == "C:\\Windows" && runtime.GOOS != "windows" {
				return
			}
			// On Linux, C:\Windows is relative (filename), so Enforced_AbsoluteWin might fail expectation if we expect it to be treated as absolute.
			// filepath.IsAbs("C:\\Windows") on Linux is false.
			// So if we are on Linux, "C:\\Windows" is treated as relative.
			// So Enforced_AbsoluteWin would NOT error on Linux.
			// I should probably skip the Windows specific test if not on Windows to avoid confusion.
			if tt.path == "C:\\Windows" && runtime.GOOS != "windows" {
				return
			}

			if tt.enforced {
				t.Setenv("MCPANY_ENFORCE_RELATIVE_PATHS", "true")
			} else {
				t.Setenv("MCPANY_ENFORCE_RELATIVE_PATHS", "false")
			}

			err := IsRelativePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsRelativePath(%q) enforced=%v error = %v, wantErr %v", tt.path, tt.enforced, err, tt.wantErr)
			}
		})
	}
}
