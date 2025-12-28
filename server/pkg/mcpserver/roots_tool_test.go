// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootsTool_Availability(t *testing.T) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager)
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

	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	// This initializes the server, which should add the RootsTool.
	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Try to call the roots tool
	// Note: The tool name is "mcp:list_roots"
	// However, tool manager might prefix it with service ID if it had one.
	// But since it currently fails to add, we expect this call to fail or the tool to be missing.

	// First, let's list tools to see if it's there.
	listResult, err := clientSession.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)

	found := false
	for _, tool := range listResult.Tools {
		// The tool name will be prefixed with the service ID (system) and sanitized.
		// "system.mcp:list_roots" -> "system.mcp_list_roots" (sanitized)
		// Or similar. Let's just check if it contains "list_roots"
		if tool.Name == "system.mcp_list_roots" || tool.Name == "system.mcp:list_roots" {
			found = true
			break
		}
	}

	assert.True(t, found, "mcp:list_roots tool should be available (as system.mcp_list_roots or similar)")
}
