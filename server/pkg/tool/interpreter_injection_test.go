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

func TestInterpreterInjection(t *testing.T) {
	// Vulnerability: Ruby interpolates #{...} inside double-quoted strings.
	t.Run("ruby_double_quote_should_detect_interpolation", func(t *testing.T) {
		cmd := "ruby"
		tool := createTestTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "#{system('echo pwned')}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err, "Should detect ruby interpolation injection")
		if err != nil {
			assert.Contains(t, err.Error(), "injection detected")
		}
	})

	// irb is also vulnerable
	t.Run("irb_double_quote_should_detect_interpolation", func(t *testing.T) {
		cmd := "irb"
		tool := createTestTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "#{system('echo pwned')}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err, "Should detect irb interpolation injection")
		if err != nil {
			assert.Contains(t, err.Error(), "injection detected")
		}
	})

	t.Run("ruby_normal_usage_allowed", func(t *testing.T) {
		cmd := "ruby"
		tool := createTestTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "hello world"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if err != nil {
			assert.NotContains(t, err.Error(), "injection detected")
		}
	})

	// Vulnerability: Perl interpolates @... (arrays) inside double-quoted strings.
	// We block '@' in arguments for perl commands if double quoted.
	t.Run("perl_double_quote_should_detect_interpolation", func(t *testing.T) {
		cmd := "perl"
		tool := createTestTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "@{[system('echo pwned')]}"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err, "Should detect perl interpolation injection")
		if err != nil {
			assert.Contains(t, err.Error(), "injection detected")
		}
	})

	t.Run("perl_normal_usage_allowed", func(t *testing.T) {
		cmd := "perl"
		tool := createTestTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test",
			ToolInputs: []byte(`{"input": "hello world"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if err != nil {
			assert.NotContains(t, err.Error(), "injection detected")
		}
	})
}

func createTestTool(command string) Tool {
	toolDef := &v1.Tool{Name: proto.String("test-tool")}
	service := &configv1.CommandLineUpstreamService{
		Command: &command,
	}
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-e", "print \"{{input}}\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("input")},
			},
		},
	}
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}
