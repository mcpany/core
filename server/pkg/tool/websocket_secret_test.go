// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketTool_Execute_ResolvesSecrets(t *testing.T) {
	// Set up the environment variable for the secret
	secretKey := "TEST_SECRET_KEY"
	secretValue := "super_secret_value"
	t.Setenv(secretKey, secretValue)

	t.Run("resolves environment variable secret", func(t *testing.T) {
		// Mock server to verify the received message
		server := mockWebsocketServer(t, func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()

			_, msg, err := conn.ReadMessage()
			require.NoError(t, err)

			// Verify that the secret was resolved and sent in the message
			var received map[string]interface{}
			err = json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, secretValue, received["apiKey"])

			// Respond to complete the execution
			err = conn.WriteMessage(websocket.TextMessage, []byte(`{"status": "ok"}`))
			require.NoError(t, err)
		})
		defer server.Close()

		// Setup WebSocket connection
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		wrapper := &client.WebsocketClientWrapper{Conn: conn}

		// Setup Mock Pool
		pm := pool.NewManager()
		mockPool := &mockWebsocketPool{
			getFunc: func(_ context.Context) (*client.WebsocketClientWrapper, error) {
				return wrapper, nil
			},
			putFunc: func(c *client.WebsocketClientWrapper) {
				_ = c.Close()
			},
		}
		serviceID := "ws-secret-test"
		pm.Register(serviceID, mockPool)

		// Setup Tool Definition
		toolProto := v1.Tool_builder{}.Build()
		toolProto.SetName("secret-tool")
		toolProto.SetServiceId(serviceID)

		// Setup Call Definition with Secret Parameter
		callDef := &configv1.WebsocketCallDefinition{}

		// Build ParameterSchema
		paramSchema := configv1.ParameterSchema_builder{
			Name: proto.String("apiKey"),
		}.Build()

		// Build SecretValue
		secretVal := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String(secretKey),
		}.Build()

		// Build WebsocketParameterMapping
		secretParam := configv1.WebsocketParameterMapping_builder{
			Schema: paramSchema,
			Secret: secretVal,
		}.Build()

		callDef.SetParameters([]*configv1.WebsocketParameterMapping{secretParam})

		wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)

		// Execute
		inputs := json.RawMessage(`{}`)
		req := &ExecutionRequest{
			ToolName:   serviceID + "/-/secret-tool",
			ToolInputs: inputs,
		}

		result, err := wsTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"status": "ok"}, result)
	})
}
