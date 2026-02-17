// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// MockBus implements bus.Bus[T] for testing.
type MockBus[T any] struct {
	PublishFunc       func(ctx context.Context, topic string, msg T) error
	SubscribeFunc     func(ctx context.Context, topic string, handler func(T)) (unsubscribe func())
	SubscribeOnceFunc func(ctx context.Context, topic string, handler func(T)) (unsubscribe func())
}

func (m *MockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, msg)
	}
	return nil
}

func (m *MockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, topic, handler)
	}
	return func() {}
}

func (m *MockBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) (unsubscribe func()) {
	if m.SubscribeOnceFunc != nil {
		return m.SubscribeOnceFunc(ctx, topic, handler)
	}
	return func() {}
}

func TestToolManager_AsyncHandler_Success(t *testing.T) {
	// Override Timeout for speed
	origTimeout := ToolExecutionTimeout
	ToolExecutionTimeout = 2 * time.Second
	defer func() { ToolExecutionTimeout = origTimeout }()

	// Setup real in-memory bus
	messageBusConfig := bus_pb.MessageBus_builder{}.Build()
	messageBusConfig.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBusConfig)
	require.NoError(t, err)

	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Add tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "async-tool"
	serviceID := "test-service"
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId:   proto.String(serviceID),
				Name:        proto.String(toolName),
				Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	err = tm.AddTool(mockTool)
	require.NoError(t, err)

	// Setup Worker to respond
	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](busProvider, "tool_execution_requests")
	require.NoError(t, err)
	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](busProvider, "tool_execution_results")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requestBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
		// Echo back result
		res := &bus.ToolExecutionResult{
			Result: json.RawMessage(`{"status": "success"}`),
		}
		res.SetCorrelationID(req.CorrelationID())
		_ = resultBus.Publish(ctx, req.CorrelationID(), res)
	})

	// Client call
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect server and client
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer func() { _ = serverSession.Close() }()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: serviceID + "." + toolName,
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	// Verify result content
	require.NotEmpty(t, result.Content)
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "success")
}

func TestToolManager_AsyncHandler_BusFailure_GetBus(t *testing.T) {
	// Inject error for GetBus
	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_results" {
			return nil, errors.New("injected bus error")
		}
		return nil, nil // Fallback
	}
	defer func() { bus.GetBusHook = nil }()

	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Add tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{
				ServiceId:   proto.String("s"),
				Name:        proto.String("t"),
				Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	tm.AddTool(mockTool)

	// Client call
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, _ := mcpServer.Connect(context.Background(), serverTransport, nil)
	defer serverSession.Close()
	clientSession, _ := client.Connect(context.Background(), clientTransport, nil)
	defer clientSession.Close()

	_, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "s.t"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get result bus")
	assert.Contains(t, err.Error(), "injected bus error")
}

func TestToolManager_AsyncHandler_BusFailure_Publish(t *testing.T) {
	// Mock Request Bus to fail on Publish
	mockReqBus := &MockBus[*bus.ToolExecutionRequest]{
		PublishFunc: func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
			return errors.New("injected publish error")
		},
	}

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return mockReqBus, nil
		}
		return nil, nil // Fallback
	}
	defer func() { bus.GetBusHook = nil }()

	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{ServiceId: proto.String("s"), Name: proto.String("t"), Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build()}.Build()
		},
	}
	tm.AddTool(mockTool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, _ := mcpServer.Connect(context.Background(), serverTransport, nil)
	defer serverSession.Close()
	clientSession, _ := client.Connect(context.Background(), clientTransport, nil)
	defer clientSession.Close()

	_, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "s.t"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish request")
	assert.Contains(t, err.Error(), "injected publish error")
}

func TestToolManager_AsyncHandler_Timeout(t *testing.T) {
	// Override Timeout
	origTimeout := ToolExecutionTimeout
	ToolExecutionTimeout = 50 * time.Millisecond
	defer func() { ToolExecutionTimeout = origTimeout }()

	busProvider, _ := bus.NewProvider(nil) // InMemory bus
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{ServiceId: proto.String("s"), Name: proto.String("t"), Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build()}.Build()
		},
	}
	tm.AddTool(mockTool)

	// No worker running to respond!

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, _ := mcpServer.Connect(context.Background(), serverTransport, nil)
	defer serverSession.Close()
	clientSession, _ := client.Connect(context.Background(), clientTransport, nil)
	defer clientSession.Close()

	start := time.Now()
	_, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{Name: "s.t"})
	duration := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for tool execution")
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

func TestToolManager_AsyncHandler_ContextCancel(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return v1.Tool_builder{ServiceId: proto.String("s"), Name: proto.String("t"), Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build()}.Build()
		},
	}
	tm.AddTool(mockTool)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, _ := mcpServer.Connect(context.Background(), serverTransport, nil)
	defer serverSession.Close()
	clientSession, _ := client.Connect(context.Background(), clientTransport, nil)
	defer clientSession.Close()

	// Use very short context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Ensure execution takes longer than context
	// We do this by not starting a worker, so it waits.

	_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "s.t"})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "context deadline exceeded")
}
