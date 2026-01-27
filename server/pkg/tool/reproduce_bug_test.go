// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMixedQuotingVulnerability(t *testing.T) {
	// Vulnerability: When a parameter is used in mixed quoting contexts (Single AND Double),
	// the validator takes the minimum level (Double), allowing single quotes to pass through.
	// But single quotes are dangerous in the Single Quoted context.
	t.Run("mixed_quoting_vulnerability", func(t *testing.T) {
		cmd := "bash"
		// Template uses {{input}} in both single and double quotes
		template := "echo '{{input}}' \"{{input}}\""
		tool := createTestCommandToolWithTemplate(cmd, template)

		// Input contains a single quote.
		// In Single Quoted context: ' breaks out.
		// In Double Quoted context: ' is safe.
		// Validator (Double Quoted logic) thinks it is safe.
		// We expect this to fail validation.
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "'"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err, "Should detect shell injection in mixed quoting context")
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected")
		}
	})

	// Coverage for analyzeQuoteContext escape handling
	t.Run("escaped_quotes_in_template", func(t *testing.T) {
		cmd := "bash"
		// Template: echo \"{{input}}\"
		// The \" escapes the quote, so {{input}} is NOT inside quotes (it's unquoted)
		template := "echo \\\"{{input}}\\\""
		tool := createTestCommandToolWithTemplate(cmd, template)

		// Input contains ; which is dangerous in Unquoted.
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": ";"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err, "Should detect shell injection in unquoted context (due to escaped quotes)")
	})

	// Coverage for escaped backslash in template
	t.Run("escaped_backslash_in_template", func(t *testing.T) {
		cmd := "bash"
		// Template: echo \\{{input}}
		// First \ escapes second \. {{input}} is exposed.
		template := "echo \\\\{{input}}"
		tool := createTestCommandToolWithTemplate(cmd, template)

		// Context Unquoted.
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": ";"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	// Coverage for 'env' command special check
	t.Run("env_command_injection", func(t *testing.T) {
		cmd := "env" // isShellCommand returns true for "env"
		template := "{{input}}"
		tool := createTestCommandToolWithTemplate(cmd, template)

		// = is dangerous for env in unquoted context
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "VAR=val"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "contains dangerous character '='")
		}
	})
}
