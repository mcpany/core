package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketTool_Execute_Success(t *testing.T) {
	t.Parallel()
	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() {
			if err := c.Close(); err != nil {
				t.Logf("failed to close connection: %v", err)
			}
		}()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Echo back
			if err := c.WriteMessage(mt, message); err != nil {
				t.Logf("failed to write message: %v", err)
				break
			}
		}
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("failed to close response body: %v", err)
				}
			}()
		}
		if err != nil {
			return nil, err
		}
		return &client.WebsocketClientWrapper{Conn: c}, nil
	}

	p, err := pool.New(factory, 1, 1, 1, 0, false)
	require.NoError(t, err)
	defer func() {
		if err := p.Close(); err != nil {
			t.Logf("failed to close pool: %v", err)
		}
	}()

	pm := pool.NewManager()
	pm.Register("s1", p)

	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"s1",
		nil, // Authenticator
		configv1.WebsocketCallDefinition_builder{}.Build(),
	)

	req := &ExecutionRequest{
		ToolInputs: json.RawMessage(`{"foo":"bar"}`),
	}
	res, err := wt.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "bar", resMap["foo"])
}

func TestWebsocketTool_Execute_NoPool(t *testing.T) {
	t.Parallel()
	pm := pool.NewManager()
	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"missing-service",
		nil,
		configv1.WebsocketCallDefinition_builder{}.Build(),
	)
	_, err := wt.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no websocket pool found")
}

func TestWebsocketTool_Execute_PoolGetError(t *testing.T) {
	t.Parallel()
	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		return nil, fmt.Errorf("factory failed")
	}
	p, err := pool.New(factory, 0, 0, 1, 0, true) // 0 min size, disable health check? Or just empty.
	require.NoError(t, err)

	pm := pool.NewManager()
	pm.Register("s2", p)

	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"s2",
		nil,
		configv1.WebsocketCallDefinition_builder{}.Build(),
	)

	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get websocket connection")
}

func TestWebsocketTool_Execute_WriteError(t *testing.T) {
	t.Parallel()
	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err == nil {
			_ = c.Close()
		}
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, err := pool.New(factory, 1, 1, 1, 0, true)
	require.NoError(t, err)
	defer func() {
		_ = p.Close()
	}()

	pm := pool.NewManager()
	pm.Register("s3", p)

	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"s3",
		nil,
		configv1.WebsocketCallDefinition_builder{}.Build(),
	)

	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.Error(t, err)
}

func TestWebsocketTool_Execute_Transformer(t *testing.T) {
	t.Parallel()
	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := upgrader.Upgrade(w, r, nil)
		defer func() {
			_ = c.Close()
		}()
		_, _, _ = c.ReadMessage()
		_ = c.WriteMessage(websocket.TextMessage, []byte(`raw text response`))
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, _ := pool.New(factory, 1, 1, 1, 0, false)
	defer func() {
		_ = p.Close()
	}()
	pm := pool.NewManager()
	pm.Register("s4", p)

	format := configv1.OutputTransformer_TEXT
	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"s4",
		nil,
		configv1.WebsocketCallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format: &format,
				ExtractionRules: map[string]string{
					"content": "(.*)",
				},
			}.Build(),
		}.Build(),
	)

	res, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.NoError(t, err)

	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "raw text response", resMap["content"])
}

func TestWebsocketTool_Execute_MalformedInputs(t *testing.T) {
	t.Parallel()
	pm := pool.NewManager()
	wt := NewWebsocketTool(
		pb.Tool_builder{Name: proto.String("test-tool")}.Build(),
		pm,
		"s1",
		nil,
		configv1.WebsocketCallDefinition_builder{}.Build(),
	)

	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := upgrader.Upgrade(w, r, nil)
		_ = c.Close()
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer func() {
				_ = resp.Body.Close()
			}()
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, _ := pool.New(factory, 1, 1, 1, 0, false)
	pm.Register("s5", p)

	wt.serviceID = "s5"
	wt.poolManager = pm

	_, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

func TestWebsocketTool_Execute_TemplateError(t *testing.T) {
	t.Parallel()
	pm := pool.NewManager()
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		return &client.WebsocketClientWrapper{}, nil
	}
	p, _ := pool.New(factory, 0, 0, 1, 0, true)
	pm.Register("s", p)

	callDef := configv1.WebsocketCallDefinition_builder{
		InputTransformer: configv1.InputTransformer_builder{}.Build(),
	}.Build()
	callDef.GetInputTransformer().SetTemplate("{{.invalid") // Bad template syntax

	wt := NewWebsocketTool(pb.Tool_builder{Name: proto.String("t")}.Build(), pm, "s", nil, callDef)

	_, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create input template")
}
