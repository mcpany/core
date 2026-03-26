// Copyright 2026 Author(s) of MCP Any.
// SPDX-License-Identifier: Apache-2.0.

package tool

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/mock"
)

// MockToolManager is a mock of ManagerInterface.
type MockToolManager struct {
	mock.Mock
}

// AddTool is a mock method.
func (m *MockToolManager) AddTool(t Tool) error {
	return m.Called(t).Error(0)
}

// GetTool is a mock method.
func (m *MockToolManager) GetTool(name string) (Tool, bool) {
	args := m.Called(name)
	return args.Get(0).(Tool), args.Bool(1)
}

// ListTools is a mock method.
func (m *MockToolManager) ListTools() []Tool {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]Tool)
}

// ListMCPTools is a mock method.
func (m *MockToolManager) ListMCPTools() []*mcp.Tool {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*mcp.Tool)
}

// ClearToolsForService is a mock method.
func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.Called(serviceID)
}

// ExecuteTool is a mock method.
func (m *MockToolManager) ExecuteTool(ctx context.Context, req *ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// SetMCPServer is a mock method.
func (m *MockToolManager) SetMCPServer(mcpServer MCPServerProvider) {
	m.Called(mcpServer)
}

// AddMiddleware is a mock method.
func (m *MockToolManager) AddMiddleware(middleware ExecutionMiddleware) {
	m.Called(middleware)
}

// AddServiceInfo is a mock method.
func (m *MockToolManager) AddServiceInfo(serviceID string, info *ServiceInfo) {
	m.Called(serviceID, info)
}

// GetServiceInfo is a mock method.
func (m *MockToolManager) GetServiceInfo(serviceID string) (*ServiceInfo, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*ServiceInfo), args.Bool(1)
}

// ListServices is a mock method.
func (m *MockToolManager) ListServices() []*ServiceInfo {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*ServiceInfo)
}

// SetProfiles is a mock method.
func (m *MockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}

// IsServiceAllowed is a mock method.
func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return m.Called(serviceID, profileID).Bool(0)
}

// ToolMatchesProfile is a mock method.
func (m *MockToolManager) ToolMatchesProfile(t Tool, profileID string) bool {
	return m.Called(t, profileID).Bool(0)
}

// GetAllowedServiceIDs is a mock method.
func (m *MockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

// GetToolCountForService is a mock method.
func (m *MockToolManager) GetToolCountForService(serviceID string) int {
	return m.Called(serviceID).Int(0)
}
