// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
)

func TestRedactJSON_SecurityCoverage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		covered  bool // Expect to be redacted
	}{
		{
			name:    "client_secret (covered by 'secret')",
			input:   `{"client_secret": "sensitive"}`,
			want:    `{"client_secret": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "access_token (covered by 'token')",
			input:   `{"access_token": "sensitive"}`,
			want:    `{"access_token": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "refresh_token (covered by 'token')",
			input:   `{"refresh_token": "sensitive"}`,
			want:    `{"refresh_token": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "access_key (covered by 'access_key')",
			input:   `{"access_key": "sensitive"}`,
			want:    `{"access_key": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "aws_access_key_id (covered by 'access_key')",
			input:   `{"aws_access_key_id": "sensitive"}`,
			want:    `{"aws_access_key_id": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "session_id (covered by 'session_id')",
			input:   `{"session_id": "sensitive"}`,
			want:    `{"session_id": "[REDACTED]"}`,
			covered: true,
		},
		{
			name:    "csrf_token (covered by 'csrf')",
			input:   `{"csrf_token": "sensitive"}`,
			want:    `{"csrf_token": "[REDACTED]"}`,
			covered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RedactJSON([]byte(tt.input)))
			if tt.covered {
				if got != tt.want {
					t.Errorf("RedactJSON() = %v, want %v", got, tt.want)
				}
			} else {
				// verify it is NOT redacted (sanity check)
				if got == tt.want {
					t.Logf("Unexpectedly redacted: %v", tt.name)
				}
			}
		})
	}
}
