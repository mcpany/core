// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	topologyv1 "github.com/mcpany/core/proto/topology/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/tool"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock implementation of tool.ManagerInterface
// MockToolManager is a mock implementation of tool.ManagerInterface
// Summary: MockToolManager
	mock.Mock
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(t)
	return args.Error(0)
}

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(name)
	if t, ok := args.Get(0).(tool.Tool); ok {
		return t, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

// ListMCPTools ...
// Summary: ListMCPTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]*mcp_sdk.Tool)
}

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(id, info)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(id)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID)
}

// ExecuteTool ...
// Summary: ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(mcpServer)
}

// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(middleware)
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(enabled, defs)
}

// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// ToolMatchesProfile ...
// Summary: ToolMatchesProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

// GetToolCountForService ...
// Summary: GetToolCountForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0
}

// MockTool is a mock implementation of tool.Tool
// MockTool is a mock implementation of tool.Tool
// Summary: MockTool
	mock.Mock
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	if t, ok := args.Get(0).(*mcp_router_v1.Tool); ok {
		return t
	}
	return nil
}

// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	if t, ok := args.Get(0).(*mcp_sdk.Tool); ok {
		return t
	}
	return nil
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
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// GetCacheConfig ...
// Summary: GetCacheConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	if c, ok := args.Get(0).(*configv1.CacheConfig); ok {
		return c
	}
	return nil
}

// MockServiceRegistry is a mock implementation of serviceregistry.ServiceRegistryInterface
// MockServiceRegistry is a mock implementation of serviceregistry.ServiceRegistryInterface
// Summary: MockServiceRegistry
	mock.Mock
}

// RegisterService ...
// Summary: RegisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

// UnregisterService ...
// Summary: UnregisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

// GetAllServices ...
// Summary: GetAllServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceConfig ...
// Summary: GetServiceConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if config, ok := args.Get(0).(*configv1.UpstreamServiceConfig); ok {
		return config, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceError ...
// Summary: GetServiceError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

// TestManager_RecordActivity ...
// Summary: TestManager_RecordActivity
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	m.RecordActivity("session-1", map[string]interface{}{
		"userAgent": "test-agent",
		"count":     1,
	}, 10*time.Millisecond, false, "", 100)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["session-1"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session, exists := m.sessions["session-1"]
	lastActive := session.LastActive
	m.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, "session-1", session.ID)
	assert.Equal(t, "test-agent", session.Metadata["userAgent"])
	assert.NotZero(t, session.LastActive)
	assert.Equal(t, int64(1), session.RequestCount)
	assert.Equal(t, 10*time.Millisecond, session.TotalLatency)
	assert.Equal(t, int64(100), session.TotalBytes)

	time.Sleep(100 * time.Millisecond)
	m.RecordActivity("session-1", nil, 20*time.Millisecond, true, "", 200)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		s := m.sessions["session-1"]
		count := s.RequestCount
		m.mu.RUnlock()
		return count == 2
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session2 := m.sessions["session-1"]
	m.mu.RUnlock()

	assert.True(t, session2.LastActive.After(lastActive))
	assert.Equal(t, int64(2), session2.RequestCount)
	assert.Equal(t, 30*time.Millisecond, session2.TotalLatency)
	assert.Equal(t, int64(1), session2.ErrorCount)
	assert.Equal(t, int64(300), session2.TotalBytes)
}

// TestManager_GetStats ...
// Summary: TestManager_GetStats
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	m.RecordActivity("session-1", nil, 100*time.Millisecond, false, "", 0)
	m.RecordActivity("session-2", nil, 200*time.Millisecond, true, "", 0)

	assert.Eventually(t, func() bool {
		stats := m.GetStats("")
		return stats.TotalRequests == 2
	}, 1*time.Second, 10*time.Millisecond)

	stats := m.GetStats("")

	assert.Equal(t, int64(2), stats.TotalRequests)
	assert.Equal(t, 150*time.Millisecond, stats.AvgLatency)
	assert.Equal(t, 0.5, stats.ErrorRate)
}

