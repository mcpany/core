// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// Rely on shared mocks (mockTool, mockResourceManager, etc.) from other test files in mcpserver_test package.

func TestAuthBypass_NoProfile(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// Create a restricted service and tool
	restrictedService := &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			Name: proto.String("restricted-service"),
			Id:   proto.String("restricted-service"),
		}.Build(),
	}

	restrictedTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("restricted-tool"),
			ServiceId: proto.String("restricted-service"),
		}.Build(),
	}

	// Create Tool Manager and register tool
	tm := tool.NewManager(busProvider)
	tm.AddServiceInfo("restricted-service", restrictedService)
	tm.AddTool(restrictedTool)

	// Set up profiles: "admin-profile" allows "restricted-service"
	adminProfile := configv1.ProfileDefinition_builder{
		Name: proto.String("admin-profile"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"restricted-service": configv1.ProfileServiceConfig_builder{}.Build(),
		},
	}.Build()

	// We need to set profiles in ToolManager to enable filtering logic (IsServiceAllowed)
	tm.SetProfiles([]string{"admin-profile"}, []*configv1.ProfileDefinition{adminProfile})

	rm := &mockResourceManager{}
	pm := &mockPromptManager{}
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// Scenario 1: User with NO Profile ID and NO Admin role (e.g. regular user via root endpoint)
	// Expectation: Should NOT see restricted tool (Fix applied)

	// We simulate an authenticated user but without profile selection
	userCtx := auth.ContextWithUser(ctx, "regular-user")
	// userCtx does NOT have ProfileID
	// userCtx does NOT have Admin role

	res, err := server.ToolListFilteringMiddleware(next)(userCtx, consts.MethodToolsList, &mcp.ListToolsRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListToolsResult)
	require.True(t, ok)

	// Fix Verification:
	// Should be 0.
	assert.Len(t, lRes.Tools, 0, "Fix Verified: Restricted tool NOT visible without profile for regular user")

	// Scenario 2: Call Tool without Profile ID
	execReq := &tool.ExecutionRequest{
		ToolName: "restricted-service.restricted-tool",
	}
	// Execute CallTool
	result, err := server.CallTool(userCtx, execReq)
	// Expect Access Denied
	require.Error(t, err, "Fix Verified: Restricted tool execution failed without profile")
	assert.Contains(t, err.Error(), "access denied")
	require.Nil(t, result)
}

func TestAuthBypass_Admin_NoProfile(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	restrictedService := &tool.ServiceInfo{
		Config: configv1.UpstreamServiceConfig_builder{
			Name: proto.String("restricted-service"),
			Id:   proto.String("restricted-service"),
		}.Build(),
	}

	restrictedTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("restricted-tool"),
			ServiceId: proto.String("restricted-service"),
		}.Build(),
	}

	tm := tool.NewManager(busProvider)
	tm.AddServiceInfo("restricted-service", restrictedService)
	tm.AddTool(restrictedTool)

	// No profiles needed for admin check logic, but setting them to ensure environment matches
	adminProfile := configv1.ProfileDefinition_builder{
		Name: proto.String("admin-profile"),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"restricted-service": configv1.ProfileServiceConfig_builder{}.Build(),
		},
	}.Build()
	tm.SetProfiles([]string{"admin-profile"}, []*configv1.ProfileDefinition{adminProfile})

	rm := &mockResourceManager{}
	pm := &mockPromptManager{}
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// Scenario: Admin User with NO Profile ID
	// Expectation: Should see ALL tools

	adminCtx := auth.ContextWithUser(ctx, "admin-user")
	adminCtx = auth.ContextWithRoles(adminCtx, []string{"admin"})

	res, err := server.ToolListFilteringMiddleware(next)(adminCtx, consts.MethodToolsList, &mcp.ListToolsRequest{})
	require.NoError(t, err)
	lRes, ok := res.(*mcp.ListToolsResult)
	require.True(t, ok)

	// Should see the tool (and maybe builtin)
	assert.GreaterOrEqual(t, len(lRes.Tools), 1, "Admin should see tools")

	found := false
	for _, tool := range lRes.Tools {
		if tool.Name == "restricted-service.restricted-tool" {
			found = true
			break
		}
	}
	assert.True(t, found, "Restricted tool should be visible to Admin")

	// Call Tool
	execReq := &tool.ExecutionRequest{
		ToolName: "restricted-service.restricted-tool",
	}
	result, err := server.CallTool(adminCtx, execReq)
	require.NoError(t, err, "Admin should be able to call tool without profile")
	require.NotNil(t, result)
}
