// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandUsabilityBug(t *testing.T) {
	// Case 1: Python with safe argument containing semicolon
	// Python handles arguments via execve, so semicolon should be treated as data, not shell separator.
	t.Run("python_semicolon_data", func(t *testing.T) {
		cmd := "python"
		// Template passes input as a raw argument
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; bar"}`),
		}

		// Now should pass
		_, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})

	// Case 2: Python with safe argument containing single quote
	t.Run("python_quote_data", func(t *testing.T) {
		cmd := "python"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "It's me"}`),
		}

		// Now should pass
		_, err := tool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})

	// Case 3: SH (Shell) with semicolon - MUST FAIL
	// For 'sh', unquoted arguments might be interpreted if sh -c is used improperly,
	// or if the tool config treats sh arguments as shell code.
	// Since we can't distinguish, we keep strict checks for shells.
	t.Run("sh_semicolon_strict", func(t *testing.T) {
		cmd := "sh"
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; rm -rf /"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})

	// Case 4: Python with QUOTED template and breakout attempt - MUST FAIL
	// If the template quotes the argument, we must prevent breaking out of quotes.
	t.Run("python_quoted_breakout", func(t *testing.T) {
		cmd := "python"
		// Template uses single quotes
		tool := createTestCommandToolWithTemplate(cmd, "'{{input}}'")
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo'bar"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})
}
