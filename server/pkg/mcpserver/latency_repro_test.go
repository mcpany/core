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
	for i := 0; i < 20; i++ {
		data = sink.Data()
		// Since execution takes ~50ms, metrics might span multiple intervals.
		// We need to wait until we see some data.
		if len(data) > 0 {
			// Check if we have the latency metric in any interval
			found := false
			for _, interval := range data {
				for k := range interval.Samples {
					if len(k) >= len("mcpany.tools.call.latency") && k[:len("mcpany.tools.call.latency")] == "mcpany.tools.call.latency" {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NotEmpty(t, data)

	unlabelledExists := false
	foundLabelled := false
	expectedPrefix := "mcpany.tools.call.latency;tool="

	for _, interval := range data {
		samples := interval.Samples
		if samples == nil {
			continue
		}

		// Check for unlabelled metric
		if _, exists := samples["mcpany.tools.call.latency"]; exists {
			unlabelledExists = true
		}

		// Check for labelled metric
		for k := range samples {
			if len(k) >= len(expectedPrefix) && k[:len(expectedPrefix)] == expectedPrefix {
				foundLabelled = true
			}
		}
	}

	// This assertion should FAIL currently, demonstrating the bug.
	assert.False(t, unlabelledExists, "Unlabelled metric 'mcpany.tools.call.latency' should not exist")
	assert.True(t, foundLabelled, "Labelled metric should exist")
}
