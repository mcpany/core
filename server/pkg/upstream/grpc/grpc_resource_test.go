package grpc

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockResourceManager is a mock of resource.ManagerInterface.
type MockResourceManager struct {
	mock.Mock
}

func (m *MockResourceManager) GetResource(uri string) (resource.Resource, bool) {
	args := m.Called(uri)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(resource.Resource), args.Bool(1)
}

func (m *MockResourceManager) AddResource(r resource.Resource) {
	m.Called(r)
}

func (m *MockResourceManager) RemoveResource(uri string) {
	m.Called(uri)
}

func (m *MockResourceManager) ListResources() []resource.Resource {
	args := m.Called()
	return args.Get(0).([]resource.Resource)
}

func (m *MockResourceManager) OnListChanged(f func()) {
	m.Called(f)
}

func (m *MockResourceManager) ClearResourcesForService(serviceID string) {
	m.Called(serviceID)
}

// MockTool needs to implement tool.Tool
type MockTool struct {
	mock.Mock
}

func (m *MockTool) Tool() *v1.Tool {
	return nil
}

func (m *MockTool) MCPTool() *mcp.Tool {
	return nil
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

func (m *MockTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// Custom MockToolManager for gRPC tests to use testify/mock
type TestMockToolManager struct {
	mock.Mock
}

func (m *TestMockToolManager) AddTool(tool tool.Tool) error {
	args := m.Called(tool)
	return args.Error(0)
}

func (m *TestMockToolManager) GetTool(toolID string) (tool.Tool, bool) {
	args := m.Called(toolID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

func (m *TestMockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *TestMockToolManager) ListTools() []tool.Tool {
	return nil
}

func (m *TestMockToolManager) ListMCPTools() []*mcp.Tool {
	return nil
}

func (m *TestMockToolManager) ClearToolsForService(serviceID string) {
	m.Called(serviceID)
}

func (m *TestMockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}

func (m *TestMockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *TestMockToolManager) SetMCPServer(provider tool.MCPServerProvider) {
	m.Called(provider)
}

func (m *TestMockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *TestMockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {
}

func (m *TestMockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}

func (m *TestMockToolManager) SetProfiles(enabled []string, defs []*configv1.ProfileDefinition) {
	m.Called(enabled, defs)
}

func (m *TestMockToolManager) IsServiceAllowed(serviceID, profileID string) bool {
	return true
}

func (m *TestMockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	return nil, true
}

func (m *TestMockToolManager) GetToolCountForService(serviceID string) int {
	return 0
}

func TestRegisterDynamicResources_Detailed(t *testing.T) {
	u := &Upstream{}
	serviceID := "test-service"

	t.Run("success", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("myTool"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("myResource"),
					Dynamic: configv1.DynamicResource_builder{
						GrpcCall: configv1.GrpcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		mockToolManager.On("GetTool", "test-service.myTool").Return(new(MockTool), true)
		mockResourceManager.On("AddResource", mock.Anything).Return()

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockToolManager.AssertExpectations(t)
		mockResourceManager.AssertExpectations(t)
	})

	t.Run("disabled resource", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("myTool"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name:    proto.String("myResource"),
					Disable: proto.Bool(true),
					Dynamic: configv1.DynamicResource_builder{
						GrpcCall: configv1.GrpcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("tool not found for call ID", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				// No tool matching call ID
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("myResource"),
					Dynamic: configv1.DynamicResource_builder{
						GrpcCall: configv1.GrpcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("tool not found in manager", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("myTool"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("myResource"),
					Dynamic: configv1.DynamicResource_builder{
						GrpcCall: configv1.GrpcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		mockToolManager.On("GetTool", "test-service.myTool").Return(nil, false)

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockToolManager.AssertExpectations(t)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("invalid dynamic definition (missing grpc call)", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("myResource"),
					Dynamic: configv1.DynamicResource_builder{
						// Missing call definition
					}.Build(),
				}.Build(),
			},
		}.Build()

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("new dynamic resource failure", func(t *testing.T) {
		mockToolManager := new(TestMockToolManager)
		mockResourceManager := new(MockResourceManager)

		grpcService := configv1.GrpcUpstreamService_builder{
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("myTool"),
					CallId: proto.String("call1"),
				}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{
					Name: proto.String("myResource"),
					Dynamic: configv1.DynamicResource_builder{
						GrpcCall: configv1.GrpcCallDefinition_builder{
							Id: proto.String("call1"),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()

		// Return nil tool but true to trigger error in NewDynamicResource
		mockToolManager.On("GetTool", "test-service.myTool").Return(nil, true)

		u.registerDynamicResources(serviceID, grpcService, mockResourceManager, mockToolManager)

		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})
}
