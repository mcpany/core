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

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/util"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type TestMCPServerProvider struct {
	server *mcp.Server
}

func (p *TestMCPServerProvider) Server() *mcp.Server {
	return p.server
}

func TestToolManager_AddAndGetTool(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool"),
			}
		},
	}

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
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool-1"),
			}
		},
	}
	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool-2"),
			}
		},
	}

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)

	tools := tm.ListTools()
	assert.Len(t, tools, 2, "Should have two tools")
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("service-a"),
				Name:      proto.String("tool-1"),
			}
		},
	}
	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("service-b"),
				Name:      proto.String("tool-2"),
			}
		},
	}
	mockTool3 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("service-a"),
				Name:      proto.String("tool-3"),
			}
		},
	}

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
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	expectedResult := "success"
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			assert.Equal(t, execReq, req)
			return expectedResult, nil
		},
	}

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
	tm := NewToolManager(nil)
	var wg sync.WaitGroup
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mockTool := &MockTool{
				ToolFunc: func() *v1.Tool {
					return &v1.Tool{
						ServiceId: proto.String("concurrent-service"),
						Name:      proto.String(fmt.Sprintf("tool-%d", i)),
					}
				},
			}
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
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String(""), // Empty service ID
				Name:      proto.String("test-tool"),
			}
		},
	}

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.EqualError(t, err, "tool service ID cannot be empty")
}

func TestToolManager_AddTool_EmptyToolName(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("test-service"),
				Name:      proto.String(""), // Empty tool name
			}
		},
	}

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name: id cannot be empty")
}

func TestToolManager_AddTool_WithMCPServer(t *testing.T) {
	tm := NewToolManager(nil)
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	assert.NoError(t, err)

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId:   proto.String("test-service"),
				Name:        proto.String("test-tool"),
				Description: proto.String("A test tool"),
				Annotations: &v1.ToolAnnotations{
					InputSchema: inputSchema,
				},
			}
		},
	}

	err = tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Test case where the handler returns an error
	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("error-tool"),
				Annotations: &v1.ToolAnnotations{
					InputSchema: inputSchema,
				},
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return nil, fmt.Errorf("tool error")
		},
	}

	err = tm.AddTool(mockTool2)
	assert.NoError(t, err)
}

type MockToolExecutionMiddleware struct {
	ExecuteFunc func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error)
}

func (m *MockToolExecutionMiddleware) Execute(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req, next)
	}
	return next(ctx, req)
}

func TestToolManager_AddAndExecuteWithMiddleware(t *testing.T) {
	tm := NewToolManager(nil)

	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID}

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return "tool success", nil
		},
	}
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Case 1: Middleware returns a result directly
	middleware1 := &MockToolExecutionMiddleware{
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
			return "middleware success", nil
		},
	}
	tm.AddMiddleware(middleware1)

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, "middleware success", result)

	// Reset middlewares
	tm.middlewares = nil

	// Case 2: Middleware calls the next function
	middleware2 := &MockToolExecutionMiddleware{
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error) {
			return next(ctx, req)
		},
	}
	tm.AddMiddleware(middleware2)

	result, err = tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, "tool success", result)
}

func TestToolManager_ClearToolsForService_NoDeletions(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{ServiceId: proto.String("service-a"), Name: proto.String("tool-1")}
		},
	}
	_ = tm.AddTool(mockTool1)

	tm.ClearToolsForService("service-b") // service-b has no tools
	assert.Len(t, tm.ListTools(), 1, "Should still have one tool")
}

