// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

// GranularMockToolManager for testing
type GranularMockToolManager struct {
	mock.Mock
}

func (m *GranularMockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *GranularMockToolManager) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	args := m.Called(id)
	if s := args.Get(0); s != nil {
		return s.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// Implement other interface methods as no-ops or panics if needed
func (m *GranularMockToolManager) ListTools() []tool.Tool                           { return nil }
func (m *GranularMockToolManager) ListMCPTools() []*mcp.Tool                        { return nil }
func (m *GranularMockToolManager) AddTool(t tool.Tool) error                        { return nil }
func (m *GranularMockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *GranularMockToolManager) ClearToolsForService(_ string)                   {}
func (m *GranularMockToolManager) ExecuteTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *GranularMockToolManager) SetMCPServer(_ tool.MCPServerProvider)                                   {}
func (m *GranularMockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
func (m *GranularMockToolManager) IsServiceAllowed(serviceID, profileID string) bool      { return true }
func (m *GranularMockToolManager) AddMiddleware(_ tool.ExecutionMiddleware)               {}
func (m *GranularMockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool      { return true }
func (m *GranularMockToolManager) ListServices() []*tool.ServiceInfo                      { return nil }
func (m *GranularMockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

type GranularMockTool struct {
	name      string
	serviceID string
}

func (t *GranularMockTool) Tool() *v1.Tool {
	return &v1.Tool{Name: &t.name, ServiceId: &t.serviceID}
}
func (t *GranularMockTool) MCPTool() *mcp.Tool { return &mcp.Tool{Name: t.name} }
func (t *GranularMockTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (t *GranularMockTool) GetCacheConfig() *configv1.CacheConfig { return nil }
func (t *GranularMockTool) Service() string { return t.serviceID }

func TestRateLimitMiddleware_Granular(t *testing.T) {
	tm := &GranularMockToolManager{}
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
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(100),
			Burst:             proto.Int64(100),
			ToolLimits: map[string]*configv1.RateLimitConfig{
				toolName: configv1.RateLimitConfig_builder{
					IsEnabled:         proto.Bool(true),
					RequestsPerSecond: proto.Float64(1),
					Burst:             proto.Int64(1),
				}.Build(),
			},
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&GranularMockTool{name: toolName, serviceID: serviceID}, true)
	tm.On("GetTool", otherToolName).Return(&GranularMockTool{name: otherToolName, serviceID: serviceID}, true)
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
	tm := &GranularMockToolManager{}
	mw := middleware.NewRateLimitMiddleware(tm)

	serviceID := "test-service-fallback"
	toolName := "normal-tool"

	// Service Limit: 1 RPS
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(serviceID),
		RateLimit: configv1.RateLimitConfig_builder{
			IsEnabled:         proto.Bool(true),
			RequestsPerSecond: proto.Float64(1),
			Burst:             proto.Int64(1),
		}.Build(),
	}.Build()

	tm.On("GetTool", toolName).Return(&GranularMockTool{name: toolName, serviceID: serviceID}, true)
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
