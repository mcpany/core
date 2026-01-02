// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveS3Path(t *testing.T) {
	u := &Upstream{}

	tests := []struct {
		name        string
		virtualPath string
		expectErr   bool
		expected    string
	}{
		{
			name:        "Simple path",
			virtualPath: "file.txt",
			expectErr:   false,
			expected:    "file.txt",
		},
		{
			name:        "Path with leading slash",
			virtualPath: "/file.txt",
			expectErr:   false,
			expected:    "file.txt",
		},
		{
			name:        "Nested path",
			virtualPath: "dir/file.txt",
			expectErr:   false,
			expected:    "dir/file.txt",
		},
		{
			name:        "Path traversal attempt",
			virtualPath: "../file.txt",
			expectErr:   false,
			expected:    "file.txt", // cleaned to root
		},
		{
			name:        "Empty path",
			virtualPath: "",
			expectErr:   true,
		},
		{
			name:        "Dot path",
			virtualPath: ".",
			expectErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path, err := u.resolveS3Path(tc.virtualPath)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, path)
			}
		})
	}
}
