// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
)

func TestWebsocketTool_NewWebsocketTool(t *testing.T) {
	toolDef := &v1.Tool{Name: ptr("websocket-tool")}
	callDef := &configv1.WebsocketCallDefinition{
		Parameters: []*configv1.WebsocketParameterMapping{},
	}
	poolManager := pool.NewManager()

	tool := NewWebsocketTool(toolDef, poolManager, "service-id", nil, callDef)
	require.NotNil(t, tool)
	assert.Equal(t, toolDef, tool.Tool())
	assert.Equal(t, "service-id", tool.serviceID)
}

func TestWebsocketTool_MCPTool(t *testing.T) {
	toolDef := &v1.Tool{
		Name:        ptr("websocket-tool"),
		Description: ptr("A test websocket tool"),
	}
	callDef := &configv1.WebsocketCallDefinition{}

	tool := NewWebsocketTool(toolDef, nil, "", nil, callDef)
	mcpTool := tool.MCPTool()
	require.NotNil(t, mcpTool)
	assert.Contains(t, mcpTool.Name, "websocket-tool")
	assert.Equal(t, "A test websocket tool", mcpTool.Description)
}

func TestWebsocketTool_GetCacheConfig(t *testing.T) {
	enabled := true
	cacheConfig := &configv1.CacheConfig{IsEnabled: &enabled}
	callDef := &configv1.WebsocketCallDefinition{
		Cache: cacheConfig,
	}
	tool := NewWebsocketTool(&v1.Tool{}, nil, "", nil, callDef)
	assert.Equal(t, cacheConfig, tool.GetCacheConfig())
}

func TestWebsocketTool_Execute_Custom(t *testing.T) {
	// Start a mock websocket server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			// Echo back
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer s.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	// Setup pool
	poolManager := pool.NewManager()
	serviceID := "test-service"
	p, err := pool.New(func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			return nil, err
		}
		return &client.WebsocketClientWrapper{Conn: c}, nil
	}, 1, 1, time.Minute, true)
	require.NoError(t, err)
	poolManager.Register(serviceID, p)

	// Create tool
	toolDef := &v1.Tool{Name: ptr("ws-echo")}
	callDef := &configv1.WebsocketCallDefinition{}

	wsTool := NewWebsocketTool(toolDef, poolManager, serviceID, nil, callDef)

	// Execute
	ctx := context.Background()
	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"msg": "hello"}`),
	}

	res, err := wsTool.Execute(ctx, req)
	require.NoError(t, err)

	// The echo server returns the JSON input as is
	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "hello", resMap["msg"])
}
