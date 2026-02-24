package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startTestWebsocketServer(t *testing.T, handler func(*websocket.Conn)) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade failed: %v", err)
			return
		}
		defer c.Close()
		handler(c)
	}))
	return s
}

type noOpAuthenticator struct{}

func (n *noOpAuthenticator) Authenticate(req *http.Request) error { return nil }

func TestNewWebsocketTool(t *testing.T) {

	pm := pool.NewManager()
	serviceID := "test-service"
	toolProto := v1.Tool_builder{
		Name:      stringPtrWs("test-tool"),
		ServiceId: stringPtrWs(serviceID),
	}.Build()
	callDef := &configv1.WebsocketCallDefinition{}

	wsTool := NewWebsocketTool(toolProto, pm, serviceID, nil, callDef)
	require.NotNil(t, wsTool)
	assert.Equal(t, toolProto, wsTool.Tool())
	assert.Equal(t, pm, wsTool.poolManager)
	assert.Equal(t, serviceID, wsTool.serviceID)
}

func TestWebsocketTool_Execute(t *testing.T) {


	t.Run("Happy Path - Echo", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			// Echo back
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				return
			}
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-1", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("echo"),
		}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "service-1", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{"message": "hello"}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "hello", resMap["message"])
	})

	t.Run("Input Transformation", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			// Respond with received message wrapped
			response := map[string]string{"received": string(msg)}
			out, _ := json.Marshal(response)
			c.WriteMessage(websocket.TextMessage, out)
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-2", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("transform-input"),
		}.Build()
		callDef := configv1.WebsocketCallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: stringPtrWs(`{"wrapper": "{{input}}"}`),
			}.Build(),
		}.Build()

		wt := NewWebsocketTool(toolDef, m, "service-2", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{"input": "value"}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]any)
		require.True(t, ok)
		// Check that server received the transformed input
		assert.Contains(t, resMap["received"], `{"wrapper": "value"}`)
	})

	t.Run("Secret Resolution", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			_, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			// Echo back
			c.WriteMessage(websocket.TextMessage, msg)
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-3", p)
		defer m.CloseAll()

		os.Setenv("TEST_SECRET", "secret-value")
		defer os.Unsetenv("TEST_SECRET")

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("secret-tool"),
		}.Build()

		callDef := configv1.WebsocketCallDefinition_builder{
			Parameters: []*configv1.WebsocketParameterMapping{
				configv1.WebsocketParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: stringPtrWs("token"),
					}.Build(),
					Secret: configv1.SecretValue_builder{
						EnvironmentVariable: stringPtrWs("TEST_SECRET"),
					}.Build(),
				}.Build(),
			},
		}.Build()

		wt := NewWebsocketTool(toolDef, m, "service-3", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "secret-value", resMap["token"])
	})

	t.Run("Pool Error", func(t *testing.T) {
		m := pool.NewManager()
		toolDef := v1.Tool_builder{
			Name: stringPtrWs("fail-tool"),
		}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "non-existent-service", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		_, err := wt.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no websocket pool found")
	})

	t.Run("Connection Error", func(t *testing.T) {
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			return nil, fmt.Errorf("connection failed")
		}
		p, err := pool.New(factory, 0, 1, 1, 0, false)
		require.NoError(t, err)

		m := pool.NewManager()
		m.Register("fail-service", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("fail-conn-tool"),
		}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "fail-service", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		_, err = wt.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get websocket connection")
	})

	t.Run("Output Transformation - JQ", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			_, _, err := c.ReadMessage()
			if err != nil {
				return
			}
			response := `{"data": {"result": "success", "value": 42}}`
			c.WriteMessage(websocket.TextMessage, []byte(response))
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-output", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("output-transform"),
		}.Build()

		jqFormat := configv1.OutputTransformer_JQ
		callDef := configv1.WebsocketCallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format:  &jqFormat,
				JqQuery: stringPtrWs(".data.value"),
			}.Build(),
		}.Build()

		wt := NewWebsocketTool(toolDef, m, "service-output", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)

		assert.Equal(t, float64(42), res)
	})

	t.Run("Output Transformation - Regex", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			c.ReadMessage()
			c.WriteMessage(websocket.TextMessage, []byte(`raw text response: 12345`))
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-regex", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("regex-transform"),
		}.Build()

		format := configv1.OutputTransformer_TEXT
		callDef := configv1.WebsocketCallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format: &format,
				ExtractionRules: map[string]string{
					"extracted": "response: (\\d+)",
				},
			}.Build(),
		}.Build()

		wt := NewWebsocketTool(toolDef, m, "service-regex", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)

		resMap, ok := res.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "12345", resMap["extracted"])
	})

	t.Run("Input Template Error", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			c.ReadMessage()
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		// Need a registered pool so we don't fail on "no pool found"
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, _ := pool.New(factory, 0, 1, 1, 0, false)
		m.Register("service", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{
			Name: stringPtrWs("template-error"),
		}.Build()
		callDef := configv1.WebsocketCallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: stringPtrWs(`{{.invalid`),
			}.Build(),
		}.Build()

		wt := NewWebsocketTool(toolDef, m, "service", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		_, err := wt.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create input template")
	})

	t.Run("Write Message Error", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			c.ReadMessage()
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			// Close connection immediately to cause write error
			conn.Close()
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-write-error", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{Name: stringPtrWs("write-error")}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "service-write-error", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		_, err = wt.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send message")
	})

	t.Run("Read Message Error", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			c.ReadMessage()
			// Close immediately without writing response
			c.Close()
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-read-error", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{Name: stringPtrWs("read-error")}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "service-read-error", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		_, err = wt.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read message")
	})

	t.Run("Non-JSON Response", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {
			c.ReadMessage()
			c.WriteMessage(websocket.TextMessage, []byte("plain text"))
		})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-non-json", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{Name: stringPtrWs("non-json")}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "service-non-json", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{}`),
		}

		res, err := wt.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "plain text", res)
	})

	t.Run("Bad Tool Input", func(t *testing.T) {
		server := startTestWebsocketServer(t, func(c *websocket.Conn) {})
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		m := pool.NewManager()
		factory := func(ctx context.Context) (*client.WebsocketClientWrapper, error) {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				return nil, err
			}
			return &client.WebsocketClientWrapper{Conn: conn}, nil
		}
		p, err := pool.New(factory, 1, 1, 1, 0, false)
		require.NoError(t, err)
		m.Register("service-bad-input", p)
		defer m.CloseAll()

		toolDef := v1.Tool_builder{Name: stringPtrWs("bad-input")}.Build()
		callDef := &configv1.WebsocketCallDefinition{}

		wt := NewWebsocketTool(toolDef, m, "service-bad-input", &noOpAuthenticator{}, callDef)

		req := &ExecutionRequest{
			ToolInputs: []byte(`{invalid`),
		}

		_, err = wt.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})
}

func stringPtrWs(s string) *string {
	return &s
}
