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

package tool

import (
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMCPClient is a mock for the MCPClient interface
type MockMCPClient struct {
	mock.Mock
}

func (m *MockMCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mcp.CallToolResult), args.Error(1)
}

func TestToolManager_AddTool_WithCacheConfig(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool-cache")
	mockTool.On("Tool").Return(toolProto)
	mockTool.On("GetCacheConfig").Return(&configv1.CacheConfig{})

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-tool-cache")
	toolID := "test-service" + "." + sanitizedToolName
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok)
	assert.Equal(t, mockTool, retrievedTool)
}

func TestToolManager_AddTool_Duplicate(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("duplicate-tool")
	mockTool.On("Tool").Return(toolProto)
	mockTool.On("GetCacheConfig").Return(nil)

	err1 := tm.AddTool(mockTool)
	assert.NoError(t, err1)

	err2 := tm.AddTool(mockTool)
	assert.NoError(t, err2, "Adding a duplicate tool should not return an error, it should overwrite")
}

func TestToolManager_AddTool_Sanitization(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("tool with spaces")
	mockTool.On("Tool").Return(toolProto)
	mockTool.On("GetCacheConfig").Return(nil)

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("tool with spaces")
	toolID := "test-service" + "." + sanitizedToolName
	_, ok := tm.GetTool(toolID)
	assert.True(t, ok, "Tool should be found with sanitized name")
}

func TestToolManager_AddTool_EmptyID(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("")
	toolProto.SetName("")
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.Error(t, err, "Should return an error for tool with empty ID")
}

func TestGetFullyQualifiedToolName(t *testing.T) {
	serviceID := "test-service"
	toolName := "test-tool"
	expected := "test-service.test-tool"
	actual := GetFullyQualifiedToolName(serviceID, toolName)
	assert.Equal(t, expected, actual)
}

func TestToolManager_ToolMethods(t *testing.T) {
	tm := NewToolManager(nil)

	// Test HTTPTool
	httpToolProto := &v1.Tool{}
	httpToolProto.SetServiceId("http-service")
	httpToolProto.SetName("http-tool")
	httpTool := NewHTTPTool(httpToolProto, nil, "http-service", nil, &configv1.HttpCallDefinition{})
	err := tm.AddTool(httpTool)
	assert.NoError(t, err)
	assert.NotNil(t, httpTool.Tool())
	assert.Nil(t, httpTool.GetCacheConfig())

	// Test CommandTool
	commandToolProto := &v1.Tool{}
	commandToolProto.SetServiceId("command-service")
	commandToolProto.SetName("command-tool")
	commandTool := NewCommandTool(commandToolProto, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{})
	err = tm.AddTool(commandTool)
	assert.NoError(t, err)
	assert.NotNil(t, commandTool.Tool())
	assert.Nil(t, commandTool.GetCacheConfig())

	// Test MCPTool
	mcpToolProto := &v1.Tool{}
	mcpToolProto.SetServiceId("mcp-service")
	mcpToolProto.SetName("mcp-tool")
	mockMCPClient := new(MockMCPClient)
	mcpTool := NewMCPTool(mcpToolProto, mockMCPClient, &configv1.MCPCallDefinition{})
	err = tm.AddTool(mcpTool)
	assert.NoError(t, err)
	assert.NotNil(t, mcpTool.Tool())
	assert.Nil(t, mcpTool.GetCacheConfig())
}

// MiddlewareFunc is a function that implements the ToolExecutionMiddleware interface.
type MiddlewareFunc func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error)

func (f MiddlewareFunc) Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
	return f(ctx, req, next)
}

func TestToolManager_ExecuteTool_WithMiddleware(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	expectedResult := "success"
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}

	mockTool.On("Tool").Return(toolProto)
	mockTool.On("Execute", mock.Anything, execReq).Return(expectedResult, nil)

	_ = tm.AddTool(mockTool)

	// Middleware that adds a value to the context
	middleware := MiddlewareFunc(func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
		ctx = context.WithValue(ctx, "middleware-key", "middleware-value")
		return next(ctx, req)
	})
	tm.AddMiddleware(middleware)

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockTool.AssertExpectations(t)
}
