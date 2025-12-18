// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_ArbitraryArgsInjection_Blocked(t *testing.T) {
	// Vulnerability: A tool configured with fixed args can be manipulated to run with extra args
	// if the user provides "args" in the input, even if "args" is not in the schema.
	// FIX: This should now be BLOCKED.

	tool := &v1.Tool{
		Name:        proto.String("safe-echo"),
		Description: proto.String("Echoes a fixed string"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}
	// No parameters defined!
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"fixed"},
	}

	localTool := NewLocalCommandTool(tool, service, callDef)

	// User injects "args"
	req := &ExecutionRequest{
		ToolName: "safe-echo",
		Arguments: map[string]interface{}{
			"args": []interface{}{"hacked"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Secure behavior: "args" input is ignored because it's not in the schema.
	assert.Equal(t, "fixed\n", resultMap["stdout"])
}

func TestLocalCommandTool_AuthorizedArgs_Allowed(t *testing.T) {
	// Feature: If "args" IS defined in the schema, it should be allowed.

	tool := &v1.Tool{
		Name:        proto.String("flexible-echo"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}
	// "args" IS defined as a parameter
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"fixed"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("args"),
					Type: configv1.ParameterType_ARRAY.Enum(),
				},
			},
		},
	}

	localTool := NewLocalCommandTool(tool, service, callDef)

	// User provides "args"
	req := &ExecutionRequest{
		ToolName: "flexible-echo",
		Arguments: map[string]interface{}{
			"args": []interface{}{"extra"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// Should work because it is authorized.
	assert.Equal(t, "fixed extra\n", resultMap["stdout"])
}
