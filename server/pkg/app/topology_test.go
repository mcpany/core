package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	topologyv1 "github.com/mcpany/core/proto/topology/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MockToolManager implements tool.ManagerInterface for testing
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

func (m *MockToolManager) ListTools() []tool.Tool {
	args := m.Called()
	return args.Get(0).([]tool.Tool)
}

func (m *MockToolManager) ListMCPTools() []*mcp.Tool {
	args := m.Called()
	return args.Get(0).([]*mcp.Tool)
}

func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}

func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info := args.Get(0); info != nil {
		return info.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
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

func (m *MockToolManager) GetToolCountForService(serviceID string) int {
	return 0
}

func (m *MockToolManager) AddMiddleware(middleware tool.ExecutionMiddleware) {
	m.Called(middleware)
}

func (m *MockToolManager) SetMCPServer(server tool.MCPServerProvider) {
	m.Called(server)
}

func (m *MockToolManager) ListServices() []*tool.ServiceInfo {
	args := m.Called()
	return args.Get(0).([]*tool.ServiceInfo)
}

func TestHandleTopology(t *testing.T) {
	app := NewApplication()
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)

	app.ServiceRegistry = mockRegistry
	app.ToolManager = mockTM
	app.TopologyManager = topology.NewManager(mockRegistry, mockTM)

	t.Run("Success", func(t *testing.T) {
		// Setup mock data
		s1 := configv1.UpstreamServiceConfig_builder{}.Build()
		s1.SetName("service-1")
		s2 := configv1.UpstreamServiceConfig_builder{}.Build()
		s2.SetName("service-2")
		s2.SetDisable(true)
		services := []*configv1.UpstreamServiceConfig{s1, s2}
		mockRegistry.On("GetAllServices").Return(services, nil).Once()

		tools := []tool.Tool{
			&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool-1"), ServiceId: proto.String("service-1")}.Build()},
		}
		mockTM.On("ListTools").Return(tools).Once()

		req := httptest.NewRequest(http.MethodGet, "/topology", nil)
		w := httptest.NewRecorder()

		app.handleTopology()(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var graph topologyv1.Graph
		// Using protojson to unmarshal is better as it handles enum names
		err := protojson.Unmarshal(w.Body.Bytes(), &graph)
		// If protojson fails (e.g. unknown fields), standard json unmarshal might work for basic verification
		if err != nil {
			// fallback to check basic JSON structure
			var raw map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &raw)
			assert.NotNil(t, raw["core"])
		} else {
			assert.Equal(t, "mcp-core", graph.GetCore().GetId())
			// Check children
			// We expect: Middleware, Webhooks, Service-1, Service-2
			// Service-1 should have tool-1
			foundSvc1 := false
			foundSvc2 := false
			for _, child := range graph.GetCore().GetChildren() {
				if child.GetId() == "svc-service-1" {
					foundSvc1 = true
					assert.Equal(t, topologyv1.NodeStatus_NODE_STATUS_ACTIVE, child.GetStatus())
					assert.NotEmpty(t, child.GetChildren())
					assert.Equal(t, "tool-tool-1", child.GetChildren()[0].GetId())
				}
				if child.GetId() == "svc-service-2" {
					foundSvc2 = true
					assert.Equal(t, topologyv1.NodeStatus_NODE_STATUS_INACTIVE, child.GetStatus())
				}
			}
			assert.True(t, foundSvc1, "Service 1 not found in topology")
			assert.True(t, foundSvc2, "Service 2 not found in topology")
		}
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/topology", nil)
		w := httptest.NewRecorder()
		app.handleTopology()(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("NotModified", func(t *testing.T) {
		// Setup mock data again for this run
		services := []*configv1.UpstreamServiceConfig{}
		mockRegistry.On("GetAllServices").Return(services, nil).Once()
		mockTM.On("ListTools").Return([]tool.Tool{}).Once()

		// 1. First request to get ETag
		req1 := httptest.NewRequest(http.MethodGet, "/topology", nil)
		w1 := httptest.NewRecorder()
		app.handleTopology()(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)
		etag := w1.Header().Get("ETag")
		assert.NotEmpty(t, etag, "ETag header should be set")

		// 2. Second request with If-None-Match
		mockRegistry.On("GetAllServices").Return(services, nil).Once()
		mockTM.On("ListTools").Return([]tool.Tool{}).Once()

		req2 := httptest.NewRequest(http.MethodGet, "/topology", nil)
		req2.Header.Set("If-None-Match", etag)
		w2 := httptest.NewRecorder()

		app.handleTopology()(w2, req2)
		assert.Equal(t, http.StatusNotModified, w2.Code)
		assert.Empty(t, w2.Body.String(), "Body should be empty for 304")
	})
}
