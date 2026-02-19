// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockBus implements bus.Bus[T] for testing
type mockBus[T any] struct {
	mu          sync.Mutex
	publishFunc func(ctx context.Context, topic string, msg T) error
	handlers    map[string][]func(T)
}

func newMockBus[T any]() *mockBus[T] {
	return &mockBus[T]{
		handlers: make(map[string][]func(T)),
	}
}

func (m *mockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	m.mu.Lock()
	if m.publishFunc != nil {
		m.mu.Unlock()
		return m.publishFunc(ctx, topic, msg)
	}
	m.mu.Unlock()
	return nil
}

func (m *mockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) func() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[topic] = append(m.handlers[topic], handler)
	return func() {}
}

func (m *mockBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) func() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[topic] = append(m.handlers[topic], handler)
	return func() {}
}

// TriggerHandler triggers registered handlers for a topic
func (m *mockBus[T]) TriggerHandler(topic string, msg T) {
	m.mu.Lock()
	handlers := m.handlers[topic]
	m.mu.Unlock()
	for _, h := range handlers {
		go h(msg)
	}
}

func TestHandler_Success(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Mock Bus
	reqBus := newMockBus[*bus.ToolExecutionRequest]()
	resBus := newMockBus[*bus.ToolExecutionResult]()

	var publishedReq *bus.ToolExecutionRequest
	var reqWg sync.WaitGroup
	reqWg.Add(1)

	reqBus.publishFunc = func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
		publishedReq = msg
		reqWg.Done()
		return nil
	}

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return reqBus, nil
		}
		if topic == "tool_execution_results" {
			return resBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	tool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("svc"),
				Name:        proto.String("tool"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	require.NoError(t, tm.AddTool(tool))

	// Connect Client
	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	ct, st := mcp_sdk.NewInMemoryTransports()
	go mcpServer.Connect(context.Background(), st, nil)

	clientSession, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Simulate Worker
	go func() {
		reqWg.Wait()
		require.NotNil(t, publishedReq)
		require.Equal(t, "svc.tool", publishedReq.ToolName)

		resBytes, err := json.Marshal(map[string]any{"status": "ok"})
		require.NoError(t, err)
		res := &bus.ToolExecutionResult{
			Result: resBytes,
		}
		resBus.TriggerHandler(publishedReq.CorrelationID(), res)
	}()

	// Call Tool
	sanitizedName, err := util.SanitizeToolName("tool")
	require.NoError(t, err)
	result, err := clientSession.CallTool(context.Background(), &mcp_sdk.CallToolParams{
		Name: "svc." + sanitizedName,
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	require.Len(t, result.Content, 1)
	textContent, ok := result.Content[0].(*mcp_sdk.TextContent)
	require.True(t, ok)

	var resMap map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resMap)
	require.NoError(t, err)
	assert.Equal(t, "ok", resMap["status"])
}

func TestHandler_Timeout(t *testing.T) {
	// Set short timeout
	origTimeout := ToolExecutionTimeout
	ToolExecutionTimeout = 100 * time.Millisecond
	defer func() { ToolExecutionTimeout = origTimeout }()

	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Mock Bus (Normal behavior, but we won't reply)
	reqBus := newMockBus[*bus.ToolExecutionRequest]()
	resBus := newMockBus[*bus.ToolExecutionResult]()

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return reqBus, nil
		}
		if topic == "tool_execution_results" {
			return resBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	tool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("svc"),
				Name:        proto.String("timeout"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	require.NoError(t, tm.AddTool(tool))

	// Connect Client
	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	ct, st := mcp_sdk.NewInMemoryTransports()
	go mcpServer.Connect(context.Background(), st, nil)

	clientSession, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call Tool - Should timeout
	start := time.Now()
	_, err = clientSession.CallTool(context.Background(), &mcp_sdk.CallToolParams{
		Name: "svc.timeout",
	})
	duration := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out waiting for tool execution")
	assert.GreaterOrEqual(t, duration, 100*time.Millisecond)
}

func TestHandler_ContextCancelled(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Mock Bus
	reqBus := newMockBus[*bus.ToolExecutionRequest]()
	resBus := newMockBus[*bus.ToolExecutionResult]()

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return reqBus, nil
		}
		if topic == "tool_execution_results" {
			return resBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	tool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("svc"),
				Name:        proto.String("cancel"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	require.NoError(t, tm.AddTool(tool))

	// Connect Client
	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	ct, st := mcp_sdk.NewInMemoryTransports()
	go mcpServer.Connect(context.Background(), st, nil)

	clientSession, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	// Call Tool with short context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: "svc.cancel",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestHandler_BusError(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		return nil, errors.New("bus failure")
	}
	defer func() { bus.GetBusHook = nil }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	tool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("svc"),
				Name:        proto.String("buserr"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	require.NoError(t, tm.AddTool(tool))

	// Connect Client
	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	ct, st := mcp_sdk.NewInMemoryTransports()
	go mcpServer.Connect(context.Background(), st, nil)

	clientSession, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	_, err = clientSession.CallTool(context.Background(), &mcp_sdk.CallToolParams{
		Name: "svc.buserr",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get result bus")
}

func TestHandler_PublishError(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	tm := NewManager(busProvider)
	mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
	tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

	// Mock Bus with Publish Error
	reqBus := newMockBus[*bus.ToolExecutionRequest]()
	resBus := newMockBus[*bus.ToolExecutionResult]()

	reqBus.publishFunc = func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
		return errors.New("publish failed")
	}

	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == "tool_execution_requests" {
			return reqBus, nil
		}
		if topic == "tool_execution_results" {
			return resBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	// Add Tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	tool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId:   proto.String("svc"),
				Name:        proto.String("puberr"),
				Annotations: mcp_router_v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
			}.Build()
		},
	}
	require.NoError(t, tm.AddTool(tool))

	// Connect Client
	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
	ct, st := mcp_sdk.NewInMemoryTransports()
	go mcpServer.Connect(context.Background(), st, nil)

	clientSession, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	defer func() { _ = clientSession.Close() }()

	_, err = clientSession.CallTool(context.Background(), &mcp_sdk.CallToolParams{
		Name: "svc.puberr",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to publish request")
}
