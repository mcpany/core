// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestExtractIP_Hunter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3.4", "1.2.3.4"},
		{"1.2.3.4:80", "1.2.3.4"},
		{"[::1]", "::1"},
		{"[::1]:80", "::1"},
		{"[fe80::1%eth0]", "fe80::1"},
		{"[fe80::1%eth0]:80", "fe80::1"},
		{"localhost", ""},
		{"localhost:80", ""},
		{"[2001:db8::1]bad", ""},
		{"127.0.0.1", "127.0.0.1"},
		{"::1", "::1"},
		{"2001:db8::1", "2001:db8::1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExtractIP(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestRedactDSN_Hunter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"postgres://user:password@host:5432/db", "postgres://user:[REDACTED]@host:5432/db"},
		{"postgres://user:p@ssword@host:5432/db", "postgres://user:[REDACTED]@host:5432/db"},
		{"redis://:password@host:6379", "redis://:[REDACTED]@host:6379"},
		{"redis://:p@ssword@host:6379", "redis://:[REDACTED]@host:6379"},
		{"redis://:password", "redis://:[REDACTED]"},
		{`parse "postgres://user:pass@host": invalid port ":pass"`, `parse "postgres://user:pass@host": invalid port ":[REDACTED]"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RedactDSN(tt.input)
			if tt.expected != "" {
				assert.Contains(t, got, "[REDACTED]")
				if tt.input == `parse "postgres://user:pass@host": invalid port ":pass"` {
					assert.NotContains(t, got, ":pass")
				} else {
					assert.NotContains(t, got, "password")
					assert.NotContains(t, got, "p@ssword")
				}
			}
		})
	}
}

func TestRedactJSON_Hunter(t *testing.T) {
	tests := []struct {
		name string
		input string
		wantRedacted bool
	}{
		{"simple", `{"token": "secret"}`, true},
		{"nested", `{"a": {"auth": "secret"}}`, true},
		{"array", `[{"token": "secret"}]`, true},
		{"comment", `{"token": 123} // comment`, true},
		{"escaped key", `{"to\u006ben": "secret"}`, true}, // token
		{"boundary", `{"authorization": "secret"}`, true},
		{"prefix match", `{"auth1": "val"}`, true},
		{"no match", `{"public": "val"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactJSON([]byte(tt.input))
			if tt.wantRedacted {
				assert.Contains(t, string(got), "[REDACTED]")
				assert.NotContains(t, string(got), "secret")
			} else {
				assert.NotContains(t, string(got), "[REDACTED]")
			}
		})
	}
}

func TestRedactDSN_FalsePositive_Hunter(t *testing.T) {
	// A valid URL where path contains : and @
	// This is common in some API endpoints or just random paths.
	input := "http://example.com/path/with:colon@something"
	// Should NOT be redacted because it's not a user:pass structure in the authority section.
	// url.Parse handles this correctly (User is nil).
	// But fallback regex might trigger.
	got := RedactDSN(input)
	assert.Equal(t, input, got, "Should not redact path components looking like credentials")

	input2 := "http://example.com:8080/path/with:colon@something"
	got2 := RedactDSN(input2)
	assert.Equal(t, input2, got2, "Should not redact path components looking like credentials (with port)")
}