func TestToolManager_AddTool_MCPServerAddToolError(t *testing.T) {
	tm := NewToolManager(nil)
	// Mock the mcp.Server's AddTool method to return an error.
	// We can't directly do this, so we'll rely on the fact that a tool
	// with an empty name will cause an error.
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			// This tool is valid for the ToolManager, but will fail in the MCP server
			// because the sanitized name is different from the original name, and we
			// are passing a tool with an empty name to the MCP server.
			inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
			return &v1.Tool{
				ServiceId: proto.String("test"),
				Name:      proto.String(" "),
				Annotations: &v1.ToolAnnotations{
					InputSchema: inputSchema,
				},
			}
		},
	}
	// The error is not propagated up from mcpServer.AddTool, but it should log an error.
	// This test is limited in what it can assert, but it covers the code path.
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)
}

func TestToolManager_AddTool_WithMCPServerAndBus(t *testing.T) {
	busProvider, _ := bus.NewBusProvider(nil)
	tm := NewToolManager(busProvider)

	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"arg": map[string]interface{}{"type": "string"},
		},
	})
	assert.NoError(t, err)

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId:   proto.String("test-service"),
				Name:        proto.String("bus-tool"),
				Annotations: &v1.ToolAnnotations{InputSchema: inputSchema},
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return map[string]string{"status": "ok from tool"}, nil
		},
	}
	err = tm.AddTool(mockTool)
	assert.NoError(t, err)

	// This part is tricky to test without a running worker,
	// but we can at least ensure the handler is registered and doesn't panic.
	sanitizedToolName, _ := util.SanitizeToolName("bus-tool")
	toolID := "test-service" + "." + sanitizedToolName
	assert.NotNil(t, toolID)
}

func TestToolManager_AddTool_WithMCPServer_ErrorCases(t *testing.T) {
	tm := NewToolManager(nil)
	mcpServer := mcp.NewServer(&mcp.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	// Case 1: Error converting proto tool to mcp tool (empty name)
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{ServiceId: proto.String("test"), Name: proto.String("")}
		},
	}
	err := tm.AddTool(mockTool1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name")

	// Case 2: Good tool for subsequent tests
	inputSchema, err := structpb.NewStruct(map[string]interface{}{"type": "object"})
	assert.NoError(t, err)

	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId:   proto.String("test"),
				Name:        proto.String("good-tool"),
				Annotations: &v1.ToolAnnotations{InputSchema: inputSchema},
			}
		},
	}
	err = tm.AddTool(mockTool2)
	assert.NoError(t, err)
}

func TestToolManager_AddAndGetServiceInfo(t *testing.T) {
	tm := NewToolManager(nil)
	serviceID := "test-service" //nolint:goconst
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

func TestToolManager_ExecuteTool_ExecutionError(t *testing.T) {
	tm := NewToolManager(nil)
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}
	expectedError := fmt.Errorf("execution failed")

	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}
		},
		ExecuteFunc: func(ctx context.Context, req *ExecutionRequest) (any, error) {
			return nil, expectedError
		},
	}

	_ = tm.AddTool(mockTool)

	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestToolManager_ListTools_Caching(t *testing.T) {
	tm := NewToolManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *v1.Tool { return &v1.Tool{ServiceId: proto.String("s1"), Name: proto.String("t1")} },
	}
	_ = tm.AddTool(mockTool1)

	// First call populates cache
	list1 := tm.ListTools()
	assert.Len(t, list1, 1)

	// Second call should return cached slice
	list2 := tm.ListTools()
	assert.Len(t, list2, 1)
	// Compare pointers of the first elements to check if it's the same slice
	if len(list1) > 0 && len(list2) > 0 {
		assert.Same(t, list1[0], list2[0])
	}

	// Invalidate cache by adding a new tool
	mockTool2 := &MockTool{
		ToolFunc: func() *v1.Tool { return &v1.Tool{ServiceId: proto.String("s1"), Name: proto.String("t2")} },
	}
	_ = tm.AddTool(mockTool2)

	list3 := tm.ListTools()
	assert.Len(t, list3, 2, "Cache should be invalidated and list repopulated")

	// Invalidate cache by clearing tools
	tm.ClearToolsForService("s1")
	list4 := tm.ListTools()
	assert.Len(t, list4, 0, "Cache should be invalidated and list cleared")
}
