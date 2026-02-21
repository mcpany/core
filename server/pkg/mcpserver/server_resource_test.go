// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"log/slog"
	"testing"

	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockResourceTool struct {
	name   string
	result map[string]any
}

func (m *mockResourceTool) Tool() *v1.Tool {
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

func (m *mockResourceTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return m.result, nil
}

func (m *mockResourceTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockResourceTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.Tool())
	return t
}

func TestServer_CallTool_ResourceResult(t *testing.T) {
	t.Log("Starting TestServer_CallTool_ResourceResult")
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)
	t.Log("Bus provider created")

	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	t.Log("Creating server...")
	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, nil, busProvider, false)
	require.NoError(t, err)
	t.Log("Server created")

	tm := server.ToolManager().(*tool.Manager)

	// 1. Valid Resource with Text
	resourceTextTool := &mockResourceTool{
		name: "resource-text-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":      "internal://test-resource",
						"text":     "some text content",
						"mimeType": "text/plain",
					},
				},
			},
			"isError": false,
		},
	}
	t.Log("Adding resourceTextTool")
	require.NoError(t, tm.AddTool(resourceTextTool))

	// 2. Valid Resource with Blob (Base64 String)
	blobData := []byte("hello blob")
	resourceBlobTool := &mockResourceTool{
		name: "resource-blob-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":      "internal://test-blob",
						"blob":     base64.StdEncoding.EncodeToString(blobData),
						"mimeType": "application/octet-stream",
					},
				},
			},
			"isError": false,
		},
	}
	require.NoError(t, tm.AddTool(resourceBlobTool))

	// 3. Valid Resource with Blob (Byte Slice) - testing direct []byte handling
	resourceByteTool := &mockResourceTool{
		name: "resource-byte-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":      "internal://test-byte",
						"blob":     blobData, // Direct []byte
						"mimeType": "application/octet-stream",
					},
				},
			},
			"isError": false,
		},
	}
	require.NoError(t, tm.AddTool(resourceByteTool))

	// 4. Invalid Resource (Missing URI)
	invalidURITool := &mockResourceTool{
		name: "invalid-uri-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						// "uri": "missing",
						"text": "some text",
					},
				},
			},
			"isError": false,
		},
	}
	require.NoError(t, tm.AddTool(invalidURITool))

	// 5. Invalid Resource (Bad Base64)
	invalidBase64Tool := &mockResourceTool{
		name: "invalid-base64-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":  "internal://test-bad-base64",
						"blob": "not-base64!!!",
					},
				},
			},
			"isError": false,
		},
	}
	require.NoError(t, tm.AddTool(invalidBase64Tool))

	// Connect Client
	t.Log("Connecting client...")
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()
	t.Log("Client connected")

	t.Run("Valid Resource Text", func(t *testing.T) {
		t.Log("Running Valid Resource Text subtest")
		sanitizedToolName, _ := util.SanitizeToolName("resource-text-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)

		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		require.NotNil(t, resContent.Resource)
		assert.Equal(t, "internal://test-resource", resContent.Resource.URI)
		assert.Equal(t, "some text content", resContent.Resource.Text)
		assert.Equal(t, "text/plain", resContent.Resource.MIMEType)
	})

	t.Run("Valid Resource Blob String", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("resource-blob-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)

		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		require.NotNil(t, resContent.Resource)
		assert.Equal(t, "internal://test-blob", resContent.Resource.URI)
		assert.Equal(t, blobData, resContent.Resource.Blob)
		assert.Equal(t, "application/octet-stream", resContent.Resource.MIMEType)
	})

	t.Run("Valid Resource Blob Bytes", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("resource-byte-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})
		require.NoError(t, err)

		require.Len(t, result.Content, 1)
		resContent, ok := result.Content[0].(*mcp.EmbeddedResource)
		require.True(t, ok)
		require.NotNil(t, resContent.Resource)
		assert.Equal(t, "internal://test-byte", resContent.Resource.URI)
		assert.Equal(t, blobData, resContent.Resource.Blob)
	})

	t.Run("Invalid Resource Missing URI", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("invalid-uri-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "some text")
		assert.Contains(t, textContent.Text, "resource")
	})

	t.Run("Invalid Resource Bad Base64", func(t *testing.T) {
		sanitizedToolName, _ := util.SanitizeToolName("invalid-base64-tool")
		toolID := "test-service" + "." + sanitizedToolName
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: toolID,
		})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "not-base64!!!")
	})
}

func TestServer_CallTool_Logging(t *testing.T) {
	// Reset logger to capture output
	logging.ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	logging.Init(slog.LevelInfo, &buf, "", "json")

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

	// Tool with sensitive blob data
	blobData := []byte("very-sensitive-secret-blob")
	resourceBlobTool := &mockResourceTool{
		name: "sensitive-blob-tool",
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "resource",
					"resource": map[string]any{
						"uri":      "internal://sensitive",
						"blob":     base64.StdEncoding.EncodeToString(blobData),
						"mimeType": "application/octet-stream",
					},
				},
			},
			"isError": false,
		},
	}
	require.NoError(t, tm.AddTool(resourceBlobTool))

	// Connect Client
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call tool
	sanitizedToolName, _ := util.SanitizeToolName("sensitive-blob-tool")
	toolID := "test-service" + "." + sanitizedToolName
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
	})
	require.NoError(t, err)

	// Verify logs
	logOutput := buf.String()
	t.Logf("Log Output: %s", logOutput)

	// Should NOT contain the secret data
	assert.NotContains(t, logOutput, "very-sensitive-secret-blob")
	// Should contain summary
	assert.Contains(t, logOutput, "Resource(uri=internal://sensitive)")
	assert.Contains(t, logOutput, "blob=") // blob size summary
}
