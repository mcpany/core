// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"
)

func TestIsSecurePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid file", "test.txt", false},
		{"valid path", "path/to/test.txt", false},
		{"valid with dots in name", "my..file.txt", false},
		{"valid with dots in dir name", "my..dir/file.txt", false},
		{"traversal up", "../test.txt", true},
		{"traversal up nested", "dir/../../test.txt", true},
		{"traversal up double", "../../test.txt", true},
		// Absolute paths are allowed, as long as they don't result in relative traversal up
		{"absolute path", "/etc/passwd", false},
		{"absolute path with resolved traversal", "/var/../etc/passwd", true},
		{"traversal resolved within path", "safe/../safe/file.txt", true},
		{"absolute traversal out", "/var/www/../../etc/passwd", true},
		{"current dir prefix", "./file.txt", false},
		{"multiple slashes", "safe//file.txt", false},
		{"just dot dot", "..", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSecurePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSecurePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestIsSecureRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid file", "test.txt", false},
		{"valid path", "path/to/test.txt", false},
		{"valid with dots in name", "my..file.txt", false},
		{"valid with dots in dir name", "my..dir/file.txt", false},
		{"traversal up", "../test.txt", true},
		{"traversal up nested", "dir/../../test.txt", true},
		{"absolute path", "/etc/passwd", true},
		{"absolute path with resolved traversal", "/var/../etc/passwd", true},
		{"current dir prefix", "./file.txt", false},
		{"multiple slashes", "safe//file.txt", false},
		{"leading slash", "/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSecureRelativePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSecureRelativePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
