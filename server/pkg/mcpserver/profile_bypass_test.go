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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockBypassTool struct {
	tool *v1.Tool
}

func (m *mockBypassTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockBypassTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func (m *mockBypassTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return "success", nil
}

func (m *mockBypassTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestServer_CallTool_ProfileBypass_Repro(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Define service config with restricted profile
	restrictedServiceID := "restricted-service"
	restrictedProfileID := "admin-profile"

	serviceInfo := &tool.ServiceInfo{
		Name: restrictedServiceID,
		Config: configv1.UpstreamServiceConfig_builder{}.Build(),
	}
	tm.AddServiceInfo(restrictedServiceID, serviceInfo)

	// Add restricted tool
	restrictedTool := &mockBypassTool{
		tool: v1.Tool_builder{
			Name:      proto.String("restricted-tool"),
			ServiceId: proto.String(restrictedServiceID),
			Annotations: v1.ToolAnnotations_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(&structpb.Struct{}),
					},
				},
			}.Build(),
		}.Build(),
	}
	_ = tm.AddTool(restrictedTool)

	// Configure profiles to allow access
	profileDef := configv1.ProfileDefinition_builder{
		Name: proto.String(restrictedProfileID),
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			restrictedServiceID: configv1.ProfileServiceConfig_builder{
				Enabled: proto.Bool(true),
			}.Build(),
		},
	}.Build()
	tm.SetProfiles([]string{restrictedProfileID}, []*configv1.ProfileDefinition{profileDef})

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// 1. Verify Tool is NOT visible in ListTools for user with different profile
	userProfileID := "user-profile"
	userCtx := auth.ContextWithProfileID(ctx, userProfileID)

	execReq := &tool.ExecutionRequest{
		ToolName: restrictedServiceID + "." + "restricted-tool",
		ToolInputs: []byte("{}"),
	}

	// Call with unauthorized profile
	res, err := server.CallTool(userCtx, execReq)

	// Ideally this should fail.
	// If the bug exists, it will succeed (err == nil).
	if err == nil {
		t.Fatalf("Expected error when calling restricted tool with unauthorized profile, but got success: %v", res)
	} else {
		// If it failed, check if it's the right error?
		// Currently we expect it to NOT fail because the bug exists.
		// So if err != nil, the bug is NOT present (or other error).
	}

	// 2. Call with authorized profile
	adminCtx := auth.ContextWithProfileID(ctx, restrictedProfileID)
	res, err = server.CallTool(adminCtx, execReq)
	assert.NoError(t, err, "Should succeed with authorized profile")

	ctr, ok := res.(*mcp.CallToolResult)
	require.True(t, ok)
	require.Len(t, ctr.Content, 1)
	textContent, ok := ctr.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	// Expect quoted "success" because it's JSON encoded
	assert.Equal(t, "\"success\"", textContent.Text)
}
