/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/mcpxy/core/pkg/util"
	pb "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
)

// MockTool is a mock implementation of the Tool interface for testing purposes.
type MockTool struct {
	mock.Mock
	tool *pb.Tool
}

func (m *MockTool) Tool() *pb.Tool {
	args := m.Called()
	return args.Get(0).(*pb.Tool)
}

func (m *MockTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
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

func (m *MockMCPServerProvider) RemoveTool(toolName string) {
	m.Called(toolName)
}

func TestToolManager_AddAndGetTool(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := new(MockTool)
	toolProto := &pb.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool")
	mockTool.On("Tool").Return(toolProto)

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	toolID, _ := util.GenerateToolID("test-service", "test-tool")
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok, "Tool should be found")
	assert.Equal(t, mockTool, retrievedTool, "Retrieved tool should be the one that was added")
}

func TestToolManager_ListTools(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := new(MockTool)
	toolProto1 := &pb.Tool{}
	toolProto1.SetServiceId("test-service")
	toolProto1.SetName("test-tool-1")
	mockTool1.On("Tool").Return(toolProto1)

	mockTool2 := new(MockTool)
	toolProto2 := &pb.Tool{}
	toolProto2.SetServiceId("test-service")
	toolProto2.SetName("test-tool-2")
	mockTool2.On("Tool").Return(toolProto2)

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)

	tools := tm.ListTools()
	assert.Len(t, tools, 2, "Should have two tools")
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := new(MockTool)
	toolProto1 := &pb.Tool{}
	toolProto1.SetServiceId("service-a")
	toolProto1.SetName("tool-1")
	mockTool1.On("Tool").Return(toolProto1)

	mockTool2 := new(MockTool)
	toolProto2 := &pb.Tool{}
	toolProto2.SetServiceId("service-b")
	toolProto2.SetName("tool-2")
	mockTool2.On("Tool").Return(toolProto2)

	mockTool3 := new(MockTool)
	toolProto3 := &pb.Tool{}
	toolProto3.SetServiceId("service-a")
	toolProto3.SetName("tool-3")
	mockTool3.On("Tool").Return(toolProto3)

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
	toolProto := &pb.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	toolID, _ := util.GenerateToolID("exec-service", "exec-tool")
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

func TestToolManager_AddServiceInfo(t *testing.T) {
	tm := NewToolManager(nil)
	serviceInfo := &ServiceInfo{Name: "Test Service"}
	tm.AddServiceInfo("service1", serviceInfo)

	retrievedInfo, ok := tm.GetServiceInfo("service1")
	assert.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	_, ok = tm.GetServiceInfo("non-existent-service")
	assert.False(t, ok)
}

type MockMCPToolServer struct {
	*mcp.Server
	mu    sync.Mutex
	tools map[string]mcp.ToolHandler
}

func NewMockMCPToolServer() *MockMCPToolServer {
	return &MockMCPToolServer{
		Server: mcp.NewServer(&mcp.Implementation{}, nil),
		tools:  make(map[string]mcp.ToolHandler),
	}
}

func (s *MockMCPToolServer) AddTool(tool *mcp.Tool, handler mcp.ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = handler
}

func (s *MockMCPToolServer) RemoveTool(toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tools, toolName)
}

func (s *MockMCPToolServer) GetTool(toolName string) (mcp.ToolHandler, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	handler, ok := s.tools[toolName]
	return handler, ok
}

func TestToolManager_ClearToolsForService_UnregistersFromMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	mockMCPServer := NewMockMCPToolServer()
	mockProvider := new(MockMCPServerProvider)
	mockProvider.On("Server").Return(mockMCPServer.Server)
	tm.SetMCPServer(mockProvider)

	mockTool := new(MockTool)
	toolProto := &pb.Tool{}
	toolProto.SetServiceId("service-to-clear")
	toolProto.SetName("tool-to-clear")
	// Create a valid input schema using structpb
	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"test": map[string]interface{}{
				"type": "string",
			},
		},
	})
	assert.NoError(t, err)
	annotations := &pb.ToolAnnotations{}
	annotations.SetInputSchema(inputSchema)
	toolProto.SetAnnotations(annotations)
	mockTool.On("Tool").Return(toolProto)

	err = tm.AddTool(mockTool)
	assert.NoError(t, err)

	toolID, _ := util.GenerateToolID("service-to-clear", "tool-to-clear")

	// Expect the RemoveTool method to be called with the correct tool ID
	mockProvider.On("RemoveTool", toolID).Run(func(args mock.Arguments) {
		mockMCPServer.RemoveTool(args.String(0))
	}).Once()

	tm.ClearToolsForService("service-to-clear")

	_, ok := mockMCPServer.GetTool(toolID)
	assert.False(t, ok, "Tool should be unregistered from the MCP server after clearing")
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
			toolProto := &pb.Tool{}
			toolProto.SetServiceId("concurrent-service")
			toolProto.SetName(fmt.Sprintf("tool-%d", i))
			mockTool.On("Tool").Return(toolProto)
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
			toolID, _ := util.GenerateToolID("concurrent-service", fmt.Sprintf("tool-%d", i))
			_, ok := tm.GetTool(toolID)
			assert.True(t, ok)
		}(i)
	}
	wg.Wait()
}
