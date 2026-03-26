// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

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

// TestRegisterDynamicResources ...
// Summary: TestRegisterDynamicResources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	u := NewOpenAPIUpstream().(*OpenAPIUpstream)
	serviceID := "test-service"

	t.Run("success", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		definitions := []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("myTool"),
				CallId: proto.String("call1"),
			}.Build(),
		}

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name: proto.String("myResource"),
				Dynamic: configv1.DynamicResource_builder{
					HttpCall: configv1.HttpCallDefinition_builder{
						Id: proto.String("call1"),
					}.Build(),
				}.Build(),
			}.Build(),
		}

		mockToolManager.On("GetTool", "test-service.myTool").Return(new(MockTool), true)
		mockResourceManager.On("AddResource", mock.Anything).Return()

		u.registerDynamicResources(serviceID, definitions, resources, mockResourceManager, mockToolManager)

		mockToolManager.AssertExpectations(t)
		mockResourceManager.AssertExpectations(t)
	})

	t.Run("disabled resource", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		definitions := []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("myTool"),
				CallId: proto.String("call1"),
			}.Build(),
		}

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name:    proto.String("myResource"),
				Disable: proto.Bool(true),
				Dynamic: configv1.DynamicResource_builder{
					HttpCall: configv1.HttpCallDefinition_builder{
						Id: proto.String("call1"),
					}.Build(),
				}.Build(),
			}.Build(),
		}

		u.registerDynamicResources(serviceID, definitions, resources, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("tool not found for call ID", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		definitions := []*configv1.ToolDefinition{
			// No tools matching call ID
		}

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name: proto.String("myResource"),
				Dynamic: configv1.DynamicResource_builder{
					HttpCall: configv1.HttpCallDefinition_builder{
						Id: proto.String("call1"),
					}.Build(),
				}.Build(),
			}.Build(),
		}

		u.registerDynamicResources(serviceID, definitions, resources, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("tool not found in manager", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		definitions := []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("myTool"),
				CallId: proto.String("call1"),
			}.Build(),
		}

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name: proto.String("myResource"),
				Dynamic: configv1.DynamicResource_builder{
					HttpCall: configv1.HttpCallDefinition_builder{
						Id: proto.String("call1"),
					}.Build(),
				}.Build(),
			}.Build(),
		}

		mockToolManager.On("GetTool", "test-service.myTool").Return(nil, false)

		u.registerDynamicResources(serviceID, definitions, resources, mockResourceManager, mockToolManager)

		mockToolManager.AssertExpectations(t)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("invalid dynamic definition (missing http call)", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name:    proto.String("myResource"),
				Dynamic: configv1.DynamicResource_builder{
					// Missing CallDefinition
				}.Build(),
			}.Build(),
		}

		u.registerDynamicResources(serviceID, nil, resources, mockResourceManager, mockToolManager)

		mockToolManager.AssertNotCalled(t, "GetTool", mock.Anything)
		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})

	t.Run("new dynamic resource failure", func(t *testing.T) {
		mockToolManager := new(MockToolManager)
		mockResourceManager := new(MockResourceManager)

		definitions := []*configv1.ToolDefinition{
			configv1.ToolDefinition_builder{
				Name:   proto.String("myTool"),
				CallId: proto.String("call1"),
			}.Build(),
		}

		resources := []*configv1.ResourceDefinition{
			configv1.ResourceDefinition_builder{
				Name: proto.String("myResource"),
				Dynamic: configv1.DynamicResource_builder{
					HttpCall: configv1.HttpCallDefinition_builder{
						Id: proto.String("call1"),
					}.Build(),
				}.Build(),
			}.Build(),
		}

		// Return nil tool but true to trigger error in NewDynamicResource
		mockToolManager.On("GetTool", "test-service.myTool").Return(nil, true)

		u.registerDynamicResources(serviceID, definitions, resources, mockResourceManager, mockToolManager)

		mockResourceManager.AssertNotCalled(t, "AddResource", mock.Anything)
	})
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
