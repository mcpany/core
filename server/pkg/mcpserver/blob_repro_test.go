// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_CallTool_Blob_ByteSlice_Repro(t *testing.T) {
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

	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add test tool that returns a map with blob as []byte
	blobTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String("blob-tool"),
				ServiceId: proto.String("test-service"),
				Annotations: v1.ToolAnnotations_builder{
					InputSchema: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"type":       structpb.NewStringValue("object"),
							"properties": structpb.NewStructValue(&structpb.Struct{}),
						},
					},
				}.Build(),
			}.Build()
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{
				Name: "test-service.blob-tool",
			}
		},
		ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":      "test://blob",
							"mimeType": "application/octet-stream",
							"blob":     []byte("test-data"), // Sending []byte directly
						},
					},
				},
			}, nil
		},
	}
	_ = tm.AddTool(blobTool)

	// Create client-server connection
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call the tool
	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "test-service.blob-tool",
	})
	require.NoError(t, err)

	require.Len(t, res.Content, 1)
	embeddedRes, ok := res.Content[0].(*mcp.EmbeddedResource)
	require.True(t, ok, "Content should be EmbeddedResource")

	// This should pass if the bug is NOT present
	assert.Equal(t, []byte("test-data"), embeddedRes.Resource.Blob, "Blob data should match")
}