// TestManager_GetGraph ...
// Summary: TestManager_GetGraph
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	mockTool := new(MockTool)
	mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
		Name:      proto.String("test-tool"),
		ServiceId: proto.String("test-service"),
	}.Build())
	mockTM.On("ListTools").Return([]tool.Tool{mockTool})

	m.RecordActivity("session-1", map[string]interface{}{"userAgent": "client-1"}, 10*time.Millisecond, false, "", 0)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["session-1"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	graph := m.GetGraph(context.Background())

	assert.Equal(t, "mcp-core", graph.GetCore().GetId())
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_CORE, graph.GetCore().GetType())

	assert.True(t, len(graph.GetCore().GetChildren()) >= 3)

	var svcNode *topologyv1.Node
	for _, child := range graph.GetCore().GetChildren() {
		if child.GetId() == "svc-test-service" {
			svcNode = child
			break
		}
	}
	require.NotNil(t, svcNode)
	assert.Equal(t, "test-service", svcNode.GetLabel())
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_SERVICE, svcNode.GetType())

	require.Len(t, svcNode.GetChildren(), 1)
	toolNode := svcNode.GetChildren()[0]
	assert.Equal(t, "tool-test-tool", toolNode.GetId())
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_TOOL, toolNode.GetType())

	require.Len(t, toolNode.GetChildren(), 1)
	apiNode := toolNode.GetChildren()[0]
	assert.Equal(t, "api-test-tool", apiNode.GetId())
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_API_CALL, apiNode.GetType())

	require.Len(t, graph.GetClients(), 1)
	clientNode := graph.GetClients()[0]
	assert.Equal(t, "client-session-1", clientNode.GetId())
	assert.Equal(t, "client-1", clientNode.GetLabel())
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_CLIENT, clientNode.GetType())
}

