package middleware_test

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolManager for testing
type MockToolManager struct {
	mock.Mock
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManager) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	args := m.Called(id)
	if s := args.Get(0); s != nil {
		return s.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// Implement other interface methods as no-ops or panics if needed
func (m *MockToolManager) ListTools() []tool.Tool                           { return nil }
func (m *MockToolManager) ListMCPTools() []*mcp.Tool                        { return nil }
func (m *MockToolManager) AddTool(t tool.Tool) error                        { return nil }
func (m *MockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *MockToolManager) ClearToolsForService(_ string)                   {}
func (m *MockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *MockToolManager) SetMCPServer(_ tool.MCPServerProvider)                                   {}
func (m *MockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool      { return true }
func (m *MockToolManager) AddMiddleware(_ tool.ExecutionMiddleware)               {}
func (m *MockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool      { return true }
func (m *MockToolManager) ListServices() []*tool.ServiceInfo                      { return nil }

type MockTool struct {
	name      string
	serviceID string
}

func (t *MockTool) Tool() *v1.Tool {
	return v1.Tool_builder{Name: proto.String(t.name), ServiceId: proto.String(t.serviceID)}.Build()
}
func (t *MockTool) MCPTool() *mcp.Tool { return &mcp.Tool{Name: t.name} }
func (t *MockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (t *MockTool) GetCacheConfig() *configv1.CacheConfig { return nil }

func TestRateLimitMiddleware_Granular(t *testing.T) {
	tm := &MockToolManager{}
	mw := middleware.NewRateLimitMiddleware(tm)

	serviceID := "test-service"
	toolName := "restricted-tool"
	otherToolName := "normal-tool"

	// Setup Config
	// Service Limit: 100 RPS
	// Tool Limit: 1 RPS (Burst 1)
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		RateLimit: configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 100.0,
			Burst:             100,
			ToolLimits: map[string]*configv1.RateLimitConfig{
				toolName: configv1.RateLimitConfig_builder{
					IsEnabled:         true,
					RequestsPerSecond: 1.0,
					Burst:             1,
				}.Build(),
			},
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&MockTool{name: toolName, serviceID: serviceID}, true)
	tm.On("GetTool", otherToolName).Return(&MockTool{name: otherToolName, serviceID: serviceID}, true)
	tm.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{Name: serviceID, Config: config}, true)

	ctx := context.Background()
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	// Test Case 1: Restricted Tool
	// 1st call should succeed
	req1 := &tool.ExecutionRequest{ToolName: toolName}
	res, err := mw.Execute(ctx, req1, next)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)

	// 2nd call immediate should fail (Burst 1)
	res, err = mw.Execute(ctx, req1, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded for tool")

	// Test Case 2: Normal Tool (falls back to service limit)
	// Should pass easily
	req2 := &tool.ExecutionRequest{ToolName: otherToolName}
	res, err = mw.Execute(ctx, req2, next)
	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestRateLimitMiddleware_ServiceLimitFallback(t *testing.T) {
	tm := &MockToolManager{}
	mw := middleware.NewRateLimitMiddleware(tm)

	serviceID := "test-service-fallback"
	toolName := "normal-tool"

	// Service Limit: 1 RPS
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		RateLimit: configv1.RateLimitConfig_builder{
			IsEnabled:         true,
			RequestsPerSecond: 1.0,
			Burst:             1,
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&MockTool{name: toolName, serviceID: serviceID}, true)
	tm.On("GetServiceInfo", serviceID).Return(&tool.ServiceInfo{Name: serviceID, Config: config}, true)

	ctx := context.Background()
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	req := &tool.ExecutionRequest{ToolName: toolName}

	// 1st call ok
	_, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)

	// 2nd call blocked
	_, err = mw.Execute(ctx, req, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded for service")
}

func (m *MockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

func (m *MockToolManager) GetToolCountForService(serviceID string) int {
	return 0
}
