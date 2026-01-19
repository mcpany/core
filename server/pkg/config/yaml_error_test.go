// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestYamlErrorEnhancement(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		errorContains []string
	}{
		{
			name: "Malformed Indentation",
			yamlContent: `
tools:
  - name: my-tool
    type: http
    url: "https://example.com/api"
    method: "GET"
      headers: # Malformed indentation
      Authorization: "Bearer invalid"
`,
			errorContains: []string{"did not find expected key", "Hint: Check your indentation"},
		},
		{
			name: "Tabs instead of spaces",
			yamlContent: "tools:\n\t- name: test",
			errorContains: []string{"found character that cannot start any token", "tabs"},
		},
		{
			name: "Mapping values not allowed",
			yamlContent: `
foo: bar
  baz: qux
`,
			errorContains: []string{"mapping values are not allowed", "Hint: You might be trying to use a list item"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewEngine("config.yaml")
			require.NoError(t, err)

			var cfg configv1.McpAnyServerConfig
			err = engine.Unmarshal([]byte(tt.yamlContent), &cfg)
			require.Error(t, err)

            // Uncomment to debug actual error message
			// t.Logf("Error: %v", err)

			for _, expected := range tt.errorContains {
				assert.Contains(t, err.Error(), expected)
			}
		})
	}
}
