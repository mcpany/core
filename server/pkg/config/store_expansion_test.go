// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandRecursive_RecursionLimit verifies that the environment variable expansion
// logic prevents infinite recursion by enforcing a maximum depth.
func TestExpandRecursive_RecursionLimit(t *testing.T) {
	// Create a deeply nested variable reference chain: ${A} -> ${A}
	// Or just a loop: A=${A}
	// We pass the raw bytes to expandRecursive with an initial depth that is already near the limit,
	// or we mock the recursion behavior by constructing a string that triggers it.

	// Since we can't easily set up 101 chained env vars in the OS for this test without polluting it,
	// we will call expandRecursive with a high initial depth to simulate deep nesting.
	// But expandRecursive is private and we are in package config, so we can access it.
	// However, the signature is `func expandRecursive(b []byte, depth int) ([]byte, error)`.

	input := []byte("val")

	// Test at limit
	_, err := expandRecursive(input, maxExpandRecursionDepth)
	assert.NoError(t, err, "Depth equal to limit should pass")

	// Test exceeding limit
	_, err = expandRecursive(input, maxExpandRecursionDepth+1)
	assert.Error(t, err, "Depth exceeding limit should fail")
	assert.Contains(t, err.Error(), "recursion depth exceeded")
}

// TestSplitByCommaIgnoringBraces verifies the custom splitting logic used for parsing
// list-like environment variables (CSV-ish).
func TestSplitByCommaIgnoringBraces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple CSV",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "CSV with whitespace",
			input:    "a, b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Empty parts",
			input:    "a,,c",
			expected: []string{"a", "", "c"},
		},
		{
			name:     "Quoted strings",
			input:    `"a,b",c`,
			expected: []string{`"a,b"`, "c"},
		},
		{
			name:     "Quoted strings with spaces",
			input:    `"a, b", c`,
			expected: []string{`"a, b"`, "c"},
		},
		{
			name:     "JSON array",
			input:    `[1, 2], [3, 4]`,
			expected: []string{`[1, 2]`, `[3, 4]`},
		},
		{
			name:     "JSON object",
			input:    `{"k": "v,1"}, {"k": "v,2"}`,
			expected: []string{`{"k": "v,1"}`, `{"k": "v,2"}`},
		},
		{
			name:     "Mixed quotes and braces",
			input:    `{"key": "val,ue"}, "quoted,val", simple`,
			expected: []string{`{"key": "val,ue"}`, `"quoted,val"`, "simple"},
		},
		{
			name:     "Escaped quotes",
			input:    `"foo\"bar", baz`,
			expected: []string{`"foo\"bar"`, "baz"},
		},
		{
			name:     "Nested braces",
			input:    `{"a": [1, 2]}, b`,
			expected: []string{`{"a": [1, 2]}`, "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByCommaIgnoringBraces(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUnquoteCSV verifies the unquoting logic.
func TestUnquoteCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No quotes",
			input:    "foo",
			expected: "foo",
		},
		{
			name:     "Simple quotes",
			input:    `"foo"`,
			expected: "foo",
		},
		{
			name:     "Quotes with escaped quotes inside (doubled)",
			input:    `"foo""bar"`,
			expected: `foo"bar`,
		},
		{
			name:     "Quotes with comma",
			input:    `"foo,bar"`,
			expected: "foo,bar",
		},
		{
			name:     "Incomplete quotes (start)",
			input:    `"foo`,
			expected: `"foo`,
		},
		{
			name:     "Incomplete quotes (end)",
			input:    `foo"`,
			expected: `foo"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unquoteCSV(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestResolveEnvValue verifies mapping from environment variable strings to Protobuf fields.
func TestResolveEnvValue(t *testing.T) {
	// We use GlobalSettings as the root message for testing.

	tests := []struct {
		name          string
		path          []string
		value         string
		expectedCheck func(t *testing.T, val interface{})
	}{
		{
			name:  "Scalar string",
			path:  []string{"api_key"},
			value: "secret-key",
			expectedCheck: func(t *testing.T, val interface{}) {
				assert.Equal(t, "secret-key", val)
			},
		},
		{
			name:  "Scalar bool (true)",
			path:  []string{"use_sudo_for_docker"},
			value: "true",
			expectedCheck: func(t *testing.T, val interface{}) {
				assert.Equal(t, true, val)
			},
		},
		{
			name:  "Scalar bool (1)",
			path:  []string{"use_sudo_for_docker"},
			value: "1",
			expectedCheck: func(t *testing.T, val interface{}) {
				assert.Equal(t, true, val)
			},
		},
		{
			name:  "Scalar list (strings, CSV)",
			path:  []string{"allowed_ips"},
			value: "127.0.0.1, 192.168.1.1",
			expectedCheck: func(t *testing.T, val interface{}) {
				list, ok := val.([]interface{})
				require.True(t, ok)
				assert.Len(t, list, 2)
				assert.Equal(t, "127.0.0.1", list[0])
				assert.Equal(t, "192.168.1.1", list[1])
			},
		},
		{
			name:  "Scalar list (strings, JSON Array)",
			path:  []string{"allowed_ips"},
			value: `["10.0.0.1", "10.0.0.2"]`,
			expectedCheck: func(t *testing.T, val interface{}) {
				list, ok := val.([]interface{})
				require.True(t, ok)
				assert.Len(t, list, 2)
				assert.Equal(t, "10.0.0.1", list[0])
				assert.Equal(t, "10.0.0.2", list[1])
			},
		},
		{
			name:  "Message list (JSON Objects)",
			path:  []string{"middlewares"},
			value: `{"name": "logger", "priority": 1}, {"name": "auth", "priority": 2}`,
			expectedCheck: func(t *testing.T, val interface{}) {
				list, ok := val.([]interface{})
				require.True(t, ok)
				assert.Len(t, list, 2)

				m1, ok := list[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "logger", m1["name"])
				// JSON numbers are often float64 in generic unmarshal
				assert.Equal(t, float64(1), m1["priority"])

				m2, ok := list[1].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "auth", m2["name"])
				assert.Equal(t, float64(2), m2["priority"])
			},
		},
		{
			name:  "Message list (JSON Array of Objects)",
			path:  []string{"middlewares"},
			value: `[{"name": "logger", "priority": 1}]`,
			expectedCheck: func(t *testing.T, val interface{}) {
				list, ok := val.([]interface{})
				require.True(t, ok)
				assert.Len(t, list, 1)

				m1, ok := list[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "logger", m1["name"])
			},
		},
		{
			name: "Nested path scalar",
			path: []string{"audit", "output_path"},
			value: "/var/log/audit.log",
			expectedCheck: func(t *testing.T, val interface{}) {
				assert.Equal(t, "/var/log/audit.log", val)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := configv1.GlobalSettings_builder{}.Build()
			// resolveEnvValue signature: (root proto.Message, path []string, value string) interface{}
			// We iterate through the path to find the leaf field descriptor kind, but the function
			// uses the root message to walk the descriptors.

			// We need to pass the *actual root message* corresponding to the start of the path.
			// The tests assume we start at GlobalSettings.

			val := resolveEnvValue(root, tt.path, tt.value)
			tt.expectedCheck(t, val)
		})
	}
}

// TestResolveEnvValue_ListIndex verifies mapping when path includes array indices.
func TestResolveEnvValue_ListIndex(t *testing.T) {
	// Case: allowed_ips[0] = 1.2.3.4
	root := configv1.GlobalSettings_builder{}.Build()
	path := []string{"allowed_ips", "0"}
	val := resolveEnvValue(root, path, "1.2.3.4")
	assert.Equal(t, "1.2.3.4", val)

	// Case: middlewares[0].name = "logger"
	// resolveEnvValue returns the value for the *leaf*.
	path = []string{"middlewares", "0", "name"}
	val = resolveEnvValue(root, path, "logger")
	assert.Equal(t, "logger", val)
}
