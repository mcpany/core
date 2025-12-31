// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
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
func (m *MockToolManager) AddTool(t tool.Tool) error                        { return nil }
func (m *MockToolManager) AddServiceInfo(id string, info *tool.ServiceInfo) {}
func (m *MockToolManager) ClearToolsForService(id string)                   {}
func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *MockToolManager) SetMCPServer(s tool.MCPServerProvider)                                   {}
func (m *MockToolManager) SetProfiles(enabled []string, definitions []*configv1.ProfileDefinition) {}
func (m *MockToolManager) AddMiddleware(mw tool.ExecutionMiddleware)                               {}
func (m *MockToolManager) ListServices() []*tool.ServiceInfo                                       { return nil }

type MockTool struct {
	name      string
	serviceID string
}

func (t *MockTool) Tool() *v1.Tool {
	return &v1.Tool{Name: &t.name, ServiceId: &t.serviceID}
}
func (t *MockTool) MCPTool() *mcp.Tool { return &mcp.Tool{Name: t.name} }
func (t *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
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
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String(serviceID),
		RateLimit: &configv1.RateLimitConfig{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(100),
			Burst:             proto.Int64(100),
			ToolLimits: map[string]*configv1.RateLimitConfig{
				toolName: {
					IsEnabled:         proto.Bool(true),
					RequestsPerSecond: proto.Float64(1),
					Burst:             proto.Int64(1),
				},
			},
		},
	}

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
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String(serviceID),
		RateLimit: &configv1.RateLimitConfig{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
		},
	}

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
