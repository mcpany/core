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
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock implementation of tool.ManagerInterface
type MockToolManager struct {
	mock.Mock
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if t, ok := args.Get(0).(tool.Tool); ok {
		return t, args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

func (m *MockToolManager) AddServiceInfo(id string, info *tool.ServiceInfo) {
	m.Called(id, info)
}

func (m *MockToolManager) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	args := m.Called(id)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {
	m.Called(mcpServer)
}

func (m *MockToolManager) AddMiddleware(middleware tool.ExecutionMiddleware) {
	m.Called(middleware)
}

func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

func (m *MockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}

func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return true
}

func (m *MockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}

// MockTool is a mock implementation of tool.Tool
type MockTool struct {
	mock.Mock
}

func (m *MockTool) Tool() *mcp_router_v1.Tool {
	args := m.Called()
	if t, ok := args.Get(0).(*mcp_router_v1.Tool); ok {
		return t
	}
	return nil
}

func (m *MockTool) MCPTool() *mcp.Tool {
	args := m.Called()
	if t, ok := args.Get(0).(*mcp.Tool); ok {
		return t
	}
	return nil
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	args := m.Called()
	if c, ok := args.Get(0).(*configv1.CacheConfig); ok {
		return c
	}
	return nil
}

// MockServiceRegistry is a mock implementation of serviceregistry.ServiceRegistryInterface
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if config, ok := args.Get(0).(*configv1.UpstreamServiceConfig); ok {
		return config, args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestManager_RecordActivity(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	m.RecordActivity("session-1", map[string]interface{}{
		"userAgent": "test-agent",
		"count":     1, // Should be ignored as it's not a string
	})

	m.mu.RLock()
	session, exists := m.sessions["session-1"]
	lastActive := session.LastActive
	m.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, "session-1", session.ID)
	assert.Equal(t, "test-agent", session.Metadata["userAgent"])
	assert.NotZero(t, session.LastActive)
	assert.Equal(t, int64(1), session.RequestCount)

	// Record again to update stats
	time.Sleep(100 * time.Millisecond) // Ensure time advances
	m.RecordActivity("session-1", nil)

	m.mu.RLock()
	session2 := m.sessions["session-1"]
	m.mu.RUnlock()

	assert.True(t, session2.LastActive.After(lastActive))
	assert.Equal(t, int64(2), session2.RequestCount)
}

func TestManager_GetGraph(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	// Add a service
	svcConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	// Add a tool to mock manager
	mockTool := new(MockTool)
	mockTool.On("Tool").Return(&mcp_router_v1.Tool{
		Name:      proto.String("test-tool"),
		ServiceId: proto.String("test-service"),
	})
	mockTM.On("ListTools").Return([]tool.Tool{mockTool})

	// Record a session
	m.RecordActivity("session-1", map[string]interface{}{"userAgent": "client-1"})

	graph := m.GetGraph(context.Background())

	// Check Core
	assert.Equal(t, "mcp-core", graph.Core.Id)
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_CORE, graph.Core.Type)

	// Check Children (Services, Middleware, Webhooks)
	// We expect: test-service, middleware-pipeline, webhooks
	assert.True(t, len(graph.Core.Children) >= 3)

	var svcNode *topologyv1.Node
	for _, child := range graph.Core.Children {
		if child.Id == "svc-test-service" {
			svcNode = child
			break
		}
	}
	require.NotNil(t, svcNode)
	assert.Equal(t, "test-service", svcNode.Label)
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_SERVICE, svcNode.Type)

	// Check Tool inside Service
	require.Len(t, svcNode.Children, 1)
	toolNode := svcNode.Children[0]
	assert.Equal(t, "tool-test-tool", toolNode.Id)
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_TOOL, toolNode.Type)

	// Check API Call inside Tool
	require.Len(t, toolNode.Children, 1)
	apiNode := toolNode.Children[0]
	assert.Equal(t, "api-test-tool", apiNode.Id)
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_API_CALL, apiNode.Type)

	// Check Client
	require.Len(t, graph.Clients, 1)
	clientNode := graph.Clients[0]
	assert.Equal(t, "client-session-1", clientNode.Id)
	assert.Equal(t, "client-1", clientNode.Label)
	assert.Equal(t, topologyv1.NodeType_NODE_TYPE_CLIENT, clientNode.Type)
}

func TestManager_Middleware(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)

	nextHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	wrapped := m.Middleware(nextHandler)

	// Test case 1: Authenticated User
	ctx := auth.ContextWithUser(context.Background(), "user123")
	_, err := wrapped(ctx, "method", &mcp.CallToolRequest{})
	require.NoError(t, err)

	m.mu.RLock()
	session, exists := m.sessions["user-user123"]
	m.mu.RUnlock()
	require.True(t, exists)
	assert.Equal(t, "authenticated_user", session.Metadata["type"])

	// Test case 2: Anonymous IP
	ctx = context.WithValue(context.Background(), consts.ContextKeyRemoteAddr, "127.0.0.1")
	_, err = wrapped(ctx, "method", &mcp.CallToolRequest{})
	require.NoError(t, err)

	m.mu.RLock()
	session, exists = m.sessions["ip-127.0.0.1"]
	m.mu.RUnlock()
	require.True(t, exists)
	assert.Equal(t, "anonymous_ip", session.Metadata["type"])

	// Test case 3: Unknown
	_, err = wrapped(context.Background(), "method", &mcp.CallToolRequest{})
	require.NoError(t, err)

	m.mu.RLock()
	session, exists = m.sessions["unknown"]
	m.mu.RUnlock()
	require.True(t, exists)
}

func TestManager_GetGraph_InactiveService(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	mockTM.On("ListTools").Return([]tool.Tool{})
	m := NewManager(mockRegistry, mockTM)

	svcConfig := &configv1.UpstreamServiceConfig{
		Name:    proto.String("disabled-service"),
		Disable: proto.Bool(true),
	}
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svcConfig}, nil)

	graph := m.GetGraph(context.Background())

	var svcNode *topologyv1.Node
	for _, child := range graph.Core.Children {
		if child.Id == "svc-disabled-service" {
			svcNode = child
			break
		}
	}
	require.NotNil(t, svcNode)
	assert.Equal(t, topologyv1.NodeStatus_NODE_STATUS_INACTIVE, svcNode.Status)
}

func TestManager_GetGraph_OldSession(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	mockTM.On("ListTools").Return([]tool.Tool{})
	// Mock GetAllServices for GetGraph call
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{}, nil)

	m := NewManager(mockRegistry, mockTM)

	m.RecordActivity("old-session", nil)
	// Manually age the session
	m.mu.Lock()
	m.sessions["old-session"].LastActive = time.Now().Add(-2 * time.Hour)
	m.mu.Unlock()

	graph := m.GetGraph(context.Background())

	assert.Len(t, graph.Clients, 0)
}