// TestManager_Middleware ...
// Summary: TestManager_Middleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	nextHandler := func(ctx context.Context, method string, req mcp_sdk.Request) (mcp_sdk.Result, error) {
		return &mcp_sdk.CallToolResult{}, nil
	}

	wrapped := m.Middleware(nextHandler)

	ctx := auth.ContextWithUser(context.Background(), "user123")
	_, err := wrapped(ctx, "method", &mcp_sdk.CallToolRequest{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["user-user123"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session, exists := m.sessions["user-user123"]
	m.mu.RUnlock()
	require.True(t, exists)
	assert.Equal(t, "authenticated_user", session.Metadata["type"])

	ctx = context.WithValue(context.Background(), consts.ContextKeyRemoteAddr, "127.0.0.1")
	_, err = wrapped(ctx, "method", &mcp_sdk.CallToolRequest{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["ip-127.0.0.1"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session, exists = m.sessions["ip-127.0.0.1"]
	m.mu.RUnlock()
	require.True(t, exists)
	assert.Equal(t, "anonymous_ip", session.Metadata["type"])

	_, err = wrapped(context.Background(), "method", &mcp_sdk.CallToolRequest{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["unknown"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session, exists = m.sessions["unknown"]
	m.mu.RUnlock()
	require.True(t, exists)

	mockTool := new(MockTool)
	mockTool.On("Tool").Return(mcp_router_v1.Tool_builder{
		Name:      proto.String("test-tool"),
		ServiceId: proto.String("extracted-service"),
	}.Build())
	mockTM.On("GetTool", "test-tool").Return(mockTool, true)

	ctx = context.Background()
	req := &mcp_sdk.CallToolRequest{
		Params: &mcp_sdk.CallToolParamsRaw{
			Name: "test-tool",
		},
	}
	_, err = wrapped(ctx, "tools/call", req)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		s := m.sessions["unknown"]
		if s == nil {
			m.mu.RUnlock()
			return false
		}
		count, ok := s.ServiceCounts["extracted-service"]
		m.mu.RUnlock()
		return ok && count == 1
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.RLock()
	session, exists = m.sessions["unknown"]
	require.True(t, exists)
	count := session.ServiceCounts["extracted-service"]
	m.mu.RUnlock()
	assert.Equal(t, int64(1), count)
}

// TestManager_GetGraph_InactiveService ...
// Summary: TestManager_GetGraph_InactiveService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	mockTM.On("ListTools").Return([]tool.Tool{})
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name:    proto.String("disabled-service"),
		Disable: proto.Bool(true),
	}.Build()
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	graph := m.GetGraph(context.Background())

	var svcNode *topologyv1.Node
	for _, child := range graph.GetCore().GetChildren() {
		if child.GetId() == "svc-disabled-service" {
			svcNode = child
			break
		}
	}
	require.NotNil(t, svcNode)
	assert.Equal(t, topologyv1.NodeStatus_NODE_STATUS_INACTIVE, svcNode.GetStatus())
}

// TestManager_GetGraph_OldSession ...
// Summary: TestManager_GetGraph_OldSession
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	mockTM.On("ListTools").Return([]tool.Tool{})
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{}, nil)

	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	m.RecordActivity("old-session", nil, 0, false, "", 0)

	assert.Eventually(t, func() bool {
		m.mu.RLock()
		_, exists := m.sessions["old-session"]
		m.mu.RUnlock()
		return exists
	}, 1*time.Second, 10*time.Millisecond)

	m.mu.Lock()
	m.sessions["old-session"].LastActive = time.Now().Add(-2 * time.Hour)
	m.mu.Unlock()

	graph := m.GetGraph(context.Background())

	assert.Len(t, graph.GetClients(), 0)
}

// TestManager_GetTrafficHistory ...
// Summary: TestManager_GetTrafficHistory
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	m.RecordActivity("session-1", nil, 100*time.Millisecond, false, "", 500)

	assert.Eventually(t, func() bool {
		history := m.GetTrafficHistory("")
		if len(history) == 0 {
			return false
		}
		lastPoint := history[len(history)-1]
		return lastPoint.Total == 1
	}, 1*time.Second, 10*time.Millisecond)

	history := m.GetTrafficHistory("")
	require.NotEmpty(t, history)

	lastPoint := history[len(history)-1]

	assert.Equal(t, int64(1), lastPoint.Total)
	assert.Equal(t, int64(0), lastPoint.Errors)
	assert.Equal(t, int64(100), lastPoint.Latency)
	assert.Equal(t, int64(500), lastPoint.Bytes)
}

// TestManager_SeedTrafficHistory ...
// Summary: TestManager_SeedTrafficHistory
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	points := []TrafficPoint{
		{Time: "12:00", Total: 10, Errors: 1, Latency: 50, Bytes: 100},
		{Time: "12:01", Total: 20, Errors: 2, Latency: 60, Bytes: 200},
	}

	m.SeedTrafficHistory(points)

	m.mu.RLock()
	defer m.mu.RUnlock()

	assert.NotEmpty(t, m.trafficHistory)
	assert.Equal(t, 2, len(m.trafficHistory))

	seedSession := m.sessions["seed-data"]
	require.NotNil(t, seedSession)
	assert.Equal(t, int64(30), seedSession.RequestCount)
	assert.Equal(t, int64(3), seedSession.ErrorCount)
	assert.Equal(t, 1700*time.Millisecond, seedSession.TotalLatency)
	assert.Equal(t, int64(300), seedSession.TotalBytes)
}

// TestManager_GetGraph_Metrics ...
// Summary: TestManager_GetGraph_Metrics
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	mockTM.On("ListTools").Return([]tool.Tool{})
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	svcConfig := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	// Simulate high traffic with errors
	now := time.Now()
	minuteKey := now.Truncate(time.Minute).Unix()

	m.mu.Lock()
	m.trafficHistory[minuteKey] = &MinuteStats{
		ServiceStats: map[string]*ServiceTrafficStats{
			"test-service": {
				Requests: 100,
				Errors:   10,   // 10% error rate
				Latency:  5000, // 50ms average (5000ms total / 100 reqs)
			},
		},
	}
	m.mu.Unlock()

	graph := m.GetGraph(context.Background())

	var svcNode *topologyv1.Node
	for _, child := range graph.GetCore().GetChildren() {
		if child.GetId() == "svc-test-service" {
			svcNode = child
			break
		}
	}
	require.NotNil(t, svcNode)

	// Verify Metrics
	require.NotNil(t, svcNode.GetMetrics())
	// QPS depends on how many seconds passed in current minute.
	// We can't easily assert exact QPS without mocking time, but it should be > 0.
	assert.True(t, svcNode.GetMetrics().GetQps() > 0)
	assert.Equal(t, 0.1, svcNode.GetMetrics().GetErrorRate())
	assert.Equal(t, 50.0, svcNode.GetMetrics().GetLatencyMs())

	// Verify Status Upgrade to ERROR (since error rate > 5%)
	assert.Equal(t, topologyv1.NodeStatus_NODE_STATUS_ERROR, svcNode.GetStatus())
}
