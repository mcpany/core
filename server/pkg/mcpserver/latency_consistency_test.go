// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	bus_pb "github.com/mcpany/core/proto/bus"
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
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestMetricLatencyConsistency(t *testing.T) {
	// Initialize metrics with an in-memory sink
	sink := metrics.NewInmemSink(10*time.Millisecond, 1*time.Minute)
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

	upstreamWorker := worker.NewUpstreamWorker(busProvider, toolManager)
	upstreamWorker.Start(ctx)

	// Set a very short flush interval for testing
	conf.TimerGranularity = 10 * time.Millisecond
	_, err = metrics.NewGlobal(conf, sink)
	require.NoError(t, err)

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add success tool
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

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call success tool
	sanitizedSuccessName, _ := util.SanitizeToolName("success-tool")
	successID := "test-service" + "." + sanitizedSuccessName
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{Name: successID})
	require.NoError(t, err)

	// Check metrics
	// Wait for metrics flush (TimerInterval is 10ms, Sink 10ms)
	time.Sleep(500 * time.Millisecond)

	data := sink.Data()

	// Consolidate samples from all intervals
	allSamples := make(map[string]metrics.SampledValue)
	for _, interval := range data {
		for k, v := range interval.Samples {
			allSamples[k] = v
		}
	}

	// Debug print
	fmt.Println("Available samples (all intervals):")
	for k := range allSamples {
		fmt.Println(k)
	}

	// We expect consistent naming with counters: mcpany.tools.call.latency (plural)
	// and properly labeled for tool specific metrics.

	// Check global latency metric
	assert.NotContains(t, allSamples, "mcpany.tools.call.latency")

	// Check tool-specific latency metric
	// Ideally it should be mcpany.tools.call.latency;tool=...
	// But currently it is mcpany.tool.<toolname>.call.latency (which causes high cardinality)

	// We expect tool AND service_id labels
	expectedIdealKey := "mcpany.tools.call.latency;tool=" + successID + ";service_id=test-service"

	// This assertion verifies the presence of the labeled metric
	assert.Contains(t, allSamples, expectedIdealKey, "Should have labeled latency metric")

	// Ensure we don't have the old format with embedded tool name
	currentBadKey := "mcpany.tool." + successID + ".call.latency"
	assert.NotContains(t, allSamples, currentBadKey, "Should NOT have embedded tool name in metric key")
}
