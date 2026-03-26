// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"fmt"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// TestMCPServerProvider ...
// Summary: TestMCPServerProvider
	server *mcp_sdk.Server
}

// Server ...
// Summary: Server
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return p.server
}

// ptr is already defined in hooks_test.go in the same package.

// TestToolManager_AddAndGetTool ...
// Summary: TestToolManager_AddAndGetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool"),
			}.Build()
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

// TestToolManager_ListTools ...
// Summary: TestToolManager_ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool-1"),
			}.Build()
		},
	}
	mockTool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("test-tool-2"),
			}.Build()
		},
	}

	_ = tm.AddTool(mockTool1)
	_ = tm.AddTool(mockTool2)

	tools := tm.ListTools()
	assert.Len(t, tools, 2, "Should have two tools")
}

// TestToolManager_ClearToolsForService ...
// Summary: TestToolManager_ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("service-a"),
				Name:      proto.String("tool-1"),
			}.Build()
		},
	}
	mockTool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("service-b"),
				Name:      proto.String("tool-2"),
			}.Build()
		},
	}
	mockTool3 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("service-a"),
				Name:      proto.String("tool-3"),
			}.Build()
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

// TestToolManager_ExecuteTool ...
// Summary: TestToolManager_ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	expectedResult := "success"
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, req *ExecutionRequest) (any, error) {
			assert.Equal(t, execReq, req)
			return expectedResult, nil
		},
	}

	_ = tm.AddTool(mockTool)

	result, err := tm.ExecuteTool(context.Background(), execReq)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

// TestToolManager_ExecuteTool_NotFound ...
// Summary: TestToolManager_ExecuteTool_NotFound
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	execReq := &ExecutionRequest{ToolName: "non-existent-tool", ToolInputs: []byte(`{}`)}
	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err, "Should return an error for a non-existent tool")
	assert.Equal(t, ErrToolNotFound, err, "Error should be ErrToolNotFound")
}

// TestToolManager_ConcurrentAccess ...
// Summary: TestToolManager_ConcurrentAccess
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	var wg sync.WaitGroup
	numRoutines := 50

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mockTool := &MockTool{
				ToolFunc: func() *mcp_router_v1.Tool {
					return mcp_router_v1.Tool_builder{
						ServiceId: proto.String("concurrent-service"),
						Name:      proto.String(fmt.Sprintf("tool-%d", i)),
					}.Build()
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

// TestToolManager_AddTool_NoServiceID ...
// Summary: TestToolManager_AddTool_NoServiceID
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(""), // Empty service ID
				Name:      proto.String("test-tool"),
			}.Build()
		},
	}

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.EqualError(t, err, "tool service ID cannot be empty")
}

// TestToolManager_AddTool_EmptyToolName ...
// Summary: TestToolManager_AddTool_EmptyToolName
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test-service"),
				Name:      proto.String(""), // Empty tool name
			}.Build()
		},
	}

	err := tm.AddTool(mockTool)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name: id cannot be empty")
}

// TestToolManager_AddTool_WithMCPServer ...
// Summary: TestToolManager_AddTool_WithMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	inputSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	assert.NoError(t, err)

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("test-service"),
				Name:        proto.String("test-tool"),
				Description: proto.String("A test tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}

	err = tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Test case where the handler returns an error
	mockTool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test-service"),
				Name:      proto.String("error-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return nil, fmt.Errorf("tool error")
		},
	}

	err = tm.AddTool(mockTool2)
	assert.NoError(t, err)
}

// MockToolExecutionMiddleware ...
// Summary: MockToolExecutionMiddleware
	ExecuteFunc func(ctx context.Context, req *ExecutionRequest, next ExecutionFunc) (any, error)
}

// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, req, next)
	}
	return next(ctx, req)
}

// TestToolManager_AddAndExecuteWithMiddleware ...
// Summary: TestToolManager_AddAndExecuteWithMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)

	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID}

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return "tool success", nil
		},
	}
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)

	// Case 1: Middleware returns a result directly
	middleware1 := &MockToolExecutionMiddleware{
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest, _ ExecutionFunc) (any, error) {
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

// TestToolManager_ClearToolsForService_NoDeletions ...
// Summary: TestToolManager_ClearToolsForService_NoDeletions
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("service-a"),
				Name:      proto.String("tool-1"),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool1)

	tm.ClearToolsForService("service-b") // service-b has no tools
	assert.Len(t, tm.ListTools(), 1, "Should still have one tool")
}

// TestToolManager_AddTool_MCPServerAddToolError ...
// Summary: TestToolManager_AddTool_MCPServerAddToolError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("test"),
				Name:      proto.String(" "),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	err := tm.AddTool(mockTool)
	assert.NoError(t, err)
}

// TestToolManager_AddTool_WithMCPServerAndBus ...
// Summary: TestToolManager_AddTool_WithMCPServerAndBus
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)

	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{}, nil)
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
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("test-service"),
				Name:        proto.String("bus-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return map[string]string{"status": "ok from tool"}, nil
		},
	}
	err = tm.AddTool(mockTool)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("bus-tool")
	toolID := "test-service" + "." + sanitizedToolName
	assert.NotNil(t, toolID)
}

// TestToolManager_AddTool_WithMCPServer_ErrorCases ...
// Summary: TestToolManager_AddTool_WithMCPServer_ErrorCases
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	// Case 1: Error converting proto tool to mcp tool (empty name)
	mockTool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{ServiceId: proto.String("test"), Name: proto.String("")}.Build()
		},
	}
	err := tm.AddTool(mockTool1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sanitize tool name")

	// Case 2: Good tool
	inputSchema, err := structpb.NewStruct(map[string]interface{}{"type": "object"})
	assert.NoError(t, err)

	mockTool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("test"),
				Name:        proto.String("good-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err = tm.AddTool(mockTool2)
	assert.NoError(t, err)
}

// TestToolManager_AddAndGetServiceInfo ...
// Summary: TestToolManager_AddAndGetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
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

// TestToolManager_SetMCPServer ...
// Summary: TestToolManager_SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	provider := &TestMCPServerProvider{server: nil}
	tm.SetMCPServer(provider)
	assert.Equal(t, provider, tm.mcpServer, "MCPServerProvider should be set")
}

