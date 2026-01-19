// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		os       string // "windows" or "linux" or empty for current
	}{
		{
			name:     "Normal path",
			input:    "/tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "File URI",
			input:    "file:///tmp/test",
			expected: "/tmp/test",
		},
		{
			name:     "File URI with encoded spaces",
			input:    "file:///tmp/test%20folder",
			expected: "/tmp/test folder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.os != "" && runtime.GOOS != tt.os {
				t.Skip("Skipping test for different OS")
			}

			result := normalizePath(tt.input)
			expected := tt.expected

			// On Windows, filepath.Clean will convert forward slashes to backslashes
			if runtime.GOOS == "windows" {
				expected = filepath.FromSlash(expected)
                // Also, /tmp/test on Windows is essentially relative to current drive, but starts with slash.
                // filepath.Clean("/tmp/test") on Windows -> "\tmp\test"
			}

			if result != expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, expected)
			}
		})
	}
}

func TestNormalizePath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path",
			input:    "C:\\temp",
			expected: "C:\\temp",
		},
		{
			name:     "Windows file URI",
			input:    "file:///C:/temp",
			expected: "C:\\temp",
		},
		{
			name:     "Windows file URI encoded",
			input:    "file:///c%3A/temp",
			expected: "c:\\temp",
		},
		{
			name:     "Windows file URI with lower case drive",
			input:    "file:///c:/temp",
			expected: "c:\\temp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
