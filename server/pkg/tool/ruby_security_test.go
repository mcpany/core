// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestRubyInjection_Interpolation(t *testing.T) {
	// Vulnerability: Ruby interpolates #{...} inside double-quoted strings.
	// If the template uses double quotes, e.g. "puts \"{{input}}\"",
	// an attacker can inject #{system('id')} to execute code.
	t.Run("ruby_double_quote_should_detect_interpolation", func(t *testing.T) {
		cmd := "ruby"
		tool := createRubyCommandTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "#{system('echo pwned')}"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// This should fail with injection detected error
		assert.Error(t, err, "Should detect ruby interpolation injection")
		if err != nil {
			assert.Contains(t, err.Error(), "injection detected")
		}
	})

	// Verify simple usage still works
	t.Run("ruby_normal_usage_allowed", func(t *testing.T) {
		cmd := "ruby"
		tool := createRubyCommandTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "hello world"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		// If ruby is not installed, it will fail with "executable file not found".
		// We only care that it does NOT fail with "injection detected".
		if err != nil {
			assert.NotContains(t, err.Error(), "injection detected", "Valid input should not trigger injection detection")
		}
	})
}

func createRubyCommandTool(command string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-e", "puts \"{{input}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("input")},
			},
		},
	}
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
