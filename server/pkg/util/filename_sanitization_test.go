// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename_WindowsReserved(t *testing.T) {
	reserved := []struct {
		input    string
		expected string // Expected suffix or exact match logic
	}{
		{"CON", "CON_file"},
		{"PRN", "PRN_file"},
		{"AUX", "AUX_file"},
		{"NUL", "NUL_file"},
		{"COM1", "COM1_file"},
		{"COM2", "COM2_file"},
		{"COM9", "COM9_file"},
		{"LPT1", "LPT1_file"},
		{"LPT2", "LPT2_file"},
		{"LPT9", "LPT9_file"},
		{"con", "con_file"},
		{"aux", "aux_file"},
		{"CON.txt", "CON_file.txt"},
		{"NUL.tar.gz", "NUL_file.tar.gz"}, // Check extension preservation
		// Non-reserved checks
		{"CONTROL", "CONTROL"},
		{"COM0", "COM0"}, // COM0 is typically not reserved
		{"LPT10", "LPT10"},
	}

	for _, tt := range reserved {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			assert.Equal(t, tt.expected, got, "SanitizeFilename(%q) should be sanitized to avoid Windows reserved name collision", tt.input)

			// Additional check to ensure we don't accidentally create another reserved name (unlikely but good to check)
			base := got
			if idx := strings.IndexByte(got, '.'); idx != -1 {
				base = got[:idx]
			}
			// Should NOT match reserved list logic again (which we know it won't because of _file suffix)
			assert.False(t, strings.EqualFold(base, "CON"), "Result should not be CON")
		})
	}
}
