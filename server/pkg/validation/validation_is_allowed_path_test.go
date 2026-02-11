// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"runtime"
	"strings"
	"testing"
)

func TestIsAllowedPath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		allowList string
		wantErr   bool
	}{
		// Default behavior (No allow list)
		// Relative paths inside CWD should pass
		{"Default_Relative", "config.yaml", "", false},
		{"Default_Subdir", "subdir/config.yaml", "", false},
		// ".." is blocked by IsSecurePath
		{"Default_Traversal", "../config.yaml", "", true},
		// Absolute path NOT in CWD should fail
		{"Default_Absolute", "/etc/passwd", "", true},
		{"Default_AbsoluteWin", "C:\\Windows", "", true},
		// Clean relative path inside CWD
		{"Default_CleanRelative", "subdir/../config.yaml", "", true},

		// With Allow List
		// /etc is allowed
		{"Allowed_Absolute_InList", "/etc/passwd", "/etc", false},
		// /var is NOT allowed (list is /etc)
		{"Allowed_Absolute_NotInList", "/var/log", "/etc", true},
		// Multiple allowed paths
		{"Allowed_Multiple", "/var/log", "/etc:/var", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Windows path test on non-Windows
			if tt.path == "C:\\Windows" && runtime.GOOS != "windows" {
				return
			}
			// Skip /etc /var tests on Windows?
			if (tt.path == "/etc/passwd" || tt.path == "/var/log") && runtime.GOOS == "windows" {
				return
			}

			if tt.allowList != "" {
				// t.Setenv("MCPANY_FILE_PATH_ALLOW_LIST", tt.allowList)
				parts := strings.Split(tt.allowList, ":")
				SetAllowedPaths(parts)
			} else {
				// t.Setenv("MCPANY_FILE_PATH_ALLOW_LIST", "")
				SetAllowedPaths(nil)
			}
			t.Cleanup(func() { SetAllowedPaths(nil) })

			err := IsAllowedPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedPath(%q) allowList=%q error = %v, wantErr %v", tt.path, tt.allowList, err, tt.wantErr)
			}
		})
	}
}