// TestToolManager_ExecuteTool_ExecutionError ...
// Summary: TestToolManager_ExecuteTool_ExecutionError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	sanitizedToolName, _ := util.SanitizeToolName("exec-tool")
	toolID := "exec-service" + "." + sanitizedToolName
	execReq := &ExecutionRequest{ToolName: toolID, ToolInputs: []byte(`{"arg":"value"}`)}
	expectedError := fmt.Errorf("execution failed")

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("exec-service"),
				Name:      proto.String("exec-tool"),
			}.Build()
		},
		ExecuteFunc: func(_ context.Context, _ *ExecutionRequest) (any, error) {
			return nil, expectedError
		},
	}

	_ = tm.AddTool(mockTool)

	_, err := tm.ExecuteTool(context.Background(), execReq)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

// TestToolManager_ListTools_Caching ...
// Summary: TestToolManager_ListTools_Caching
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	mockTool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("t1")}.Build()
		},
	}
	_ = tm.AddTool(mockTool1)

	list1 := tm.ListTools()
	assert.Len(t, list1, 1)

	list2 := tm.ListTools()
	assert.Len(t, list2, 1)
	if len(list1) > 0 && len(list2) > 0 {
		assert.Same(t, list1[0], list2[0])
	}

	mockTool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{ServiceId: proto.String("s1"), Name: proto.String("t2")}.Build()
		},
	}
	_ = tm.AddTool(mockTool2)

	list3 := tm.ListTools()
	assert.Len(t, list3, 2)

	tm.ClearToolsForService("s1")
	list4 := tm.ListTools()
	assert.Len(t, list4, 0)
}

// TestToolManager_AddServiceInfo_WithConfig ...
// Summary: TestToolManager_AddServiceInfo_WithConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)
	serviceID := "test-service-config"

	config := configv1.UpstreamServiceConfig_builder{
		CallPolicies: []*configv1.CallPolicy{
			configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_ALLOW.Enum()}.Build(),
		},
		PreCallHooks: []*configv1.CallHook{
			configv1.CallHook_builder{
				Webhook: configv1.WebhookConfig_builder{Url: "http://pre.com"}.Build(),
			}.Build(),
		},
		PostCallHooks: []*configv1.CallHook{
			configv1.CallHook_builder{
				Webhook: configv1.WebhookConfig_builder{Url: "http://post.com"}.Build(),
			}.Build(),
		},
	}.Build()

	serviceInfo := &ServiceInfo{
		Name:   "Test Service Config",
		Config: config,
	}

	tm.AddServiceInfo(serviceID, serviceInfo)

	retrievedInfo, ok := tm.GetServiceInfo(serviceID)
	assert.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	assert.Len(t, retrievedInfo.PreHooks, 2)
	assert.Len(t, retrievedInfo.PostHooks, 1)
}

// TestToolManager_ProfileFiltering ...
// Summary: TestToolManager_ProfileFiltering
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)

	profileDef := configv1.ProfileDefinition_builder{
		Name: proto.String("secure-profile"),
		Selector: configv1.ProfileSelector_builder{
			Tags: []string{"secure"},
		}.Build(),
	}.Build()

	tm.SetProfiles([]string{"secure-profile"}, []*configv1.ProfileDefinition{profileDef})

	tool1 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("s1"),
				Name:      proto.String("t1"),
				Tags:      []string{"secure", "other"},
			}.Build()
		},
	}

	tool2 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("s1"),
				Name:      proto.String("t2"),
				Tags:      []string{"public"},
			}.Build()
		},
	}

	tool3 := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("s1"),
				Name:      proto.String("t3"),
				Profiles:  []string{"secure-profile"},
			}.Build()
		},
	}

	_ = tm.AddTool(tool1)
	_ = tm.AddTool(tool2)
	_ = tm.AddTool(tool3)

	tools := tm.ListTools()
	assert.Len(t, tools, 2)

	foundT1 := false
	foundT3 := false
	for _, tool := range tools {
		if tool.Tool().GetName() == "t1" {
			foundT1 = true
		}
		if tool.Tool().GetName() == "t3" {
			foundT3 = true
		}
	}
	assert.True(t, foundT1)
	assert.True(t, foundT3)
}

// TestToolManager_ProfileFiltering_Properties ...
// Summary: TestToolManager_ProfileFiltering_Properties
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	t.Parallel()
	tm := NewManager(nil)

	profileDef := configv1.ProfileDefinition_builder{
		Name: proto.String("readonly-profile"),
		Selector: configv1.ProfileSelector_builder{
			ToolProperties: map[string]string{
				"read_only": "true",
			},
		}.Build(),
	}.Build()

	tm.SetProfiles([]string{"readonly-profile"}, []*configv1.ProfileDefinition{profileDef})

	toolRO := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("s1"),
				Name:      proto.String("ro"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					ReadOnlyHint: proto.Bool(true),
				}.Build(),
			}.Build()
		},
	}

	toolRW := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String("s1"),
				Name:      proto.String("rw"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					ReadOnlyHint: proto.Bool(false),
				}.Build(),
			}.Build()
		},
	}

	_ = tm.AddTool(toolRO)
	_ = tm.AddTool(toolRW)

	tools := tm.ListTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, "ro", tools[0].Tool().GetName())
}
