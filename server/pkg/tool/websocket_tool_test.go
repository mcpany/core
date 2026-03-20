// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServiceID = "test-service"
)

func TestNewWebsocketTool(t *testing.T) {
	t.Parallel()
	pm := pool.NewManager()
	serviceID := testServiceID
	toolProto := v1.Tool_builder{}.Build()
	toolProto.SetName("test-tool")
	toolProto.SetServiceId(serviceID)
	callDef := &configv1.WebsocketCallDefinition{}

	wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)
	require.NotNil(t, wsTool)
	assert.Equal(t, toolProto, wsTool.Tool())
	assert.Equal(t, pm, wsTool.poolManager)
	assert.Equal(t, serviceID, wsTool.serviceID)
}

var upgrader = websocket.Upgrader{}

func mockWebsocketServer(
	_ *testing.T,
	handler func(w http.ResponseWriter, r *http.Request),
) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(handler))
}

// mockWebsocketPool implements the pool.Pool interface for testing.
type mockWebsocketPool struct {
	getFunc func(ctx context.Context) (*client.WebsocketClientWrapper, error)
	putFunc func(c *client.WebsocketClientWrapper)
}

func (m *mockWebsocketPool) Get(ctx context.Context) (*client.WebsocketClientWrapper, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx)
	}
	return nil, errors.New("get not implemented")
}

func (m *mockWebsocketPool) Put(c *client.WebsocketClientWrapper) {
	if m.putFunc != nil {
		m.putFunc(c)
	}
}

func (m *mockWebsocketPool) Close() error {
	// No-op for mock
	return nil
}

func (m *mockWebsocketPool) Len() int {
	return 0 // Or some mock value
}

