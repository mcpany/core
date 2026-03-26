// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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
// MockResourceManager is a mock of resource.ManagerInterface.
// Summary: MockResourceManager
	mock.Mock
}

// GetResource ...
// Summary: GetResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(uri)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(resource.Resource), args.Bool(1)
}

// AddResource ...
// Summary: AddResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(r)
}

// RemoveResource ...
// Summary: RemoveResource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(uri)
}

// ListResources ...
// Summary: ListResources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]resource.Resource)
}

// OnListChanged ...
// Summary: OnListChanged
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(f)
}

// ClearResourcesForService ...
// Summary: ClearResourcesForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID)
}

// MockTool needs to implement tool.Tool
// MockTool needs to implement tool.Tool
// Summary: MockTool
	mock.Mock
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// GetCacheConfig ...
// Summary: GetCacheConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// Custom MockToolManager for gRPC tests to use testify/mock
// Custom MockToolManager for gRPC tests to use testify/mock
// Summary: TestMockToolManager
	mock.Mock
}

// AddTool ...
// Summary: AddTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(tool)
	return args.Error(0)
}

// GetTool ...
// Summary: GetTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(toolID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(tool.Tool), args.Bool(1)
}

// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ListTools ...
// Summary: ListTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ListMCPTools ...
// Summary: ListMCPTools
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// ClearToolsForService ...
// Summary: ClearToolsForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID)
}

// AddServiceInfo ...
// Summary: AddServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(serviceID, info)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, false
}

// SetMCPServer ...
// Summary: SetMCPServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(provider)
}

// ExecuteTool ...
// Summary: ExecuteTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// AddMiddleware ...
// Summary: AddMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
}

// ToolMatchesProfile ...
// Summary: ToolMatchesProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// SetProfiles ...
// Summary: SetProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.Called(enabled, defs)
}

// IsServiceAllowed ...
// Summary: IsServiceAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return true
}

// GetAllowedServiceIDs ...
// Summary: GetAllowedServiceIDs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, true
}

// GetToolCountForService ...
// Summary: GetToolCountForService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0
}

// TestRegisterDynamicResources_Detailed ...
// Summary: TestRegisterDynamicResources_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
					Name:    proto.String("myResource"),
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
