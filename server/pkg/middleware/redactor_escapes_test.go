// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Escapes(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`foo"bar`, `magic`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Escaped quote matching pattern",
			input:    `{"val": "foo\"bar"}`,
			expected: `{"val": "***REDACTED***"}`,
		},
		{
			name:     "Escaped unicode matching pattern",
			input:    `{"val": "\u006dagic"}`, // "magic"
			expected: `{"val": "***REDACTED***"}`,
		},
		{
			name:     "No escapes safe string",
			input:    `{"val": "safe"}`,
			expected: `{"val": "safe"}`,
		},
		{
			name:     "Escaped backslash",
			input:    `{"val": "safe\\back"}`,
			expected: `{"val": "safe\\back"}`,
		},
		{
			name:     "Mixed escaped and non-escaped",
			input:    `{"a": "magic", "b": "\u006dagic"}`,
			expected: `{"a": "***REDACTED***", "b": "***REDACTED***"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := r.RedactJSON([]byte(tt.input))
			assert.NoError(t, err)
			// Normalize JSON for comparison if needed, but here simple string compare might work if spacing is preserved
			// WalkStandardJSONStrings preserves structure.
			assert.JSONEq(t, tt.expected, string(output))
		})
	}
}
