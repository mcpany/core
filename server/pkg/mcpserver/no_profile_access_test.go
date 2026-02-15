// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/consts"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestAuthBypass_ListTools_NoProfile(t *testing.T) {
	// Setup dependencies
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	tm := tool.NewManager(busProvider)

	// Register a restricted tool
	toolDef := pb.Tool_builder{
		Name:      proto.String("restricted-tool"),
		ServiceId: proto.String("secure-service"),
	}.Build()

	// Create a dummy tool implementation
	restrictedTool := tool.NewLocalCommandTool(toolDef,
		configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}.Build(),
		configv1.CommandLineCallDefinition_builder{}.Build(),
		nil, "call-id")

	err = tm.AddTool(restrictedTool)
	require.NoError(t, err)

	tm.AddServiceInfo("secure-service", &tool.ServiceInfo{
		Name: "secure-service",
		Config: configv1.UpstreamServiceConfig_builder{Name: proto.String("secure-service")}.Build(),
	})

	rm := resource.NewManager()
	pm := prompt.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{}, nil
	}

	// Case 1: Attacker (User role) accesses ListTools without profile
	ctxAttacker := auth.ContextWithUser(ctx, "attacker")
	ctxAttacker = auth.ContextWithRoles(ctxAttacker, []string{"user"})
	// ProfileID is NOT set

	res, err := server.ToolListFilteringMiddleware(next)(ctxAttacker, consts.MethodToolsList, nil)
	require.NoError(t, err)

	lRes, ok := res.(*mcp.ListToolsResult)
	require.True(t, ok)

	// EXPECTED BEHAVIOR (Post-Fix): Restricted tools should NOT be visible.
	// Only built-in tools (if any) or explicitly public tools.
	// In this test environment, "restricted-tool" (namespaced) should be hidden.

	foundRestricted := false
	for _, tool := range lRes.Tools {
		if strings.Contains(tool.Name, "restricted-tool") {
			foundRestricted = true
		}
	}

	if foundRestricted {
		t.Fatal("VULNERABILITY: Restricted tool is visible without profile!")
	}
}

func TestAuthBypass_CallTool_NoProfile(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	tm := tool.NewManager(busProvider)

	toolDef := pb.Tool_builder{
		Name:      proto.String("restricted-tool"),
		ServiceId: proto.String("secure-service"),
	}.Build()

	restrictedTool := tool.NewLocalCommandTool(toolDef,
		configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}.Build(),
		configv1.CommandLineCallDefinition_builder{}.Build(),
		nil, "call-id")

	err := tm.AddTool(restrictedTool)
	require.NoError(t, err)

	tm.AddServiceInfo("secure-service", &tool.ServiceInfo{
		Name: "secure-service",
		Config: configv1.UpstreamServiceConfig_builder{Name: proto.String("secure-service")}.Build(),
	})

	rm := resource.NewManager()
	pm := prompt.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory.NewUpstreamServiceFactory(pool.NewManager(), nil), tm, pm, rm, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, tm, pm, rm, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)

	// Case 1: Attacker calls tool without profile
	ctxAttacker := auth.ContextWithUser(ctx, "attacker")
	ctxAttacker = auth.ContextWithRoles(ctxAttacker, []string{"user"})

	// Use namespaced name
	req := &tool.ExecutionRequest{
		ToolName: "secure-service.restricted-tool",
		ToolInputs: []byte("{}"),
		DryRun: true,
	}

	// EXPECTED BEHAVIOR (Post-Fix): Tool execution should be DENIED.
	_, err = server.CallTool(ctxAttacker, req)

	if err == nil {
		t.Fatal("VULNERABILITY: Restricted tool executed without profile!")
	}

	assert.Contains(t, err.Error(), "access denied", "Error should indicate access denied")
}
