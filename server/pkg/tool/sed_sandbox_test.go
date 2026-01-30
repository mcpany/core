// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSedSandbox_Prevention(t *testing.T) {
	// Case 1: sed (Protected)
	t.Run("sed_should_be_protected", func(t *testing.T) {
		cmd := "sed"
		// Use unquoted template to trigger strict checking
		tool := createTestCommandToolWithTemplate(cmd, "{{input}}")
		// Input with semicolon should be blocked if detected as shell
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "foo; bar"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "shell injection detected", "sed should be protected")
		}
	})
}
