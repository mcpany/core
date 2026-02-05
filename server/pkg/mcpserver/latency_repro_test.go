// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_CallTool_Latency_Metrics_Repro(t *testing.T) {
	// Initialize metrics with an in-memory sink
	sink := metrics.NewInmemSink(10*time.Second, 10*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

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

	// Start the worker to handle tool execution
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

	// Call the tool
	sanitizedToolName, _ := util.SanitizeToolName("success-tool")
	toolID := "test-service" + "." + sanitizedToolName
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: toolID,
	})
	require.NoError(t, err)

	// Check metrics
	// Use Eventually to wait for metrics to be flushed
	require.Eventually(t, func() bool {
		data := sink.Data()
		if len(data) == 0 {
			return false
		}
		samples := data[0].Samples
		if len(samples) == 0 {
			return false
		}

		// We expect the unlabelled metric "mcpany.tools.call.latency" NOT to exist.
		// But in the buggy version, it DOES exist.

		// Check for unlabelled metric
		_, unlabelledExists := samples["mcpany.tools.call.latency"]

		if unlabelledExists {
			// If we are testing for the absence, finding it means we failed the specific check
			// but Eventually will retry until timeout.
			// However, since we want to assert it DOES NOT exist, finding it is actually "success" for the metric collection
			// but failure for the test logic if we were asserting success.
			// Wait, the test logic is: assert.False(unlabelledExists).
			// So if unlabelledExists is true, the test fails.
			// But here we are just waiting for *any* metrics to appear to avoid the "empty map" error.
			// So we should check for the presence of the *expected* metric to confirm data arrival.
		}

		// Check for labelled metric (should exist)
		foundLabelled := false
		expectedPrefix := "mcpany.tools.call.latency;tool="
		for k := range samples {
			if len(k) >= len(expectedPrefix) && k[:len(expectedPrefix)] == expectedPrefix {
				foundLabelled = true
				break
			}
		}
		return foundLabelled
	}, 5*time.Second, 100*time.Millisecond, "Metrics should be recorded")

	data := sink.Data()
	require.NotEmpty(t, data)
	samples := data[0].Samples

	// Check for unlabelled metric
	_, unlabelledExists := samples["mcpany.tools.call.latency"]
	assert.False(t, unlabelledExists, "Unlabelled metric 'mcpany.tools.call.latency' should not exist")
}
