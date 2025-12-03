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
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestToolManager_AddTool_ErrorHandling(t *testing.T) {
	t.Parallel()

	tm := tool.NewToolManager(nil)

	t.Run("should return an error if tool service ID is empty", func(t *testing.T) {
		t.Parallel()
		mockTool := &tool.SimpleMockTool{}
		mockTool.ToolFunc = func() *v1.Tool {
			return &v1.Tool{}
		}
		err := tm.AddTool(mockTool)
		assert.Error(t, err)
	})

	t.Run("should return an error if tool name is invalid", func(t *testing.T) {
		t.Parallel()
		mockTool := &tool.SimpleMockTool{}
		mockTool.ToolFunc = func() *v1.Tool {
			return &v1.Tool{
				ServiceId: "test-service",
				Name:      "",
			}
		}
		err := tm.AddTool(mockTool)
		assert.Error(t, err)
	})
}

func TestToolManager_ExecuteTool_ErrorHandling(t *testing.T) {
	t.Parallel()

	tm := tool.NewToolManager(nil)
	mockTool := &tool.SimpleMockTool{}
	mockTool.ExecuteFunc = func(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
		return nil, errors.New("execution error")
	}
	mockTool.ToolFunc = func() *v1.Tool {
		return &v1.Tool{
			ServiceId: "test-service",
			Name:      "test-tool",
		}
	}
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	t.Run("should return an error if tool execution fails", func(t *testing.T) {
		t.Parallel()
		req := &tool.ExecutionRequest{
			ToolName: "test-service.test-tool",
		}
		_, err := tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("should return an error if tool is not found", func(t *testing.T) {
		t.Parallel()
		req := &tool.ExecutionRequest{
			ToolName: "non-existent-tool",
		}
		_, err := tm.ExecuteTool(context.Background(), req)
		assert.Error(t, err)
	})
}

func TestToolManager_Middleware(t *testing.T) {
	t.Parallel()

	tm := tool.NewToolManager(nil)
	mockTool := &tool.SimpleMockTool{}
	mockTool.ExecuteFunc = func(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
		return "original result", nil
	}
	mockTool.ToolFunc = func() *v1.Tool {
		return &v1.Tool{
			ServiceId: "test-service",
			Name:      "test-tool",
		}
	}
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	middleware := &tool.SimpleMockToolExecutionMiddleware{}
	middleware.ExecuteFunc = func(ctx context.Context, req *tool.ExecutionRequest, next tool.ToolExecutionFunc) (interface{}, error) {
		return "middleware result", nil
	}
	tm.AddMiddleware(middleware)

	req := &tool.ExecutionRequest{
		ToolName: "test-service.test-tool",
	}
	result, err := tm.ExecuteTool(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "middleware result", result)
}

func TestToolManager_GetServiceInfo(t *testing.T) {
	t.Parallel()

	tm := tool.NewToolManager(nil)
	serviceInfo := &tool.ServiceInfo{
		Name: "test-service",
	}
	tm.AddServiceInfo("test-service", serviceInfo)

	retrievedInfo, ok := tm.GetServiceInfo("test-service")
	assert.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	_, ok = tm.GetServiceInfo("non-existent-service")
	assert.False(t, ok)
}
