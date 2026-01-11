// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"fmt"
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

func TestMetricNamingConsistency(t *testing.T) {
	// Initialize metrics with an in-memory sink
	sink := metrics.NewInmemSink(10*time.Second, 30*time.Second)
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

	server, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
	require.NoError(t, err)

	tm := server.ToolManager().(*tool.Manager)

	// Add success tool
	successTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
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
			}.Build()
		},
		MCPToolFunc: func() *mcp.Tool {
			return &mcp.Tool{
				Name: "test-service.success-tool",
			}
		},
		ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return "success", nil
		},
	}
	_ = tm.AddTool(successTool)

	// Add error tool
	errorTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				Name:      proto.String("error-tool"),
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
				Name: "test-service.error-tool",
			}
		},
		ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			return nil, fmt.Errorf("execution error")
		},
	}
	_ = tm.AddTool(errorTool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Server().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// 1. Call success tool
	sanitizedSuccessName, _ := util.SanitizeToolName("success-tool")
	successID := "test-service" + "." + sanitizedSuccessName
	_, err = clientSession.CallTool(ctx, &mcp.CallToolParams{Name: successID})
	require.NoError(t, err)

	// 2. Call error tool
	sanitizedErrorName, _ := util.SanitizeToolName("error-tool")
	errorID := "test-service" + "." + sanitizedErrorName
	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: errorID})
	require.NoError(t, err)
	assert.True(t, res.IsError, "Expected IsError=true")

	// Check metrics
	data := sink.Data()
	// Flatten counters from all intervals (should be 1 interval mostly)
	counters := make(map[string]int)
	for _, interval := range data {
		for k, v := range interval.Counters {
			counters[k] += v.Count
		}
	}

	// We expect:
	// mcpany.tools.call.total = 2 (1 success + 1 error)
	// mcpany.tools.call.errors = 1 (1 error)

	// Labeled metrics:
	// mcpany.tools.call.total;tool=... = 1 each
	// mcpany.tools.call.errors;tool=... = 1 for error tool

	// Note: Current code uses "tool" (singular), so we expect failure if we look for "tools" (plural) labels.
	// But we look for "tools" because that's what we want.

	// Helper to find key with prefix
	findKey := func(prefix string) string {
		for k := range counters {
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
				return k
			}
		}
		return ""
	}

	// Print all keys for debugging
	fmt.Println("Available counters:")
	for k := range counters {
		fmt.Println(k)
	}

	// Verify Global Counters
	// Note: We removed the unlabelled counters to avoid double counting in Prometheus.
	// We only verify labelled counters below.
	assert.NotContains(t, counters, "mcpany.tools.call.total")
	assert.NotContains(t, counters, "mcpany.tools.call.errors")

	// Verify Labeled Counters (Expect plural "tools")

	// Success tool total
	// Expected key part: mcpany.tools.call.total;tool=test-service.success-tool
	// Current behavior (bug): mcpany.tool.call.total;tool=...

	// We construct the expected prefix
	expectedLabeledTotalPrefix := "mcpany.tools.call.total;tool=" + successID
	foundKey := findKey(expectedLabeledTotalPrefix)
	if foundKey == "" {
		// Fallback to check if the singular version exists (to confirm test setup is correct but code is "wrong")
		singularKey := findKey("mcpany.tool.call.total;tool=" + successID)
		if singularKey != "" {
			t.Logf("Found singular key: %s", singularKey)
		}
	}
	assert.NotEmpty(t, foundKey, "Should find labeled metric with 'tools' prefix for success tool total")

	// Error tool error count
	expectedLabeledErrorPrefix := "mcpany.tools.call.errors;tool=" + errorID
	foundKey = findKey(expectedLabeledErrorPrefix)
	if foundKey == "" {
		singularKey := findKey("mcpany.tool.call.errors;tool=" + errorID)
		if singularKey != "" {
			t.Logf("Found singular key: %s", singularKey)
		}
	}
	assert.NotEmpty(t, foundKey, "Should find labeled metric with 'tools' prefix for error tool errors")
}
