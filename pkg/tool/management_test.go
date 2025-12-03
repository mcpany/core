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
	"reflect"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/util"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/structpb"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type TestMCPServerProvider struct {
	server *mcp.Server
}

func (p *TestMCPServerProvider) Server() *mcp.Server {
	return p.server
}

func TestToolManager_AddAndGetTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool")
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()
	mockTool.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-tool")
	toolID := "test-service" + "." + sanitizedToolName
	retrievedTool, ok := tm.GetTool(toolID)
	assert.True(t, ok, "Tool should be found")
	assert.Equal(t, mockTool, retrievedTool, "Retrieved tool should be the one that was added")
}

func TestToolManager_ListTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool1 := NewMockTool(ctrl)
	toolProto1 := &v1.Tool{}
	toolProto1.SetServiceId("test-service")
	toolProto1.SetName("test-tool-1")
	mockTool1.EXPECT().Tool().Return(toolProto1).AnyTimes()
	mockTool1.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	mockTool2 := NewMockTool(ctrl)
	toolProto2 := &v1.Tool{}
	toolProto2.SetServiceId("test-service")
	toolProto2.SetName("test-tool-2")
	mockTool2.EXPECT().Tool().Return(toolProto2).AnyTimes()
	mockTool2.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)

	tools := tm.ListTools()
	assert.Len(t, tools, 2, "Should have two tools")
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool1 := NewMockTool(ctrl)
	toolProto1 := &v1.Tool{}
	toolProto1.SetServiceId("service-a")
	toolProto1.SetName("tool-1")
	mockTool1.EXPECT().Tool().Return(toolProto1).AnyTimes()
	mockTool1.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	mockTool2 := NewMockTool(ctrl)
	toolProto2 := &v1.Tool{}
	toolProto2.SetServiceId("service-b")
	toolProto2.SetName("tool-2")
	mockTool2.EXPECT().Tool().Return(toolProto2).AnyTimes()
	mockTool2.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	mockTool3 := NewMockTool(ctrl)
	toolProto3 := &v1.Tool{}
	toolProto3.SetServiceId("service-a")
	toolProto3.SetName("tool-3")
	mockTool3.EXPECT().Tool().Return(toolProto3).AnyTimes()
	mockTool3.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	expectedResult := "success"
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}

	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()
	mockTool.EXPECT().Execute(gomock.Any(), execReq).Return(expectedResult, nil)

	_ = tm.AddTool(mockTool)

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestToolManager_ExecuteTool_NotFound(t *testing.T) {
	tm := NewToolManager(nil)
	execReq := &ExecutionRequest{ToolName: "non-existent-tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err, "Should return an error for a non-existent tool")
	assert.Equal(t, ErrToolNotFound, err, "Error should be ErrToolNotFound")
}

func TestToolManager_ConcurrentAccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	var wg sync.WaitGroup
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mockTool := NewMockTool(ctrl)
			toolProto := &v1.Tool{}
			toolProto.SetServiceId("concurrent-service")
			toolProto.SetName(fmt.Sprintf("tool-%d", i))
			mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()
			mockTool.EXPECT().GetCacheConfig().Return(nil).AnyTimes()
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("") // Empty service ID
	toolProto.SetName("test-tool")
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.EqualError(t, err, "tool service ID cannot be empty")
}

func TestToolManager_AddTool_EmptyToolName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("") // Empty tool name
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name: id cannot be empty")
}

func TestToolManager_AddTool_WithMCPServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	mockTool := NewMockTool(ctrl)
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

	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	// If this doesn't panic, it means the tool was added successfully to the mcpServer.
	err = tm.AddTool(mockTool)
	assert.NoError(t, err)
}

func TestToolManager_ExecuteTool_Middleware_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockMiddleware := NewMockToolExecutionMiddleware(ctrl)
	tm.AddMiddleware(mockMiddleware)

	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID}

	expectedErr := fmt.Errorf("middleware error")
	mockMiddleware.EXPECT().Execute(gomock.Any(), execReq, gomock.Any()).Return(nil, expectedErr)

	_, err = tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestToolManager_ClearToolsForService_Cache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockTool1 := NewMockTool(ctrl)
	toolProto1 := &v1.Tool{}
	toolProto1.SetServiceId("service-a")
	toolProto1.SetName("tool-1")
	mockTool1.EXPECT().Tool().Return(toolProto1).AnyTimes()
	mockTool1.EXPECT().GetCacheConfig().Return(nil).AnyTimes()

	_ = tm.AddTool(mockTool1)

	// This will cache the tool list
	tm.ListTools()

	tm.ClearToolsForService("service-a")
	tools := tm.ListTools()
	assert.Len(t, tools, 0, "Should have no tools remaining")
}

