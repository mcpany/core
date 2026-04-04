// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketTool_New(t *testing.T) {
	// Setup a test websocket server that echoes back
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			// If it's json, modify it slightly or just echo
			if strings.HasPrefix(string(message), "{") {
				var msg map[string]interface{}
				if err := json.Unmarshal(message, &msg); err == nil {
					msg["echo"] = true
					resp, _ := json.Marshal(msg)
					err = c.WriteMessage(mt, resp)
				} else {
					err = c.WriteMessage(mt, message)
				}
			} else {
				// non-json response
				err = c.WriteMessage(mt, message)
			}

			if err != nil {
				break
			}
		}
	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	poolManager := pool.NewManager()
	serviceID := "test-ws-service"

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(wsURL, nil)
	require.NoError(t, err)
	wrapper := &client.WebsocketClientWrapper{Conn: conn}

	mockPool := &mockWebsocketPool{
		getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
			return wrapper, nil
		},
		putFunc: func(c *client.WebsocketClientWrapper) {
		},
	}
	poolManager.Register(serviceID, mockPool)

	toolProto := &v1.Tool{
		Name: proto.String("ws-tool"),
		Description: proto.String("A websocket tool"),
	}

	cacheConfig := &configv1.CacheConfig{
		Enabled: proto.Bool(true),
	}

	callDef := &configv1.WebsocketCallDefinition{
		Cache: cacheConfig,
	}

	t.Run("NewWebsocketTool", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDef)
		require.NotNil(t, tool)
		assert.Equal(t, toolProto, tool.Tool())
		assert.Equal(t, cacheConfig, tool.GetCacheConfig())
		assert.False(t, tool.IsStreaming())
	})

	t.Run("MCPTool", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDef)
		mcpTool := tool.MCPTool()
		assert.NotNil(t, mcpTool)
		assert.Equal(t, ".ws-tool", mcpTool.Name)
	})

	t.Run("Execute_Success", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{"hello": "world"}`),
		}

		res, err := tool.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "world", resMap["hello"])
		assert.Equal(t, true, resMap["echo"])
	})

	t.Run("Execute_NoPool", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, "invalid-service", nil, callDef)
		req := &ExecutionRequest{
			ToolInputs: []byte(`{"hello": "world"}`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no websocket pool found")
	})

	t.Run("Execute_InvalidJSONInput", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDef)
		req := &ExecutionRequest{
			ToolInputs: []byte(`invalid json`),
		}
		_, err := tool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("Execute_WithInputTransformer", func(t *testing.T) {
		transformerDef := &configv1.InputTransformer{
			Template: proto.String(`{"transformed": "{{.val}}"}`),
		}
		callDefWithTransform := &configv1.WebsocketCallDefinition{
			InputTransformer: transformerDef,
		}

		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDefWithTransform)
		req := &ExecutionRequest{
			ToolInputs: []byte(`{"val": "data"}`),
		}

		res, err := tool.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "data", resMap["transformed"])
	})

	t.Run("Execute_NonJSONResponse", func(t *testing.T) {
		transformerDef := &configv1.InputTransformer{
			Template: proto.String(`plain text message`),
		}
		callDefWithTransform := &configv1.WebsocketCallDefinition{
			InputTransformer: transformerDef,
		}

		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDefWithTransform)
		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		res, err := tool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "plain text message", res) // String fallback when JSON unmarshal fails
	})

	t.Run("StreamExecute", func(t *testing.T) {
		tool := NewWebsocketTool(toolProto, poolManager, serviceID, nil, callDef)
		req := &ExecutionRequest{
			ToolInputs: []byte(`{"hello": "world"}`),
		}

		ch, err := tool.StreamExecute(context.Background(), req)
		require.NoError(t, err)

		res := <-ch

		resMap, ok := res.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "world", resMap["hello"])
	})
}
