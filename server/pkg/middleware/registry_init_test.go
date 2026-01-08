// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolManagerForRegistry is a mock implementation of tool.ManagerInterface
type MockToolManagerForRegistry struct {
	mock.Mock
}

func (m *MockToolManagerForRegistry) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if t := args.Get(0); t != nil {
		return t.(tool.Tool), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManagerForRegistry) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

func (m *MockToolManagerForRegistry) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockToolManagerForRegistry) AddServiceInfo(serviceName string, config *tool.ServiceInfo) {
	m.Called(serviceName, config)
}

func (m *MockToolManagerForRegistry) GetServiceInfo(serviceName string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceName)
	if c := args.Get(0); c != nil {
		return c.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockToolManagerForRegistry) ClearToolsForService(serviceName string) {
	m.Called(serviceName)
}

func (m *MockToolManagerForRegistry) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockToolManagerForRegistry) SetMCPServer(mcpServer tool.MCPServerProvider) {
	m.Called(mcpServer)
}

func (m *MockToolManagerForRegistry) AddMiddleware(middleware tool.ExecutionMiddleware) {
	m.Called(middleware)
}

func (m *MockToolManagerForRegistry) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

func (m *MockToolManagerForRegistry) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}

func (m *MockToolManagerForRegistry) IsServiceAllowed(serviceID, profileID string) bool {
	return true
}

func (m *MockToolManagerForRegistry) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}

func TestInitStandardMiddlewares(t *testing.T) {
	// Setup dependencies
	authManager := auth.NewManager()
	mockToolManager := new(MockToolManagerForRegistry)
	auditConfig := &configv1.AuditConfig{
		Enabled: proto.Bool(false), // Disable to avoid file creation issues in test
	}
	cachingMiddleware := &CachingMiddleware{}

	// Call InitStandardMiddlewares
	standardMiddlewares, err := InitStandardMiddlewares(authManager, mockToolManager, auditConfig, cachingMiddleware, nil)
	assert.NoError(t, err)
	assert.NotNil(t, standardMiddlewares)
	if standardMiddlewares.Cleanup != nil {
		defer standardMiddlewares.Cleanup()
	}

	// Verify standard middlewares are registered in MCP registry
	expectedMiddlewares := []string{"logging", "auth", "debug", "cors", "caching", "ratelimit", "call_policy", "audit", "global_ratelimit"}

	globalRegistry.mu.RLock()
	for _, name := range expectedMiddlewares {
		assert.Contains(t, globalRegistry.mcpFactories, name, "Middleware %s should be registered", name)
	}
	globalRegistry.mu.RUnlock()

	// Test execution of registered middlewares to ensure they are created correctly
	// We'll test a few key ones wrapped in the factory

	// Helper to create a dummy next handler
	dummyNext := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "success"},
			},
		}, nil
	}

	t.Run("caching_execution", func(t *testing.T) {
		factory := globalRegistry.mcpFactories["caching"]
		handler := factory(&configv1.Middleware{})(dummyNext)

		// Mock cache execution if possible, or just checking it doesn't panic
		assert.NotNil(t, handler)
	})

	t.Run("ratelimit_execution", func(t *testing.T) {
		factory := globalRegistry.mcpFactories["ratelimit"]
		handler := factory(&configv1.Middleware{})(dummyNext)
		assert.NotNil(t, handler)
	})

	t.Run("call_policy_execution", func(t *testing.T) {
		factory := globalRegistry.mcpFactories["call_policy"]
		handler := factory(&configv1.Middleware{})(dummyNext)
		assert.NotNil(t, handler)
	})

	t.Run("audit_execution", func(t *testing.T) {
		factory := globalRegistry.mcpFactories["audit"]
		handler := factory(&configv1.Middleware{})(dummyNext)
		assert.NotNil(t, handler)
	})

	t.Run("global_ratelimit_execution", func(t *testing.T) {
		factory := globalRegistry.mcpFactories["global_ratelimit"]
		handler := factory(&configv1.Middleware{})(dummyNext)
		assert.NotNil(t, handler)
	})
}

func TestInitStandardMiddlewares_AuditError(t *testing.T) {
	// Setup dependencies
	authManager := auth.NewManager()
	mockToolManager := new(MockToolManagerForRegistry)

	// Invalid audit config to force error
	st := configv1.AuditConfig_STORAGE_TYPE_SQLITE
	auditConfig := &configv1.AuditConfig{
		Enabled:     proto.Bool(true),
		StorageType: &st,
		OutputPath:  proto.String("/invalid/path/that/does/not/exist/audit.db"),
	}
	cachingMiddleware := &CachingMiddleware{}

	// Call InitStandardMiddlewares
	standardMiddlewares, err := InitStandardMiddlewares(authManager, mockToolManager, auditConfig, cachingMiddleware, nil)
	assert.Error(t, err)
	assert.Nil(t, standardMiddlewares)
}
