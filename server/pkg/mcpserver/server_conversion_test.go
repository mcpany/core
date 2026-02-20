// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/base64"
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

type mockConversionTool struct {
	name   string
	result map[string]any
}

func (m *mockConversionTool) Tool() *v1.Tool {
	return v1.Tool_builder{
		Name:      proto.String(m.name),
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
}

func (m *mockConversionTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return m.result, nil
}

func (m *mockConversionTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockConversionTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.Tool())
	return t
}

func TestServer_CallTool_ResultConversion(t *testing.T) {
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

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// 1. Resource with Text content
	resTextTool := &mockConversionTool{
		name: "res-text-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":      "file:///text",
						"mimeType": "text/plain",
						"text":     "some text content",
					},
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(resTextTool)

	// 2. Resource with Blob content (Base64 string)
	blobData := []byte("binary-data")
	resBlobStrTool := &mockConversionTool{
		name: "res-blob-str-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":  "file:///blob-str",
						"blob": base64.StdEncoding.EncodeToString(blobData),
					},
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(resBlobStrTool)

	// 3. Resource with Blob content ([]byte)
	resBlobBytesTool := &mockConversionTool{
		name: "res-blob-bytes-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":  "file:///blob-bytes",
						"blob": blobData,
					},
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(resBlobBytesTool)

	// 4. Mixed Content
	mixedTool := &mockConversionTool{
		name: "mixed-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "text",
					"text": "text part",
				},
				map[string]any{
					"type":     "image",
					"data":     base64.StdEncoding.EncodeToString(blobData),
					"mimeType": "image/png",
				},
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":  "file:///mixed",
						"text": "resource part",
					},
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(mixedTool)

	// 5. Invalid Resource (Missing URI) -> Fallback to Raw Text
	invalidURITool := &mockConversionTool{
		name: "invalid-uri-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						// "uri": missing
						"text": "invalid",
					},
				},
			},
		},
	}
	_ = tm.AddTool(invalidURITool)

	// 6. Invalid Base64 -> Fallback to Raw Text
	invalidBase64Tool := &mockConversionTool{
		name: "invalid-base64-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":  "file:///invalid-base64",
						"blob": "not-base64!",
					},
				},
			},
		},
	}
	_ = tm.AddTool(invalidBase64Tool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	t.Run("Resource Text", func(t *testing.T) {
		toolID := "test-service.res-text-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		assert.Equal(t, "file:///text", resContent.Resource.URI)
		assert.Equal(t, "text/plain", resContent.Resource.MIMEType)
		assert.Equal(t, "some text content", resContent.Resource.Text)
	})

	t.Run("Resource Blob String", func(t *testing.T) {
		toolID := "test-service.res-blob-str-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		assert.Equal(t, "file:///blob-str", resContent.Resource.URI)
		assert.Equal(t, blobData, resContent.Resource.Blob)
	})

	t.Run("Resource Blob Bytes", func(t *testing.T) {
		toolID := "test-service.res-blob-bytes-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		assert.Equal(t, "file:///blob-bytes", resContent.Resource.URI)
		assert.Equal(t, blobData, resContent.Resource.Blob)
	})

	t.Run("Mixed Content", func(t *testing.T) {
		toolID := "test-service.mixed-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 3)

		text, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Equal(t, "text part", text.Text)

		img, ok := result.Content[1].(*mcp.ImageContent)
		require.True(t, ok)
		assert.Equal(t, blobData, img.Data)

		res, ok := result.Content[2].(*mcp.EmbeddedResource)
		require.True(t, ok)
		assert.Equal(t, "file:///mixed", res.Resource.URI)
	})

	t.Run("Invalid URI Fallback", func(t *testing.T) {
		toolID := "test-service.invalid-uri-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		// Should fall back to raw text (JSON representation)
		text, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, text.Text, "\"text\":\"invalid\"")
	})

	t.Run("Invalid Base64 Fallback", func(t *testing.T) {
		toolID := "test-service.invalid-base64-tool"
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		// Should fall back to raw text
		text, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, text.Text, "\"blob\":\"not-base64!\"")
	})
}
