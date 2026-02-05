// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func newCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = configv1.CommandLineCallDefinition_builder{}.Build()
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String(command),
	}).Build()

	properties := make(map[string]*structpb.Value)
	for _, param := range callDef.GetParameters() {
		properties[param.GetSchema().GetName()] = structpb.NewStructValue(&structpb.Struct{})
	}

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: properties,
			}),
		},
	}

	return tool.NewCommandTool(
		v1.Tool_builder{InputSchema: inputSchema}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)
}

func newJSONCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = configv1.CommandLineCallDefinition_builder{}.Build()
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command:               proto.String(command),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
	}).Build()
	return tool.NewCommandTool(
		v1.Tool_builder{}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)
}

func TestCommandTool_Execute(t *testing.T) {
	t.Parallel()

	t.Run("successful execution", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("/usr/bin/env", callDef)
		inputData := map[string]interface{}{"args": []string{"echo", "hello world"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "hello world\n", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "hello world\n", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])
		assert.Equal(t, 0, resultMap["return_code"])
	})

	t.Run("command not found", func(t *testing.T) {
		t.Parallel()
		cmdTool := newCommandTool("this-command-does-not-exist", nil)
		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("execution with environment variables", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("MY_VAR"),
					}.Build(),
				}.Build(),
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("args"),
					}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("/usr/bin/env", callDef)
		inputData := map[string]interface{}{
			"args":   []string{"printenv", "MY_VAR"},
			"MY_VAR": "hello from env",
		}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "hello from env\n", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "hello from env\n", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])
		assert.Equal(t, 0, resultMap["return_code"])
	})

	t.Run("non-zero exit code", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("/usr/bin/env", callDef)
		inputData := map[string]interface{}{"args": []string{"false"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "/usr/bin/env", resultMap["command"])
		assert.Equal(t, "", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusError, resultMap["status"])
		assert.Equal(t, 1, resultMap["return_code"])
	})

	t.Run("malformed tool inputs", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)
		inputs := json.RawMessage(`{"args": "not-an-array"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("json communication protocol", func(t *testing.T) {
		t.Parallel()
		// Build the jsonecho binary
		wd, err := os.Getwd()
		require.NoError(t, err)
		jsonechoDir := filepath.Join(wd, "testdata", "jsonecho")
		jsonechoBin := filepath.Join(jsonechoDir, "jsonecho")

		// Ensure the directory exists
		err = os.MkdirAll(jsonechoDir, 0755)
		require.NoError(t, err)

		// Build it
		cmd := exec.Command("go", "build", "-o", jsonechoBin, "main.go")
		cmd.Dir = jsonechoDir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to build jsonecho: %s", string(out))

		cmdTool := newJSONCommandTool(jsonechoBin, nil)
		inputData := map[string]interface{}{"foo": "bar"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "bar", resultMap["foo"])
	})

	t.Run("argument substitution", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{text}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("text")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)
		inputData := map[string]interface{}{"text": "hello"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "echo", resultMap["command"])
		assert.Equal(t, "hello\n", resultMap["stdout"])
	})
}

func TestCommandTool_GetCacheConfig(t *testing.T) {
	t.Parallel()
	cacheConfig := configv1.CacheConfig_builder{}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()
	callDef.SetCache(cacheConfig)
	cmdTool := newCommandTool("echo", callDef)
	assert.Equal(t, cacheConfig, cmdTool.GetCacheConfig())
}

func TestCommandTool_Tool(t *testing.T) {
	t.Parallel()
	toolProto := v1.Tool_builder{
		Name: proto.String("test-tool"),
	}.Build()
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}).Build()
	cmdTool := tool.NewCommandTool(toolProto, service, configv1.CommandLineCallDefinition_builder{}.Build(), nil, "call-id")
	assert.Equal(t, toolProto, cmdTool.Tool())
}
