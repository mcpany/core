// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Helper to create a string pointer
func strPtrSafe(s string) *string { return &s }

func TestInterpreterInjection_PHP(t *testing.T) {
	// Setup PHP tool
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: strPtrSafe("php"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: strPtrSafe("code")}.Build(),
			}.Build(),
		},
	}.Build()

	// Input schema needed for validation logic in NewLocalCommandTool or Execute
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        strPtrSafe("php_eval"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "system call unquoted",
			payload: "system('id')",
		},
		{
			name:    "exec call unquoted",
			payload: "exec('ls')",
		},
		{
			name:    "phpinfo unquoted",
			payload: "phpinfo()",
		},
		{
			name:    "backticks",
			payload: "`id`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputData := map[string]interface{}{"code": tc.payload}
			inputs, err := json.Marshal(inputData)
			require.NoError(t, err)

			req := &ExecutionRequest{ToolName: "php_eval", ToolInputs: inputs}
			_, err = cmdTool.Execute(context.Background(), req)

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "injection detected")
			}
		})
	}
}

func TestInterpreterInjection_PHP_Interpolation(t *testing.T) {
	// Setup PHP tool with double quoted argument
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: strPtrSafe("php"),
	}).Build()

	// Simulating: php -r "echo '{{code}}';"
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-r", "\"echo '{{code}}';\""},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: strPtrSafe("code")}.Build(),
			}.Build(),
		},
	}.Build()

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        strPtrSafe("php_echo"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: ${system('id')}
	// In double quotes, PHP interpolates variables.
	payload := "${system('id')}"

	inputData := map[string]interface{}{"code": payload}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &ExecutionRequest{ToolName: "php_echo", ToolInputs: inputs}
	_, err = cmdTool.Execute(context.Background(), req)

	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "injection detected")
	}
}

func TestInterpreterInjection_Expect(t *testing.T) {
	// Setup Expect tool
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: strPtrSafe("expect"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: strPtrSafe("script")}.Build(),
			}.Build(),
		},
	}.Build()

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"script": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        strPtrSafe("expect_run"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "spawn sh",
			payload: "spawn sh",
		},
		{
			name:    "system call",
			payload: "system ls",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputData := map[string]interface{}{"script": tc.payload}
			inputs, err := json.Marshal(inputData)
			require.NoError(t, err)

			req := &ExecutionRequest{ToolName: "expect_run", ToolInputs: inputs}
			_, err = cmdTool.Execute(context.Background(), req)

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "injection detected")
			}
		})
	}
}

func TestInterpreterInjection_Lua(t *testing.T) {
	// Setup Lua tool
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: strPtrSafe("lua"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: strPtrSafe("code")}.Build(),
			}.Build(),
		},
	}.Build()

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"code": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := pb.Tool_builder{
		Name:        strPtrSafe("lua_eval"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	testCases := []struct {
		name    string
		payload string
	}{
		{
			name:    "os.execute",
			payload: "os.execute('echo safe')",
		},
		{
			name:    "io.popen",
			payload: "io.popen('echo safe')",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputData := map[string]interface{}{"code": tc.payload}
			inputs, err := json.Marshal(inputData)
			require.NoError(t, err)

			req := &ExecutionRequest{ToolName: "lua_eval", ToolInputs: inputs}
			_, err = cmdTool.Execute(context.Background(), req)

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), "injection detected")
			}
		})
	}
}
