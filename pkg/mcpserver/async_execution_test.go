// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/worker"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mockSimpleTool implements tool.Tool interface
type mockSimpleTool struct {
	name string
}

func (m *mockSimpleTool) Tool() *mcp_v1.Tool {
	return mcp_v1.Tool_builder{
		Name:      proto.String(m.name),
		ServiceId: proto.String("async-service"),
	}.Build()
}

func (m *mockSimpleTool) Execute(_ context.Context, req *tool.ExecutionRequest) (any, error) {
	return "execution-success", nil
}

func (m *mockSimpleTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func TestAsyncToolExecution(t *testing.T) {
	// 1. Setup Bus
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	// 2. Setup Tool Manager with Bus
	tm := tool.NewManager(bp)

	// 3. Register a test tool
	testTool := &mockSimpleTool{name: "async-tool"}
	err = tm.AddTool(testTool)
	require.NoError(t, err)

	// 4. Setup Upstream Worker
	// Worker needs the same bus and the same manager (to execute locally).
	w := worker.NewUpstreamWorker(bp, tm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. Start Worker
	w.Start(ctx)

	// We need to wait a bit for worker subscription to be ready
	time.Sleep(100 * time.Millisecond)

	// 6. Execute Tool via Manager (which should trigger async flow)
	args := map[string]any{"some": "arg"}
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	// Testing tm.ExecuteTool (Async) -> Bus -> Worker -> tm.ExecuteToolLocally -> Tool
	result, err := tm.ExecuteTool(ctx, &tool.ExecutionRequest{
		ToolName:   "async-service.async-tool",
		ToolInputs: argsJSON,
	})

	require.NoError(t, err)
	assert.Equal(t, "execution-success", result)
}
