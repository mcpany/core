// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package filesystem

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveS3Path(t *testing.T) {
	u := &Upstream{}

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"test.txt", "test.txt", false},
		{"/test.txt", "test.txt", false},
		{"folder/file.txt", "folder/file.txt", false},
		{"/folder/file.txt", "folder/file.txt", false},
		{"../file.txt", "file.txt", false},
		{"", "", true},
		{".", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%q", tt.input), func(t *testing.T) {
			got, err := u.resolveS3Path(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
