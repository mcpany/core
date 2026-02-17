package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
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
	publishFunc       func(ctx context.Context, topic string, msg T) error
	subscribeFunc     func(ctx context.Context, topic string, handler func(T)) func()
	subscribeOnceFunc func(ctx context.Context, topic string, handler func(T)) func()
}

func (m *mockBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, topic, msg)
	}
	return nil
}

func (m *mockBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) func() {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, topic, handler)
	}
	return func() {}
}

func (m *mockBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) func() {
	if m.subscribeOnceFunc != nil {
		return m.subscribeOnceFunc(ctx, topic, handler)
	}
	// Fallback to subscribeFunc logic if set
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, topic, handler)
	}
	return func() {}
}

func TestAddTool_Handler(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		tm := NewManager(nil) // Bus provided via hooks

		// Channels to coordinate
		reqChan := make(chan *bus.ToolExecutionRequest, 1)

		var resultHandler func(*bus.ToolExecutionResult)

		// Hook for Bus
		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			if topic == "tool_execution_requests" {
				return &mockBus[*bus.ToolExecutionRequest]{
					publishFunc: func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
						// When request is published, we send it to channel so we can inspect it and trigger response
						reqChan <- msg

						// Simulate worker processing: send result back using the captured handler
						// We run this in a goroutine to not block the publish
						go func() {
							// Wait for resultHandler to be set (it should be set by SubscribeOnce called BEFORE Publish)
							// But to be safe, we check
							if resultHandler != nil {
								res := &bus.ToolExecutionResult{
									BaseMessage: bus.BaseMessage{CID: msg.CorrelationID()},
									Result:      json.RawMessage(`{"status":"success"}`),
								}
								resultHandler(res)
							}
						}()
						return nil
					},
				}, nil
			}
			if topic == "tool_execution_results" {
				return &mockBus[*bus.ToolExecutionResult]{
					subscribeOnceFunc: func(ctx context.Context, topic string, handler func(*bus.ToolExecutionResult)) func() {
						resultHandler = handler
						return func() {}
					},
				}, nil
			}
			return nil, nil
		}
		defer func() { bus.GetBusHook = nil }()

		// Setup MCP Server and Client
		mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
		tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

		// Add Tool
		inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return v1.Tool_builder{
					ServiceId: proto.String("s1"),
					Name:      proto.String("t1"),
					Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
				}.Build()
			},
		}
		err := tm.AddTool(mockTool)
		require.NoError(t, err)

		// Connect Client
		client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
		ct, st := mcp_sdk.NewInMemoryTransports()
		ctx := context.Background()

		// Connect server in background
		go mcpServer.Connect(ctx, st, nil)

		clientSession, err := client.Connect(ctx, ct, nil)
		require.NoError(t, err)
		defer clientSession.Close()

		// Call Tool
		sanitizedName, _ := util.SanitizeToolName("t1")
		res, err := clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
			Name: "s1." + sanitizedName,
			Arguments: map[string]interface{}{"arg": "val"},
		})
		require.NoError(t, err)

		// Verify Result
		require.Len(t, res.Content, 1)
		txt, ok := res.Content[0].(*mcp_sdk.TextContent)
		require.True(t, ok)
		assert.Contains(t, txt.Text, "success")

		// Verify request content
		select {
		case req := <-reqChan:
			assert.Equal(t, "s1.t1", req.ToolName)
			// ToolInputs is json.RawMessage
			var inputs map[string]interface{}
			_ = json.Unmarshal(req.ToolInputs, &inputs)
			assert.Equal(t, "val", inputs["arg"])
		default:
			t.Fatal("request not received")
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		tm := NewManager(nil)
		tm.toolExecutionTimeout = 10 * time.Millisecond

		// Hook for Bus - Do NOT respond
		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			if topic == "tool_execution_requests" {
				return &mockBus[*bus.ToolExecutionRequest]{
					publishFunc: func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
						return nil // Send to void
					},
				}, nil
			}
			if topic == "tool_execution_results" {
				return &mockBus[*bus.ToolExecutionResult]{
					subscribeOnceFunc: func(ctx context.Context, topic string, handler func(*bus.ToolExecutionResult)) func() {
						return func() {}
					},
				}, nil
			}
			return nil, nil
		}
		defer func() { bus.GetBusHook = nil }()

		mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
		tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

		inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return v1.Tool_builder{
					ServiceId: proto.String("s1"),
					Name:      proto.String("timeout_tool"),
					Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
				}.Build()
			},
		}
		err := tm.AddTool(mockTool)
		require.NoError(t, err)

		client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
		ct, st := mcp_sdk.NewInMemoryTransports()
		ctx := context.Background()
		go mcpServer.Connect(ctx, st, nil)

		clientSession, err := client.Connect(ctx, ct, nil)
		require.NoError(t, err)
		defer clientSession.Close()

		sanitizedName, _ := util.SanitizeToolName("timeout_tool")
		_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
			Name: "s1." + sanitizedName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "timed out waiting for tool execution")
	})

	t.Run("Bus Error GetBus", func(t *testing.T) {
		tm := NewManager(nil)

		// Hook for Bus - Error on GetBus
		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			return nil, fmt.Errorf("injected bus error")
		}
		defer func() { bus.GetBusHook = nil }()

		mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
		tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

		inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return v1.Tool_builder{
					ServiceId: proto.String("s1"),
					Name:      proto.String("bus_error_tool"),
					Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
				}.Build()
			},
		}
		err := tm.AddTool(mockTool)
		require.NoError(t, err)

		client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
		ct, st := mcp_sdk.NewInMemoryTransports()
		ctx := context.Background()
		go mcpServer.Connect(ctx, st, nil)

		clientSession, err := client.Connect(ctx, ct, nil)
		require.NoError(t, err)
		defer clientSession.Close()

		sanitizedName, _ := util.SanitizeToolName("bus_error_tool")
		_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
			Name: "s1." + sanitizedName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get result bus")
	})

	t.Run("Bus Error Publish", func(t *testing.T) {
		tm := NewManager(nil)

		bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
			if topic == "tool_execution_requests" {
				return &mockBus[*bus.ToolExecutionRequest]{
					publishFunc: func(ctx context.Context, topic string, msg *bus.ToolExecutionRequest) error {
						return fmt.Errorf("publish failed")
					},
				}, nil
			}
			if topic == "tool_execution_results" {
				return &mockBus[*bus.ToolExecutionResult]{}, nil
			}
			return nil, nil
		}
		defer func() { bus.GetBusHook = nil }()

		mcpServer := mcp_sdk.NewServer(&mcp_sdk.Implementation{Name: "server"}, nil)
		tm.SetMCPServer(&TestMCPServerProvider{server: mcpServer})

		inputSchema, _ := structpb.NewStruct(map[string]interface{}{"type": "object"})
		mockTool := &MockTool{
			ToolFunc: func() *v1.Tool {
				return v1.Tool_builder{
					ServiceId: proto.String("s1"),
					Name:      proto.String("publish_error_tool"),
					Annotations: v1.ToolAnnotations_builder{InputSchema: inputSchema}.Build(),
				}.Build()
			},
		}
		err := tm.AddTool(mockTool)
		require.NoError(t, err)

		client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "client"}, nil)
		ct, st := mcp_sdk.NewInMemoryTransports()
		ctx := context.Background()
		go mcpServer.Connect(ctx, st, nil)

		clientSession, err := client.Connect(ctx, ct, nil)
		require.NoError(t, err)
		defer clientSession.Close()

		sanitizedName, _ := util.SanitizeToolName("publish_error_tool")
		_, err = clientSession.CallTool(ctx, &mcp_sdk.CallToolParams{
			Name: "s1." + sanitizedName,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to publish request")
	})
}
