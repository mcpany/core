package tool

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	bus_pb "github.com/mcpany/core/proto/bus"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestToolManager_MCPHandler(t *testing.T) {
	// Setup dependencies
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	tm := NewManager(busProvider)

	// Create real MCP server/client pair
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	// Connect
	ctx := context.Background()
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Prepare a tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	toolName := "handler-test-tool"
	serviceID := "test-service"
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullToolName := serviceID + "." + sanitizedName

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	err = tm.AddTool(mockTool)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		// Start a goroutine to act as the "worker" responding to the bus
		reqBus, err := bus.GetBus[*bus.ToolExecutionRequest](busProvider, "tool_execution_requests")
		require.NoError(t, err)
		resBus, err := bus.GetBus[*bus.ToolExecutionResult](busProvider, "tool_execution_results")
		require.NoError(t, err)

		unsub := reqBus.Subscribe(ctx, "request", func(req *bus.ToolExecutionRequest) {
			if req.ToolName == fullToolName {
				res := &bus.ToolExecutionResult{
					Result: json.RawMessage(`"success-result"`),
				}
				// Publish result with same correlation ID
				_ = resBus.Publish(ctx, req.CorrelationID(), res)
			}
		})
		defer unsub()

		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: fullToolName,
		})
		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		var resStr string
		err = json.Unmarshal([]byte(textContent.Text), &resStr)
		require.NoError(t, err)
		assert.Equal(t, "success-result", resStr)
	})

	t.Run("Timeout", func(t *testing.T) {
		// Temporarily reduce timeout
		oldTimeout := ToolExecutionTimeout
		ToolExecutionTimeout = 10 * time.Millisecond
		defer func() { ToolExecutionTimeout = oldTimeout }()

		// No worker responding here
		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: fullToolName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timed out waiting for tool execution")
	})

	t.Run("ContextCancelled", func(t *testing.T) {
		// Ensure timeout is long enough so we hit context deadline first
		oldTimeout := ToolExecutionTimeout
		ToolExecutionTimeout = 60 * time.Second
		defer func() { ToolExecutionTimeout = oldTimeout }()

		// Use a short timeout on context to let the call reach server but then cancel
		shortCtx, cancelShort := context.WithTimeout(ctx, 50*time.Millisecond)
		defer cancelShort()

		_, err = clientSession.CallTool(shortCtx, &mcp.CallToolParams{
			Name: fullToolName,
		})
		require.Error(t, err)
		// The server handler sees ctx.Done() and returns "context deadline exceeded while waiting..."
		// The client receives this error.
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

func TestToolManager_MCPHandler_BusErrors(t *testing.T) {
	// Need a fresh manager to inject bus hook
	oldHook := bus.GetBusHook
	defer func() { bus.GetBusHook = oldHook }()

	tm := NewManager(nil) // Bus provider will be mocked via hook

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test-server"}, nil)
	provider := &TestMCPServerProvider{server: mcpServer}
	tm.SetMCPServer(provider)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx := context.Background()
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	// Add tool
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
	serviceID := "test-service"
	toolName := "bus-error-tool"
	sanitizedName, _ := util.SanitizeToolName(toolName)
	fullToolName := serviceID + "." + sanitizedName

	mockTool := &MockTool{
		ToolFunc: func() *mcp_router_v1.Tool {
			return mcp_router_v1.Tool_builder{
				ServiceId: proto.String(serviceID),
				Name:      proto.String(toolName),
				Annotations: mcp_router_v1.ToolAnnotations_builder{
					InputSchema: inputSchema,
				}.Build(),
			}.Build()
		},
	}
	// Note: AddTool uses GetBus inside if mcpServer is set, but that's for the handler registration?
	// No, AddTool just registers the handler. GetBus is called INSIDE the handler (when tool is called).
	// EXCEPT AddTool *might* use bus if it does something else? No.
	// Wait, AddTool in management.go:
	/*
		if tm.mcpServer != nil {
			// ...
			handler := func(...) {
				// ...
				resultBus, err := bus.GetBus[*bus.ToolExecutionResult](tm.bus, "tool_execution_results")
				// ...
			}
			tm.mcpServer.Server().AddTool(mcpTool, handler)
		}
	*/
	// So GetBus is called when tool is CALLED.
	err = tm.AddTool(mockTool)
	require.NoError(t, err)

	t.Run("GetBusError", func(t *testing.T) {
		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			return nil, errors.New("mock bus error")
		}

		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: fullToolName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get result bus")
	})

	t.Run("PublishError", func(t *testing.T) {
		// Mock bus that fails on Publish
		mockReqBus := &MockBus[*bus.ToolExecutionRequest]{
			PublishFunc: func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
				return errors.New("mock publish error")
			},
		}
		// Result bus needs to be mocked too or standard
		mockResBus := &MockBus[*bus.ToolExecutionResult]{
			SubscribeOnceFunc: func(ctx context.Context, topic string, handler func(*bus.ToolExecutionResult)) func() {
				return func() {}
			},
		}

		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			if topic == "tool_execution_requests" {
				return mockReqBus, nil
			}
			if topic == "tool_execution_results" {
				return mockResBus, nil
			}
			return nil, nil
		}

		_, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
			Name: fullToolName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to publish request")
	})
}

// MockBus helper for testing
type MockBus[T any] struct {
	PublishFunc       func(ctx context.Context, topic string, msg T) error
	SubscribeOnceFunc func(ctx context.Context, topic string, handler func(T)) func()
}

func (m *MockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, topic, msg)
	}
	return nil
}

func (m *MockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) func() {
	return func() {}
}

func (m *MockBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) func() {
	if m.SubscribeOnceFunc != nil {
		return m.SubscribeOnceFunc(ctx, topic, handler)
	}
	return func() {}
}
