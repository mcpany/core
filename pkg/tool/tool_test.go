
// Copyright 2024 Author of MCP any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCallable struct {
	CallFunc func(ctx context.Context, req *ExecutionRequest) (any, error)
}

func (m *mockCallable) Call(ctx context.Context, req *ExecutionRequest) (any, error) {
	return m.CallFunc(ctx, req)
}

func TestCallableTool_Execute(t *testing.T) {
	t.Parallel()

	mock := &mockCallable{
		CallFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			assert.Equal(t, "test-tool", req.ToolName)
			return "test-result", nil
		},
	}

	toolName := "test-tool"
	tool, err := NewCallableTool(&configv1.ToolDefinition{
		Name: &toolName,
	}, nil, mock)
	require.NoError(t, err)

	result, err := tool.Execute(context.Background(), &ExecutionRequest{
		ToolName: "test-tool",
	})
	require.NoError(t, err)
	assert.Equal(t, "test-result", result)
}

func TestCallableTool_Callable(t *testing.T) {
	t.Parallel()

	mock := &mockCallable{}
	toolName := "test-tool"
	tool, err := NewCallableTool(&configv1.ToolDefinition{
		Name: &toolName,
	}, nil, mock)
	require.NoError(t, err)

	assert.Equal(t, mock, tool.Callable())
}

func TestBaseTool_Tool(t *testing.T) {
	t.Parallel()

	toolName := "test-tool"
	toolDef := &configv1.ToolDefinition{
		Name: &toolName,
	}
	tool, err := newBaseTool(toolDef, nil, nil)
	require.NoError(t, err)

	pbTool := tool.Tool()
	assert.Equal(t, "test-tool", pbTool.GetName())
}

func TestExecutionRequest_Unmarshal(t *testing.T) {
	t.Parallel()

	jsonBytes := []byte(`{"ToolName": "test-tool", "ToolInputs": {"arg1": "value1"}}`)
	var req ExecutionRequest
	err := json.Unmarshal(jsonBytes, &req)
	require.NoError(t, err)

	assert.Equal(t, "test-tool", req.ToolName)
	assert.Equal(t, json.RawMessage(`{"arg1": "value1"}`), req.ToolInputs)
}

// createTempScript creates a temporary executable script for testing.
func createTempScript(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp("", "test-script-*.sh")
	require.NoError(t, err)
	_, err = file.WriteString(content)
	require.NoError(t, err)
	err = file.Chmod(0755)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)
	return file.Name()
}
