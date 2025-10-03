/*
 * Copyright 2025 Author(s) of MCPX
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

	"github.com/mcpxy/mcpx/pkg/tool"
	v1 "github.com/mcpxy/mcpx/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandTool_Execute(t *testing.T) {
	t.Run("successful execution with plain text output", func(t *testing.T) {
		toolProto := &v1.Tool{}
		cmdTool := tool.NewCommandTool(toolProto, "echo")

		inputData := map[string]interface{}{"args": []string{"hello world"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, "hello world\n", result)
	})

	t.Run("successful execution with json output", func(t *testing.T) {
		toolProto := &v1.Tool{}
		cmdTool := tool.NewCommandTool(toolProto, "echo")

		jsonString := `{"key": "value"}`
		inputData := map[string]interface{}{"args": []string{jsonString}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)

		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		expected := map[string]any{"key": "value"}
		assert.Equal(t, expected, result)
	})

	t.Run("command not found", func(t *testing.T) {
		toolProto := &v1.Tool{}
		cmdTool := tool.NewCommandTool(toolProto, "this-command-does-not-exist")

		req := &tool.ExecutionRequest{}
		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("execution with environment variables", func(t *testing.T) {
		toolProto := &v1.Tool{}
		cmdTool := tool.NewCommandTool(toolProto, "sh")

		// Use `sh -c` to execute a command that prints an environment variable
		inputData := map[string]interface{}{
			"args":   []string{"-c", "echo $MY_VAR"},
			"MY_VAR": "hello from env",
		}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "hello from env\n", result)
	})

	t.Run("malformed tool inputs", func(t *testing.T) {
		toolProto := &v1.Tool{}
		cmdTool := tool.NewCommandTool(toolProto, "echo")

		inputs := json.RawMessage(`{"args": "not-an-array"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}
