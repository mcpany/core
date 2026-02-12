// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandInjection_SpaceInjection(t *testing.T) {
	// Case: Argument injection via space in unquoted shell command
	// This represents a user using `sh -c "curl {{input}}"`
	// If input contains a space, it splits arguments passed to curl.
	t.Run("space_argument_injection", func(t *testing.T) {
		// Ensure SSRF validation is active (not bypassed by env var from TestMain)
		t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")

		cmd := "sh"
		// Template uses unquoted placeholder inside the command string passed to -c
		// Note: The helper creates args=["-c", template].
		// So effective command is: sh -c "curl {{input}}"
		tool := createTestCommandToolWithTemplate(cmd, "curl {{input}}")

		// Input introduces new arguments to curl: -o /etc/passwd
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "http://example.com -o /etc/passwd"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// This should fail because we want to block unquoted spaces in shell commands.
		if err == nil {
			t.Log("VULNERABILITY: Unquoted space injection was allowed!")
		}
		assert.Error(t, err, "Should detect shell injection (space)")
		if err != nil {
			errStr := err.Error()
			// Accepts either "shell injection detected" (from checkUnquotedInjection)
			// OR "unsafe URL detected" / "invalid URL" if the input looks like a URL (SSRF check catches it first).
			// Both are valid security blocks.
			// We use strings.Contains to avoid failed assertion noise if one path is taken.
			isShell := strings.Contains(errStr, "shell injection detected")
			isSSRF := strings.Contains(errStr, "unsafe URL detected") || strings.Contains(errStr, "invalid URL")
			if !isShell && !isSSRF {
				t.Errorf("Unexpected error: %v", err)
			}
		}
	})

    // Case: Safe usage with quotes
    t.Run("quoted_space_safe", func(t *testing.T) {
        cmd := "sh"
        // Template uses quoted placeholder: sh -c "echo '{{input}}'"
        tool := createTestCommandToolWithTemplate(cmd, "echo '{{input}}'")

        req := &ExecutionRequest{
            ToolName: "test",
            ToolInputs: []byte(`{"input": "hello world"}`),
        }

        _, err := tool.Execute(context.Background(), req)
        assert.NoError(t, err, "Quoted input with space should be allowed")
    })
}
