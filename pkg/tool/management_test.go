/*
 * Copyright 2025 Author(s) of MCPXY
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
	"fmt"
	"sync"
	"testing"

	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	v1 "github.com/mcpxy/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTool is a mock implementation of the Tool interface for testing.
type MockTool struct {
	tool    *v1.Tool
	execute func(ctx context.Context, req *tool.ExecutionRequest) (any, error)
}

func (m *MockTool) Tool() *v1.Tool {
	return m.tool
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if m.execute != nil {
		return m.execute(ctx, req)
	}
	return "mock result", nil
}

// mockMCPServerProvider is a mock that provides an mcp.Server.
type mockMCPServerProvider struct {
	server *mcp.Server
}

func newMockMCPServerProvider() *mockMCPServerProvider {
	impl := &mcp.Implementation{
		Name:    "mock-mcpx",
		Version: "v0.0.0",
	}
	return &mockMCPServerProvider{
		server: mcp.NewServer(impl, nil),
	}
}

func (m *mockMCPServerProvider) Server() *mcp.Server {
	return m.server
}

func TestToolManager_AddAndGetTool(t *testing.T) {
	tm := tool.NewToolManager()
	s, n := "test-service", "test-tool"
	mockTool := &MockTool{
		tool: v1.Tool_builder{
			Name:      &n,
			ServiceId: &s,
		}.Build(),
	}

	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	toolID, err := util.GenerateToolID("test-service", "test-tool")
	require.NoError(t, err)

	retrievedTool, ok := tm.GetTool(toolID)
	require.True(t, ok, "Tool should be found")
	assert.Equal(t, mockTool, retrievedTool)
}

func TestToolManager_ListTools(t *testing.T) {
	tm := tool.NewToolManager()
	s1, n1 := "service1", "tool1"
	s2, n2 := "service1", "tool2"
	s3, n3 := "service2", "tool1"

	tool1 := &MockTool{tool: v1.Tool_builder{Name: &n1, ServiceId: &s1}.Build()}
	tool2 := &MockTool{tool: v1.Tool_builder{Name: &n2, ServiceId: &s2}.Build()}
	tool3 := &MockTool{tool: v1.Tool_builder{Name: &n3, ServiceId: &s3}.Build()}

	tm.AddTool(tool1)
	tm.AddTool(tool2)
	tm.AddTool(tool3)

	tools := tm.ListTools()
	assert.Len(t, tools, 3)

	// Check if all tools are in the list
	var found1, found2, found3 bool
	for _, tl := range tools {
		if tl == tool1 {
			found1 = true
		}
		if tl == tool2 {
			found2 = true
		}
		if tl == tool3 {
			found3 = true
		}
	}
	assert.True(t, found1, "tool1 not found")
	assert.True(t, found2, "tool2 not found")
	assert.True(t, found3, "tool3 not found")
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	tm := tool.NewToolManager()
	s1, n1 := "service1", "tool1"
	s2, n2 := "service1", "tool2"
	s3, n3 := "service2", "tool1"

	tool1 := &MockTool{tool: v1.Tool_builder{Name: &n1, ServiceId: &s1}.Build()}
	tool2 := &MockTool{tool: v1.Tool_builder{Name: &n2, ServiceId: &s2}.Build()}
	tool3 := &MockTool{tool: v1.Tool_builder{Name: &n3, ServiceId: &s3}.Build()}

	tm.AddTool(tool1)
	tm.AddTool(tool2)
	tm.AddTool(tool3)

	tm.ClearToolsForService("service1")

	tools := tm.ListTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, tool3, tools[0])
}

func TestToolManager_AddAndGetServiceInfo(t *testing.T) {
	tm := tool.NewToolManager()
	serviceInfo := &tool.ServiceInfo{
		Name:   "test-service",
		Config: &configv1.UpstreamServiceConfig{},
	}

	tm.AddServiceInfo("test-service", serviceInfo)

	retrievedInfo, ok := tm.GetServiceInfo("test-service")
	require.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	_, ok = tm.GetServiceInfo("non-existent-service")
	assert.False(t, ok)
}

func TestToolManager_ExecuteTool_Success(t *testing.T) {
	tm := tool.NewToolManager()
	expectedResult := "execution successful"
	s, n := "exec-service", "exec-tool"
	mockTool := &MockTool{
		tool: v1.Tool_builder{Name: &n, ServiceId: &s}.Build(),
		execute: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return expectedResult, nil
		},
	}
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	toolID, _ := util.GenerateToolID("exec-service", "exec-tool")
	req := &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: json.RawMessage(`{}`),
	}

	result, err := tm.ExecuteTool(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestToolManager_ExecuteTool_NotFound(t *testing.T) {
	tm := tool.NewToolManager()
	req := &tool.ExecutionRequest{ToolName: "non-existent-tool"}
	_, err := tm.ExecuteTool(context.Background(), req)
	assert.ErrorIs(t, err, tool.ErrToolNotFound)
}

func TestToolManager_SetMCPServer(t *testing.T) {
	tm := tool.NewToolManager()
	mcpServer := newMockMCPServerProvider()
	tm.SetMCPServer(mcpServer)
}

func TestToolManager_AddTool_GenerateIDError(t *testing.T) {
	tm := tool.NewToolManager()
	// An empty tool name should cause an error in GenerateToolID
	s, n := "service1", ""
	mockTool := &MockTool{
		tool: v1.Tool_builder{Name: &n, ServiceId: &s}.Build(),
	}
	err := tm.AddTool(mockTool)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate tool ID")
}

func TestToolManager_Concurrency(t *testing.T) {
	tm := tool.NewToolManager()
	mcpServer := newMockMCPServerProvider()
	tm.SetMCPServer(mcpServer)

	var wg sync.WaitGroup
	numRoutines := 100

	// Test concurrent additions and clears
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			serviceID := fmt.Sprintf("service-%d", i%10)
			toolName := fmt.Sprintf("tool-%d", i)
			s, n := serviceID, toolName
			mockTool := &MockTool{
				tool: v1.Tool_builder{Name: &n, ServiceId: &s}.Build(),
			}
			err := tm.AddTool(mockTool)
			assert.NoError(t, err)

			if i%10 == 0 {
				tm.ClearToolsForService(serviceID)
			}
		}(i)
	}

	// Test concurrent reads
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tm.ListTools()
		}()
	}

	wg.Wait()
}

func TestNewToolManager(t *testing.T) {
	tm := tool.NewToolManager()
	assert.NotNil(t, tm, "NewToolManager should not return nil")
}

func TestToolManager_AddTool_NilMCPServer(t *testing.T) {
	tm := tool.NewToolManager()
	// Explicitly ensure mcpServer is nil
	tm.SetMCPServer(nil)

	s, n := "test-service", "test-tool"
	mockTool := &MockTool{
		tool: v1.Tool_builder{Name: &n, ServiceId: &s}.Build(),
	}

	// This should not panic
	err := tm.AddTool(mockTool)
	require.NoError(t, err)

	// Verify the tool was still added to the internal map
	toolID, _ := util.GenerateToolID("test-service", "test-tool")
	_, ok := tm.GetTool(toolID)
	assert.True(t, ok)
}
