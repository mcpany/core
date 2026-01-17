// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockToolManager
type MockToolManager struct {
	mock.Mock
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	return m.Called(t).Error(0)
}
func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}
func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}
func (m *MockToolManager) ListMCPTools() []*mcp.Tool {
	args := m.Called()
	return args.Get(0).([]*mcp.Tool)
}
func (m *MockToolManager) ClearToolsForService(id string) {
	m.Called(id)
}
func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}
func (m *MockToolManager) SetMCPServer(srv tool.MCPServerProvider) {
	m.Called(srv)
}
func (m *MockToolManager) AddMiddleware(mw tool.ExecutionMiddleware) {
	m.Called(mw)
}
func (m *MockToolManager) AddServiceInfo(id string, info *tool.ServiceInfo) {
	m.Called(id, info)
}
func (m *MockToolManager) GetServiceInfo(id string) (*tool.ServiceInfo, bool) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}
func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}
func (m *MockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}
func (m *MockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	args := m.Called(serviceID, profileID)
	return args.Bool(0)
}
func (m *MockToolManager) ToolMatchesProfile(t tool.Tool, profileID string) bool {
	args := m.Called(t, profileID)
	return args.Bool(0)
}
func (m *MockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}

// MockTool
type MockTool struct {
	toolDef *v1.Tool
}

func (m *MockTool) Tool() *v1.Tool {
	return m.toolDef
}
func (m *MockTool) MCPTool() *mcp.Tool {
	return nil
}
func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// MockStorage (minimal implementation for ListServices)
type MockStorage struct {
	storage.Storage
}

func (m *MockStorage) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return nil, nil // Not used if registry is present
}

func TestHandleServices_ToolCountAndError(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockToolManager := new(MockToolManager)
	mockStore := new(MockStorage)

	app := &Application{
		ServiceRegistry: mockRegistry,
		ToolManager:     mockToolManager,
	}

	// Service 1: ID="s1", Name="service-1"
	// Service 2: ID="s2", Name="service-2"
	s1 := &configv1.UpstreamServiceConfig{Id: proto.String("s1"), Name: proto.String("service-1")}
	s2 := &configv1.UpstreamServiceConfig{Id: proto.String("s2"), Name: proto.String("service-2")}

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{s1, s2}, nil)

	// Mock Errors
	mockRegistry.On("GetServiceError", "s1").Return("connection failed", true)
	mockRegistry.On("GetServiceError", "s2").Return("", false)
	// Also called with name if ID check passes (or fails) - simplified logic in api.go checks ID first then name.
	// We just need to satisfy expectations.

	// Mock Tools
	// s1 has 2 tools
	t1 := &MockTool{toolDef: &v1.Tool{ServiceId: proto.String("s1"), Name: proto.String("t1")}}
	t2 := &MockTool{toolDef: &v1.Tool{ServiceId: proto.String("s1"), Name: proto.String("t2")}}
	// s2 has 0 tools
	// s3 has 1 tool (orphaned or from another service not in list)
	t3 := &MockTool{toolDef: &v1.Tool{ServiceId: proto.String("s3"), Name: proto.String("t3")}}

	mockToolManager.On("ListTools").Return([]tool.Tool{t1, t2, t3})

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	// Verify Service 1
	var r1 map[string]any
	var r2 map[string]any
	if response[0]["id"] == "s1" {
		r1 = response[0]
		r2 = response[1]
	} else {
		r1 = response[1]
		r2 = response[0]
	}

	assert.Equal(t, "service-1", r1["name"])
	assert.Equal(t, "connection failed", r1["last_error"])
	assert.Equal(t, float64(2), r1["tool_count"]) // JSON numbers are float64

	// Verify Service 2
	assert.Equal(t, "service-2", r2["name"])
	assert.Nil(t, r2["last_error"])
	assert.Equal(t, float64(0), r2["tool_count"])
}
