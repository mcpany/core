// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRedactor_RedactJSON_Optimization(t *testing.T) {
	enabled := true
	defaultConfig := &configv1.DLPConfig{
		Enabled: &enabled,
	}
	customConfig := &configv1.DLPConfig{
		Enabled:        &enabled,
		CustomPatterns: []string{"safe"},
	}

	tests := []struct {
		name     string
		config   *configv1.DLPConfig
		input    string
		expected string
	}{
		{
			name:     "FastPath_SafeString",
			config:   defaultConfig,
			input:    `{"key": "safe string"}`,
			expected: `{"key": "safe string"}`,
		},
		{
			name:     "SlowPath_Email",
			config:   defaultConfig,
			input:    `{"key": "user@example.com"}`,
			expected: `{"key": "***REDACTED***"}`,
		},
		{
			name:     "SlowPath_CreditCard",
			config:   defaultConfig,
			input:    `{"key": "1234-5678-1234-5678"}`,
			expected: `{"key": "***REDACTED***"}`,
		},
		{
			name:     "SlowPath_Escape",
			config:   defaultConfig,
			input:    `{"key": "line\nbreak"}`,
			expected: `{"key": "line\nbreak"}`,
		},
		{
			name:     "SkippedFastPath_CustomPattern",
			config:   customConfig,
			input:    `{"key": "safe"}`,
			expected: `{"key": "***REDACTED***"}`, // matches "safe" custom pattern
		},
		{
			name:     "Complex_Mixed",
			config:   defaultConfig,
			input:    `{"safe": "hello", "unsafe": "me@here.com", "escaped": "a\nb"}`,
			expected: `{"safe": "hello", "unsafe": "***REDACTED***", "escaped": "a\nb"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRedactor(tt.config, nil)
			out, err := r.RedactJSON([]byte(tt.input))
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(out))
		})
	}
}
