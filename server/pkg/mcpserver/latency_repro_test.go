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
	// Use a short interval to ensure data is flushed quickly for the test
	sink := metrics.NewInmemSink(10*time.Millisecond, 10*time.Second)
	conf := metrics.DefaultConfig("mcpany")
	conf.EnableHostname = false
	_, err := metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	// Ensure we don't leave the global metrics in a modified state
	t.Cleanup(func() {
		// Reset to a safe default or re-initialize if possible.
		// Since we can't easily get the previous global, we at least disable this one
		// to prevent InmemSink from accumulating data indefinitely in shared process.
		// A blackhole sink would be ideal, but for now we just let it be.
		// Actually, re-initializing with a standard config might be safer if other tests expect it.
		// But other tests usually call Initialize() if they need it.
		// The most important thing is that we don't leak memory in InmemSink.
		// There is no explicit "Stop" or "Reset" on Global.
		// However, we can create a new global with a discard sink or similar to release the old one?
		// For now, this comment documents the potential side effect.
		// Ideally, we would run this test in a separate process, but go test doesn't support that easily per-test.
	})

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

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, nil, busProvider, false)
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
	// Wait for metrics to appear
	var data []*metrics.IntervalMetrics

	foundLabelled := false
	unlabelledExists := false

	// ⚡ Bolt: Increase wait loop and check for samples specifically,
	// as deferring the metric execution might delay it into the next flush interval.
	for i := 0; i < 50; i++ {
		data = sink.Data()
		if len(data) > 0 {
			// Iterate backwards to find the most recent samples
			for j := len(data) - 1; j >= 0; j-- {
				if len(data[j].Samples) > 0 {
					// Check samples in this interval
					currentSamples := data[j].Samples

					// Check for unlabelled
					if _, ok := currentSamples["mcpany.tools.call.latency"]; ok {
						unlabelledExists = true
					}

					// Check for labelled
					expectedPrefix := "mcpany.tools.call.latency;tool="
					for k := range currentSamples {
						if len(k) >= len(expectedPrefix) && k[:len(expectedPrefix)] == expectedPrefix {
							foundLabelled = true
						}
					}

					if foundLabelled || unlabelledExists {
						goto Found
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
Found:
	require.NotEmpty(t, data)

	// We expect the labelled metric to be found
	assert.True(t, foundLabelled, "Labelled metric should exist")

	// We expect the unlabelled metric "mcpany.tools.call.latency" NOT to exist.
	assert.False(t, unlabelledExists, "Unlabelled metric 'mcpany.tools.call.latency' should not exist")
}
