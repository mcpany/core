// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"
)

func TestIsSecurePath_Hardened(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Null byte",
			path:    "foo\x00bar",
			wantErr: true,
		},
		{
			name:    "Mixed separators traversal 1",
			path:    "foo/..\\bar",
			wantErr: true,
		},
		{
			name:    "Mixed separators traversal 2",
			path:    "foo\\../bar",
			wantErr: true,
		},
		{
			name:    "Mixed separators valid",
			path:    "foo\\bar",
			wantErr: false,
		},
		{
			name:    "Backslash traversal",
			path:    "..\\bar",
			wantErr: true,
		},
		{
			name:    "Double backslash traversal",
			path:    "..\\..\\bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsSecurePath(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("IsSecurePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
