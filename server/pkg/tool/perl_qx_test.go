// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPerlQXInjection(t *testing.T) {
	// Setup: A tool running perl that echoes its input
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("perl"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print \"{{input}}\""}, // Double-quoted context
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("perl_echo"),
	}.Build()

	cmdTool := tool.NewCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"perl-echo-call",
	)

	ctx := context.Background()

	t.Run("Valid input with qx substring should pass", func(t *testing.T) {
		// "foobqxbox" contains "qx" but not as an operator (part of word)
		inputData := map[string]interface{}{"input": "foobqxbox"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{
			ToolName:   "perl_echo",
			ToolInputs: inputs,
		}

		_, err := cmdTool.Execute(ctx, req)
		if err != nil {
			assert.NotContains(t, err.Error(), "perl qx execution", "Valid input 'foobqxbox' should not be blocked")
		}
	})

	t.Run("Actual qx usage should be blocked", func(t *testing.T) {
		// "qx/ls/" is dangerous
		inputData := map[string]interface{}{"input": "qx/ls/"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{
			ToolName:   "perl_echo",
			ToolInputs: inputs,
		}
		_, err := cmdTool.Execute(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "perl qx execution", "qx/ls/ should be blocked")
	})

	t.Run("Interpolated qx usage should be blocked", func(t *testing.T) {
		// "@{[ qx/ls/ ]}" is dangerous
		inputData := map[string]interface{}{"input": "@{[ qx/ls/ ]}"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{
			ToolName:   "perl_echo",
			ToolInputs: inputs,
		}
		_, err := cmdTool.Execute(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "perl qx execution")
	})
}
