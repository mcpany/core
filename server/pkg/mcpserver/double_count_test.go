package mcpserver_test

import (
	"context"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// This test verifies that we do NOT double count metrics by emitting both unlabelled and labelled versions
// of the same metric name. This is crucial for Prometheus compatibility where they would clash.
func TestServer_NoDoubleCounting(t *testing.T) {
	// Initialize metrics with an in-memory sink
	sink := metrics.NewInmemSink(10*time.Second, 30*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	// Standard server setup
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
	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add test tool
	successTool := &mockTool{
		tool: v1.Tool_builder{
			Name:      proto.String("success-tool"),
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
	}
	_ = tm.AddTool(successTool)

	// Connect
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call tool
	sanitizedToolName, _ := util.SanitizeToolName("success-tool")
	toolID := "test-service" + "." + sanitizedToolName
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{Name: toolID})
	require.NoError(t, err)

	// Verify metrics
	data := sink.Data()
	require.NotEmpty(t, data)
	counters := data[0].Counters

	// Unlabelled metric (mcpany.tools.call.total) should NOT exist
	// Currently it DOES exist, so this assertion will FAIL, proving the bug.
	_, unlabelledExists := counters["mcpany.tools.call.total"]
	assert.False(t, unlabelledExists, "Unlabelled metric 'mcpany.tools.call.total' should not exist to avoid double counting")

	// Labelled metric SHOULD exist
	// Since we don't know the exact order of labels in key, we search for prefix
	foundLabelled := false
	expectedPrefix := "mcpany.tools.call.total;tool="
	for k := range counters {
		if len(k) >= len(expectedPrefix) && k[:len(expectedPrefix)] == expectedPrefix {
			foundLabelled = true
			break
		}
	}
	// Fallback check for service_id first
	if !foundLabelled {
		expectedPrefix2 := "mcpany.tools.call.total;service_id="
		for k := range counters {
			if len(k) >= len(expectedPrefix2) && k[:len(expectedPrefix2)] == expectedPrefix2 {
				foundLabelled = true
				break
			}
		}
	}
	assert.True(t, foundLabelled, "Labelled metric should exist")
}
