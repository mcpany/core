// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// MockBus implements bus.Bus[T] for testing
type MockBus[T any] struct {
	PublishFunc func(context.Context, string, T) error
}

func (m *MockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, msg)
	}
	return nil
}

func (m *MockBus[T]) Subscribe(_ context.Context, _ string, _ func(T)) func() {
	return func() {}
}

func (m *MockBus[T]) SubscribeOnce(_ context.Context, _ string, _ func(T)) func() {
	return func() {}
}

func TestToolManager_MCPHandler_Success(t *testing.T) {
	// Setup Bus
	busProvider, err := bus.NewProvider(nil)
	require.NoError(t, err)

	// Setup Manager
	tm := NewManager(busProvider)

	// Setup MCP Server & Client
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect
	ctx := context.Background()
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "success-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err = tm.AddTool(mockTool)
	require.NoError(t, err)

	// Worker Simulation
	reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](busProvider, bus.ToolExecutionRequestTopic)
	require.NoError(t, err)

	resBus, err := bus.GetBus[*bus.ToolExecutionResult](busProvider, bus.ToolExecutionResultTopic)
	require.NoError(t, err)

	reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		// Verify Request
		if req.ToolName != serviceID+"."+toolName {
			return
		}

		// Send Result
		res := &bus.ToolExecutionResult{
			Result: json.RawMessage(`{"status": "success"}`),
		}
		res.SetCorrelationID(req.CorrelationID())
		_ = resBus.Publish(ctx, req.CorrelationID(), res)
	})

	// Call Tool
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullName := serviceID + "." + sanitizedName

	res, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      fullName,
		Arguments: map[string]interface{}{"arg": "val"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)

	txt, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, txt.Text, "success")
}

func TestToolManager_MCPHandler_Timeout(t *testing.T) {
	// Restore timeout after test
	origTimeout := ToolExecutionTimeout
	ToolExecutionTimeout = 10 * time.Millisecond
	defer func() { ToolExecutionTimeout = origTimeout }()

	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	_, _ = mcpServer.Connect(ctx, st, nil)
	clientSession, _ := client.Connect(ctx, ct, nil)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("s"),
				Name:        proto.String("timeout-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	// No worker running -> Timeout

	_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "s.timeout-tool"})
	require.Error(t, err)
	// Check for "timed out" in error message
	assert.Contains(t, err.Error(), "timed out waiting")
}

func TestToolManager_MCPHandler_BusError_GetBus(t *testing.T) {
	// Mock bus hook to fail GetBus
	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == bus.ToolExecutionResultTopic {
			return nil, fmt.Errorf("simulated bus error")
		}
		// For req bus, return valid one or loop?
		// Since we don't care about req bus success here if we fail fast.
		// However AddTool calls GetBus for result FIRST.
		// Let's check logic:
		// 1. Get result bus.
		// 2. Get request bus.
		// So if we fail results, we are good.
		return nil, nil // Should crash if casted, but we want error
	}
	defer func() { bus.GetBusHook = nil }()

	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	_, _ = mcpServer.Connect(ctx, st, nil)
	clientSession, _ := client.Connect(ctx, ct, nil)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("s"),
				Name:        proto.String("bus-error-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "s.bus-error-tool"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get result bus")
}

func TestToolManager_MCPHandler_BusError_Publish(t *testing.T) {
	mockReqBus := &MockBus[*bus.ToolExecutionRequest]{
		PublishFunc: func(_ context.Context, _ string, _ *bus.ToolExecutionRequest) error {
			return fmt.Errorf("publish failed")
		},
	}
	mockResBus := &MockBus[*bus.ToolExecutionResult]{}

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == bus.ToolExecutionRequestTopic {
			return mockReqBus, nil
		}
		if topic == bus.ToolExecutionResultTopic {
			return mockResBus, nil
		}
		return nil, fmt.Errorf("unexpected topic")
	}
	defer func() { bus.GetBusHook = nil }()

	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	_, _ = mcpServer.Connect(ctx, st, nil)
	clientSession, _ := client.Connect(ctx, ct, nil)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("s"),
				Name:        proto.String("pub-error-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "s.pub-error-tool"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish request")
}

func TestToolManager_MCPHandler_ExecutionError(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	_, _ = mcpServer.Connect(ctx, st, nil)
	clientSession, _ := client.Connect(ctx, ct, nil)

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("s"),
				Name:        proto.String("error-tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	_ = tm.AddTool(mockTool)

	// Worker responding with error
	reqBus, _ := bus.GetBus[*bus.ToolExecutionRequest](busProvider, bus.ToolExecutionRequestTopic)
	resBus, _ := bus.GetBus[*bus.ToolExecutionResult](busProvider, bus.ToolExecutionResultTopic)

	reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		res := &bus.ToolExecutionResult{
			Error: fmt.Errorf("worker execution failed"),
		}
		res.SetCorrelationID(req.CorrelationID())
		_ = resBus.Publish(ctx, req.CorrelationID(), res)
	})

	_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "s.error-tool"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "worker execution failed")
}