func TestWebsocketTool_Execute(t *testing.T) {
	t.Parallel()
	t.Run("successful execution", func(t *testing.T) {
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
			_, msg, err := conn.ReadMessage()
			require.NoError(t, err)
			// Echo the message back
			err = conn.WriteMessage(websocket.TextMessage, msg)
			require.NoError(t, err)
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		wrapper := &client.WebsocketClientWrapper{Conn: conn}

		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return wrapper, nil
			},
			putFunc: func(c *client.WebsocketClientWrapper) {
				_ = c.Close()
			},
		}
		serviceID := "ws-test-success"
		pm.Register(serviceID, mockPool)

		toolProto := v1.Tool_builder{}.Build()
		toolProto.SetName("echo")
		toolProto.SetServiceId(serviceID)
		toolProto.SetUnderlyingMethodFqn("WS " + wsURL)

		callDef := &configv1.WebsocketCallDefinition{}
		wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)

		inputs := json.RawMessage(`{"message": "hello"}`)
		req := &ExecutionRequest{
			ToolName:   serviceID + "/-/echo",
			ToolInputs: inputs,
		}

		result, err := wsTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"message": "hello"}, result)
	})

	t.Run("with input and output transformation", func(t *testing.T) {
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
			_, msg, err := conn.ReadMessage()
			require.NoError(t, err)
			assert.Equal(t, `{"transformed_message":"hello"}`, string(msg))
			// Respond with a message that can be transformed
			err = conn.WriteMessage(websocket.TextMessage, []byte(`{"response_data": "world"}`))
			require.NoError(t, err)
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		wrapper := &client.WebsocketClientWrapper{Conn: conn}

		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return wrapper, nil
			},
			putFunc: func(c *client.WebsocketClientWrapper) {
				_ = c.Close()
			},
		}
		serviceID := "ws-transform-test"
		pm.Register(serviceID, mockPool)

		toolProto := v1.Tool_builder{}.Build()
		toolProto.SetName("transform")
		toolProto.SetServiceId(serviceID)

		callDef := &configv1.WebsocketCallDefinition{}
		inputTransformer := &configv1.InputTransformer{}
		inputTransformer.SetTemplate(`{"transformed_message":"{{message}}"}`)
		callDef.SetInputTransformer(inputTransformer)
		outputTransformer := &configv1.OutputTransformer{}
		outputTransformer.SetFormat(configv1.OutputTransformer_JSON)
		outputTransformer.SetExtractionRules(map[string]string{
			"final_output": "{.response_data}",
		})
		callDef.SetOutputTransformer(outputTransformer)
		wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)

		inputs := json.RawMessage(`{"message": "hello"}`)
		req := &ExecutionRequest{
			ToolName:   serviceID + "/-/transform",
			ToolInputs: inputs,
		}

		result, err := wsTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"final_output": "world"}, result)
	})

	t.Run("pool get error", func(t *testing.T) {
		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return nil, errors.New("pool error")
			},
		}
		serviceID := "ws-pool-error"
		pm.Register(serviceID, mockPool)

		toolProto := v1.Tool_builder{}.Build()
		toolProto.SetName("error")
		callDef := &configv1.WebsocketCallDefinition{}
		wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)

		req := &ExecutionRequest{}
		_, err := wsTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pool error")
	})

	t.Run("pool not found", func(t *testing.T) {
		pm := pool.NewManager()
		toolProto := v1.Tool_builder{}.Build()
		wsTool := NewWebsocketTool(
			toolProto,
			pm,
			"non-existent-service",
			nil,
			&configv1.WebsocketCallDefinition{},
		)
		_, err := wsTool.Execute(context.Background(), &ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no websocket pool found for service")
	})

	t.Run("bad tool input", func(t *testing.T) {
		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return &client.WebsocketClientWrapper{}, nil
			},
		}
		serviceID := "ws-bad-input"
		pm.Register(serviceID, mockPool)
		toolProto := v1.Tool_builder{}.Build()
		wsTool := NewWebsocketTool(
			toolProto,
			pm,
			serviceID,
			nil,
			&configv1.WebsocketCallDefinition{},
		)
		req := &ExecutionRequest{ToolInputs: json.RawMessage(`{`)}
		_, err := wsTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("input_transformer_template_error", func(t *testing.T) {
		pm := pool.NewManager()
		// A pool must be registered for the service key, otherwise the function errors out
		// before it gets to the input transformation logic.
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				// This part should not be reached if the template parsing fails first.
				// We return a non-nil wrapper to satisfy the immediate checks.
				return &client.WebsocketClientWrapper{}, nil
			},
		}
		serviceID := "ws-template-error"
		pm.Register(serviceID, mockPool)

		callDef := &configv1.WebsocketCallDefinition{}
		it := &configv1.InputTransformer{}
		it.SetTemplate(`{{.invalid`) // Invalid template
		callDef.SetInputTransformer(it)
		wsTool := NewWebsocketTool(v1.Tool_builder{}.Build(), pm, serviceID, nil, callDef)

		req := &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := wsTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create input template")
	})

	t.Run("write_message_error", func(t *testing.T) {
		// A simple server that just upgrades the connection.
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			// The test will close the client-side connection, so we just wait here.
			// Reading a message prevents the handler from returning and closing the server early.
			_, _, _ = conn.ReadMessage()
		})
		defer server.Close()

		conn, resp, err := websocket.DefaultDialer.Dial(
			strings.Replace(server.URL, "http", "ws", 1),
			nil,
		)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }() // Close the connection on the client side immediately to ensure WriteMessage fails.
		wrapper := &client.WebsocketClientWrapper{Conn: conn}
		_ = conn.Close()

		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return wrapper, nil
			},
			putFunc: func(_ *client.WebsocketClientWrapper) { /* no-op, already closed */ },
		}
		serviceID := "ws-write-error"
		pm.Register(serviceID, mockPool)
		wsTool := NewWebsocketTool(
			v1.Tool_builder{}.Build(),
			pm,
			serviceID,
			nil,
			&configv1.WebsocketCallDefinition{},
		)

		_, err = wsTool.Execute(
			context.Background(),
			&ExecutionRequest{ToolInputs: json.RawMessage(`{}`)},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send message over websocket")
	})

	t.Run("read_message_error", func(t *testing.T) {
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
			// Read the incoming message and then immediately close with an error.
			_, _, err = conn.ReadMessage()
			require.NoError(t, err)
			_ = conn.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "read error"),
			)
		})
		defer server.Close()

		conn, resp, err := websocket.DefaultDialer.Dial(
			strings.Replace(server.URL, "http", "ws", 1),
			nil,
		)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		wrapper := &client.WebsocketClientWrapper{Conn: conn}

		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return wrapper, nil
			},
			putFunc: func(c *client.WebsocketClientWrapper) { _ = c.Close() },
		}
		serviceID := "ws-read-error"
		pm.Register(serviceID, mockPool)
		wsTool := NewWebsocketTool(
			v1.Tool_builder{}.Build(),
			pm,
			serviceID,
			nil,
			&configv1.WebsocketCallDefinition{},
		)

		_, err = wsTool.Execute(
			context.Background(),
			&ExecutionRequest{ToolInputs: json.RawMessage(`{}`)},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read message from websocket")
	})

	t.Run("non-json_response", func(t *testing.T) {
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
			_, _, _ = conn.ReadMessage()
			_ = conn.WriteMessage(websocket.TextMessage, []byte("this is not json"))
		})
		defer server.Close()

		conn, resp, err := websocket.DefaultDialer.Dial(
			strings.Replace(server.URL, "http", "ws", 1),
			nil,
		)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		wrapper := &client.WebsocketClientWrapper{Conn: conn}

		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) { return wrapper, nil },
			putFunc: func(c *client.WebsocketClientWrapper) { _ = c.Close() },
		}
		serviceID := "ws-non-json"
		pm.Register(serviceID, mockPool)
		wsTool := NewWebsocketTool(
			v1.Tool_builder{}.Build(),
			pm,
			serviceID,
			nil,
			&configv1.WebsocketCallDefinition{},
		)

		result, err := wsTool.Execute(
			context.Background(),
			&ExecutionRequest{ToolInputs: json.RawMessage(`{}`)},
		)
		require.NoError(t, err)
		assert.Equal(t, "this is not json", result)
	})
}
