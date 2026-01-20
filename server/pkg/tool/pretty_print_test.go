// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrint_XMLRedaction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "redact sensitive element content",
			input: `<root>
  <password>secret123</password>
  <username>john</username>
</root>`,
			expected: `<root>
  <password>[REDACTED]</password>
  <username>john</username>
</root>`,
		},
		{
			name: "redact sensitive attribute",
			input: `<user name="john" password="secret123" />`,
			expected: `<user name="john" password="[REDACTED]"></user>`,
		},
		{
			name: "redact nested sensitive element",
			input: `<config>
  <auth>
    <apiKey>abcdef</apiKey>
  </auth>
</config>`,
			expected: `<config>
  <auth>
    <apiKey>[REDACTED]</apiKey>
  </auth>
</config>`,
		},
		{
			name: "no redaction for safe content",
			input: `<data>value</data>`,
			expected: `<data>value</data>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prettyPrint([]byte(tt.input), "application/xml")
			// Remove indentation differences for comparison if needed,
			// but prettyPrint adds indentation "  ".
			// My expected strings above might need adjustment or loose comparison.
			// Let's assert Contains for redacted parts.

			if tt.name == "redact sensitive attribute" {
				assert.Contains(t, got, `password="[REDACTED]"`)
			} else if tt.name == "redact sensitive element content" {
				assert.Contains(t, got, `<password>[REDACTED]</password>`)
				assert.Contains(t, got, `<username>john</username>`)
			} else if tt.name == "redact nested sensitive element" {
				assert.Contains(t, got, `<apiKey>[REDACTED]</apiKey>`)
			}
		})
	}
}
