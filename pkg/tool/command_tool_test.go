/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/consts"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func newCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = &configv1.CommandLineCallDefinition{}
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String(command),
	}).Build()
	return tool.NewCommandTool(
		&v1.Tool{},
		service,
		callDef,
	)
}

func TestCommandTool_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		cmdTool := newCommandTool("echo", nil)
		inputData := map[string]interface{}{"args": []string{"hello world"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "echo", resultMap["command"])
		assert.Equal(t, "hello world\n", resultMap["stdout"])
		assert.Equal(t, "", resultMap["stderr"])
		assert.Equal(t, "hello world\n", resultMap["combined_output"])
		assert.NotNil(t, resultMap["start_time"])
		assert.NotNil(t, resultMap["end_time"])
		assert.Equal(t, consts.CommandStatusSuccess, resultMap["status"])
		assert.Equal(t, 0, resultMap["return_code"])
	})

	t.Run("command not found", func(t *testing.T) {
		cmdTool := newCommandTool("this-command-does-not-exist", nil)
		req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("execution with environment variables", func(t *testing.T) {
		cmdTool := newCommandTool("/usr/bin/env", nil)
		inputData := map[string]interface{}{
			"args":   []string{"bash", "-c", "echo $MY_VAR"},
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
		cmdTool := newCommandTool("/usr/bin/env", nil)
		inputData := map[string]interface{}{"args": []string{"bash", "-c", "exit 1"}}
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
		cmdTool := newCommandTool("echo", nil)
		inputs := json.RawMessage(`{"args": "not-an-array"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestCommandTool_GetCacheConfig(t *testing.T) {
	cacheConfig := &configv1.CacheConfig{}
	callDef := &configv1.CommandLineCallDefinition{}
	callDef.SetCache(cacheConfig)
	cmdTool := newCommandTool("echo", callDef)
	assert.Equal(t, cacheConfig, cmdTool.GetCacheConfig())
}
