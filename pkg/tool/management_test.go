/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may a copy of the License at
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
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTool is a mock implementation of the Tool interface for testing purposes.
type MockTool struct {
	mock.Mock
	tool *v1.Tool
}

func (m *MockTool) Tool() *v1.Tool {
	args := m.Called()
	return args.Get(0).(*v1.Tool)
}

func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*configv1.CacheConfig)
}

func (m *MockTool) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockMCPServerProvider is a mock implementation of the MCPServerProvider interface.
type MockMCPServerProvider struct {
	mock.Mock
}

func (m *MockMCPServerProvider) Server() *mcp.Server {
	args := m.Called()
	return args.Get(0).(*mcp.Server)
}

func TestToolManager_AddAndGetTool(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool")
	mockTool.On("Tool").Return(toolProto)
	mockTool.On("GetCacheConfig").Return(nil)

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-tool")
	toolID := "test-service" + "." + sanitizedToolName
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok, "Tool should be found")
	assert.Equal(t, mockTool, retrievedTool, "Retrieved tool should be the one that was added")
}

func TestToolManager_ListTools(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := new(MockTool)
	toolProto1 := &v1.Tool{}
	toolProto1.SetServiceId("test-service")
	toolProto1.SetName("test-tool-1")
	mockTool1.On("Tool").Return(toolProto1)
	mockTool1.On("GetCacheConfig").Return(nil)

	mockTool2 := new(MockTool)
	toolProto2 := &v1.Tool{}
	toolProto2.SetServiceId("test-service")
	toolProto2.SetName("test-tool-2")
	mockTool2.On("Tool").Return(toolProto2)
	mockTool2.On("GetCacheConfig").Return(nil)

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)

	tools := tm.ListTools()
	assert.Len(t, tools, 2, "Should have two tools")
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := new(MockTool)
	toolProto1 := &v1.Tool{}
	toolProto1.SetServiceId("service-a")
	toolProto1.SetName("tool-1")
	mockTool1.On("Tool").Return(toolProto1)
	mockTool1.On("GetCacheConfig").Return(nil)

	mockTool2 := new(MockTool)
	toolProto2 := &v1.Tool{}
	toolProto2.SetServiceId("service-b")
	toolProto2.SetName("tool-2")
	mockTool2.On("Tool").Return(toolProto2)
	mockTool2.On("GetCacheConfig").Return(nil)

	mockTool3 := new(MockTool)
	toolProto3 := &v1.Tool{}
	toolProto3.SetServiceId("service-a")
	toolProto3.SetName("tool-3")
	mockTool3.On("Tool").Return(toolProto3)
	mockTool3.On("GetCacheConfig").Return(nil)

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)
	_ = tm.AddTool(mockTool3)

	assert.Len(t, tm.ListTools(), 3, "Should have three tools initially")

	tm.ClearToolsForService("service-a")
	tools := tm.ListTools()
	assert.Len(t, tools, 1, "Should have one tool remaining")
	assert.Equal(t, "service-b", tools[0].Tool().GetServiceId(), "The remaining tool should belong to service-b")
}

func TestToolManager_ExecuteTool(t *testing.T) {
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

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockTool.AssertExpectations(t)
}

func TestToolManager_ExecuteTool_NotFound(t *testing.T) {
	tm := NewToolManager(nil)
	execReq := &ExecutionRequest{ToolName: "non-existent-tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err, "Should return an error for a non-existent tool")
	assert.Equal(t, ErrToolNotFound, err, "Error should be ErrToolNotFound")
}

func TestToolManager_ConcurrentAccess(t *testing.T) {
	tm := NewToolManager(nil)
	var wg sync.WaitGroup
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mockTool := new(MockTool)
			toolProto := &v1.Tool{}
			toolProto.SetServiceId("concurrent-service")
			toolProto.SetName(fmt.Sprintf("tool-%d", i))
			mockTool.On("Tool").Return(toolProto)
			mockTool.On("GetCacheConfig").Return(nil)
			err := tm.AddTool(mockTool)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
	assert.Len(t, tm.ListTools(), numRoutines, "All tools should be added concurrently")

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sanitizedToolName, _ := util.SanitizeToolName(fmt.Sprintf("tool-%d", i))
			toolID := "concurrent-service" + "." + sanitizedToolName
			_, ok := tm.GetTool(toolID)
			assert.True(t, ok)
		}(i)
	}
	wg.Wait()
}

func TestToolManager_AddTool_NoServiceID(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("") // Empty service ID
	toolProto.SetName("test-tool")
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.EqualError(t, err, "tool service ID cannot be empty")
}

func TestToolManager_AddTool_EmptyToolName(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("") // Empty tool name
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name: id cannot be empty")
}

func TestToolManager_AddTool_WithMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	mockProvider := new(MockMCPServerProvider)
	mockProvider.On("Server").Return(mcpServer)
	tm.SetMCPServer(mockProvider)

	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool")
	toolProto.SetDescription("A test tool")
	annotations := &v1.ToolAnnotations{}
	// Add an input schema to test the conversion logic and prevent a panic in the MCP server
	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	assert.NoError(t, err)
	annotations.SetInputSchema(inputSchema)
	toolProto.SetAnnotations(annotations)

	mockTool.On("Tool").Return(toolProto)

	// If this doesn't panic, it means the tool was added successfully to the mcpServer.
	err = tm.AddTool(mockTool)
	assert.NoError(t, err)
}

// MockToolExecutionMiddleware is a mock implementation of the ToolExecutionMiddleware interface.
type MockToolExecutionMiddleware struct {
	mock.Mock
}

func (m *MockToolExecutionMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
	args := m.Called(ctx, req, next)
	// Allow the middleware to either call the next function or return directly
	if args.Get(0) == "call_next" {
		return next(ctx, req)
	}
	return args.Get(0), args.Error(1)
}

func TestToolManager_AddAndExecuteWithMiddleware(t *testing.T) {
	tm := NewToolManager(nil)
	mockMiddleware := new(MockToolExecutionMiddleware)
	tm.AddMiddleware(mockMiddleware)

	mockTool := new(MockTool)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID}

	// Case 1: Middleware returns a result directly
	expectedResult := "middleware success"
	mockMiddleware.On("Execute", mock.Anything, execReq, mock.Anything).Return(expectedResult, nil).Once()

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	mockMiddleware.AssertExpectations(t)

	// Case 2: Middleware calls the next function
	expectedToolResult := "tool success"
	mockMiddleware.On("Execute", mock.Anything, execReq, mock.Anything).Return("call_next", nil).Once()
	mockTool.On("Execute", mock.Anything, execReq).Return(expectedToolResult, nil).Once()

	result, err = tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedToolResult, result)
	mockMiddleware.AssertExpectations(t)
	mockTool.AssertExpectations(t)
}

func TestGetFullyQualifiedToolName(t *testing.T) {
	serviceID := "test-service"
	toolName := "test-tool"
	expectedFQN := "test-service.test-tool"
	fqn := GetFullyQualifiedToolName(serviceID, toolName)
	assert.Equal(t, expectedFQN, fqn)
}
