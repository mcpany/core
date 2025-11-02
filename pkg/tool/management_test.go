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
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func (s *MockMCPToolServer) HasTool(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.tools[name]
	return ok
}

func (s *MockMCPToolServer) GetToolHandler(name string) (mcp.ToolHandler, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	handler, ok := s.tools[name]
	return handler, ok
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
