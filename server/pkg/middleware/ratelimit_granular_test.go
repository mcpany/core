package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockToolManager) GetTool(toolName string) (tool.Tool, bool) {
	args := m.Called(toolName)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
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

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}

func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if s := args.Get(0); s != nil {
		return s.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

func (m *MockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}

// MockTool is a mock implementation of tool.Tool
type MockTool struct {
	mock.Mock
}

func (m *MockTool) Tool() *v1.Tool {
	args := m.Called()
	if t := args.Get(0); t != nil {
		return t.(*v1.Tool)
	}
	return nil
}

func (m *MockTool) MCPTool() *mcp.Tool {
	args := m.Called()
	if t := args.Get(0); t != nil {
		return t.(*mcp.Tool)
	}
	return nil
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	args := m.Called()
	if c := args.Get(0); c != nil {
		return c.(*configv1.CacheConfig)
	}
	return nil
}

func enumPtr[T any](v T) *T { return &v }

func TestRateLimitMiddleware_Granular(t *testing.T) {
	// Setup
	mockManager := new(MockToolManager)
	mw := NewRateLimitMiddleware(mockManager)

	// Define Tool and Service
	toolName := "my_tool"
	serviceID := "my_service"
	mockTool := new(MockTool)
	// Using proto helpers
	mockTool.On("Tool").Return(&v1.Tool{Name: proto.String(toolName), ServiceId: proto.String(serviceID)})

	mockManager.On("GetTool", toolName).Return(mockTool, true)

	// Rate Limit Config with Granularity
	rlConfig := &configv1.RateLimitConfig{
		IsEnabled: proto.Bool(true),
		RequestsPerSecond: proto.Float64(1),
		Burst: proto.Int64(1),
		Storage: enumPtr(configv1.RateLimitConfig_STORAGE_MEMORY),
		KeyBy: enumPtr(configv1.RateLimitConfig_KEY_BY_IP),
	}

	serviceInfo := &tool.ServiceInfo{
		Name: "My Service",
		Config: &configv1.UpstreamServiceConfig{
			RateLimit: rlConfig,
		},
	}
	mockManager.On("GetServiceInfo", serviceID).Return(serviceInfo, true)

	// Test Case 1: IP A makes a request
	ctxA := util.ContextWithRemoteIP(context.Background(), "1.2.3.4")
	req := &tool.ExecutionRequest{ToolName: toolName}

	// First request should pass
	executed := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		executed = true
		return "success", nil
	}
	_, err := mw.Execute(ctxA, req, next)
	assert.NoError(t, err)
	assert.True(t, executed)

	// Second request from IP A should fail (limit 1, burst 1)
	executed = false
	_, err = mw.Execute(ctxA, req, next)
	assert.Error(t, err)
	assert.False(t, executed)
	assert.Contains(t, err.Error(), "rate limit exceeded")

	// Test Case 2: IP B makes a request (should pass, different bucket)
	ctxB := util.ContextWithRemoteIP(context.Background(), "5.6.7.8")
	executed = false
	_, err = mw.Execute(ctxB, req, next)
	assert.NoError(t, err)
	assert.True(t, executed)

	// Wait 1.1s to let refill happen
	time.Sleep(1100 * time.Millisecond)
	// IP A should be allowed again
	executed = false
	_, err = mw.Execute(ctxA, req, next)
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestRateLimitMiddleware_Granular_APIKey(t *testing.T) {
	// Setup
	mockManager := new(MockToolManager)
	mw := NewRateLimitMiddleware(mockManager)

	// Define Tool and Service
	toolName := "my_tool"
	serviceID := "my_service"
	mockTool := new(MockTool)
	// Using proto helpers
	mockTool.On("Tool").Return(&v1.Tool{Name: proto.String(toolName), ServiceId: proto.String(serviceID)})

	mockManager.On("GetTool", toolName).Return(mockTool, true)

	// Rate Limit Config with Granularity
	rlConfig := &configv1.RateLimitConfig{
		IsEnabled: proto.Bool(true),
		RequestsPerSecond: proto.Float64(1),
		Burst: proto.Int64(1),
		Storage: enumPtr(configv1.RateLimitConfig_STORAGE_MEMORY),
		KeyBy: enumPtr(configv1.RateLimitConfig_KEY_BY_API_KEY),
	}

	serviceInfo := &tool.ServiceInfo{
		Name: "My Service",
		Config: &configv1.UpstreamServiceConfig{
			RateLimit: rlConfig,
		},
	}
	mockManager.On("GetServiceInfo", serviceID).Return(serviceInfo, true)

	// Test Case 1: API Key A makes a request
	ctxA := auth.ContextWithAPIKey(context.Background(), "key-A")
	req := &tool.ExecutionRequest{ToolName: toolName}

	executed := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		executed = true
		return "success", nil
	}

	// First request should pass
	_, err := mw.Execute(ctxA, req, next)
	assert.NoError(t, err)
	assert.True(t, executed)

	// Second request from API Key A should fail
	executed = false
	_, err = mw.Execute(ctxA, req, next)
	assert.Error(t, err)
	assert.False(t, executed)

	// Test Case 2: API Key B makes a request (should pass)
	ctxB := auth.ContextWithAPIKey(context.Background(), "key-B")
	executed = false
	_, err = mw.Execute(ctxB, req, next)
	assert.NoError(t, err)
	assert.True(t, executed)
}