// MockToolExecutionMiddleware is a mock implementation of the ToolExecutionMiddleware interface.
type MockToolExecutionMiddleware struct {
	ctrl     *gomock.Controller
	recorder *MockToolExecutionMiddlewareMockRecorder
}

// MockToolExecutionMiddlewareMockRecorder is the mock recorder for MockToolExecutionMiddleware.
type MockToolExecutionMiddlewareMockRecorder struct {
	mock *MockToolExecutionMiddleware
}

// NewMockToolExecutionMiddleware creates a new mock instance.
func NewMockToolExecutionMiddleware(ctrl *gomock.Controller) *MockToolExecutionMiddleware {
	mock := &MockToolExecutionMiddleware{ctrl: ctrl}
	mock.recorder = &MockToolExecutionMiddlewareMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockToolExecutionMiddleware) EXPECT() *MockToolExecutionMiddlewareMockRecorder {
	return m.recorder
}

func (m *MockToolExecutionMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
	ret := m.ctrl.Call(m, "Execute", ctx, req, next)
	ret0, _ := ret[0].(any)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockToolExecutionMiddlewareMockRecorder) Execute(ctx, req, next interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Execute", reflect.TypeOf((*MockToolExecutionMiddleware)(nil).Execute), ctx, req, next)
}

func TestToolManager_AddAndExecuteWithMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mockMiddleware := NewMockToolExecutionMiddleware(ctrl)
	tm.AddMiddleware(mockMiddleware)

	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("exec-service")
	toolProto.SetName("exec-tool")
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID}

	// Case 1: Middleware returns a result directly
	expectedResult := "middleware success"
	mockMiddleware.EXPECT().Execute(gomock.Any(), execReq, gomock.Any()).Return(expectedResult, nil)

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Case 2: Middleware calls the next function
	expectedToolResult := "tool success"
	mockMiddleware.EXPECT().Execute(gomock.Any(), execReq, gomock.Any()).DoAndReturn(
		func(ctx context.Context, req *ExecutionRequest, next ToolExecutionFunc) (any, error) {
			return next(ctx, req)
		})
	mockTool.EXPECT().Execute(gomock.Any(), execReq).Return(expectedToolResult, nil)

	result, err = tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedToolResult, result)
}

func TestGetFullyQualifiedToolName(t *testing.T) {
	serviceID := "test-service"
	toolName := "test-tool"
	expectedFQN := "test-service.test-tool"
	fqn := GetFullyQualifiedToolName(serviceID, toolName)
	assert.Equal(t, expectedFQN, fqn)
}

func TestToolManager_AddAndGetServiceInfo(t *testing.T) {
	tm := NewToolManager(nil)
	serviceID := "test-service"
	serviceInfo := &ServiceInfo{
		Name: "Test Service",
	}

	tm.AddServiceInfo(serviceID, serviceInfo)

	retrievedInfo, ok := tm.GetServiceInfo(serviceID)
	assert.True(t, ok, "Service info should be found")
	assert.Equal(t, serviceInfo, retrievedInfo, "Retrieved service info should match the added info")

	_, ok = tm.GetServiceInfo("non-existent-service")
	assert.False(t, ok, "Service info for a non-existent service should not be found")
}

func TestToolManager_SetMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	provider := &TestMCPServerProvider{server: nil}
	tm.SetMCPServer(provider)
	assert.Equal(t, provider, tm.mcpServer, "MCPServerProvider should be set")
}

func TestToolManager_AddTool_Handler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := NewToolManager(nil)
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	mockTool := NewMockTool(ctrl)
	toolProto := &v1.Tool{}
	toolProto.SetServiceId("test-service")
	toolProto.SetName("test-tool")
	// Add an input schema to test the conversion logic and prevent a panic in the MCP server
	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	assert.NoError(t, err)
	annotations := &v1.ToolAnnotations{}
	annotations.SetInputSchema(inputSchema)
	toolProto.SetAnnotations(annotations)
	mockTool.EXPECT().Tool().Return(toolProto).AnyTimes()

	err = tm.AddTool(mockTool)
	assert.NoError(t, err)
}
