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
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type mockMapTool struct {
	tool   *v1.Tool
	result map[string]any
}

func (m *mockMapTool) Tool() *v1.Tool {
	return m.tool
}

func (m *mockMapTool) Execute(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return m.result, nil
}

func (m *mockMapTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (m *mockMapTool) MCPTool() *mcp.Tool {
	t, _ := tool.ConvertProtoToMCPTool(m.tool)
	return t
}

func BenchmarkCallTool_MapResult(b *testing.B) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(b, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(b, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add test tool returning a map
	mapTool := &mockMapTool{
		tool: v1.Tool_builder{
			Name:      proto.String("map-tool"),
			ServiceId: proto.String("test-service"),
			Annotations: v1.ToolAnnotations_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(&structpb.Struct{}),
					},
				},
			}.Build(),
		}.Build(),
		result: map[string]any{
			"content": []any{
				map[string]any{
					"type": "text",
					"text": "This is a benchmark result text that is long enough to matter.",
				},
			},
			"isError": false,
		},
	}
	_ = tm.AddTool(mapTool)

	req := &tool.ExecutionRequest{
		ToolName:   "test-service.map-tool",
		ToolInputs: []byte("{}"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.CallTool(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCallTool_PlainMapResult(b *testing.B) {
	poolManager := pool.NewManager()
	factory := factory.NewUpstreamServiceFactory(poolManager, nil)
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(b, err)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(factory, toolManager, promptManager, resourceManager, authManager)
	ctx := context.Background()

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(b, err)

	tm := server.ToolManager().(*tool.Manager)

	// Create a large result map to make allocation visible
	largeResult := make(map[string]any)
	largeData := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeData[i] = "some data string that takes up space and makes json marshal work " + string(rune(i))
	}
	largeResult["data"] = largeData
	largeResult["other"] = "fields"

	// Add test tool returning a plain map (no content/isError fields)
	mapTool := &mockMapTool{
		tool: v1.Tool_builder{
			Name:      proto.String("plain-map-tool"),
			ServiceId: proto.String("test-service"),
			Annotations: v1.ToolAnnotations_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type":       structpb.NewStringValue("object"),
						"properties": structpb.NewStructValue(&structpb.Struct{}),
					},
				},
			}.Build(),
		}.Build(),
		result: largeResult,
	}
	_ = tm.AddTool(mapTool)

	req := &tool.ExecutionRequest{
		ToolName:   "test-service.plain-map-tool",
		ToolInputs: []byte("{}"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := server.CallTool(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
