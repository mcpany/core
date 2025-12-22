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
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketTool_Execute_Success(t *testing.T) {
	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close() //nolint:errcheck
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			// Echo back
			_ = c.WriteMessage(mt, message)
		}
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer resp.Body.Close() //nolint:errcheck
		}
		if err != nil {
			return nil, err
		}
		return &client.WebsocketClientWrapper{Conn: c}, nil
	}

	p, err := pool.New(factory, 1, 1, 0, false)
	require.NoError(t, err)
	defer p.Close() //nolint:errcheck

	pm := pool.NewManager()
	pm.Register("s1", p)

	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"s1",
		nil, // Authenticator
		&configv1.WebsocketCallDefinition{},
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
	pm := pool.NewManager()
	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"missing-service",
		nil,
		&configv1.WebsocketCallDefinition{},
	)
	_, err := wt.Execute(context.Background(), &ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no websocket pool found")
}

func TestWebsocketTool_Execute_PoolGetError(t *testing.T) {
	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		return nil, fmt.Errorf("factory failed")
	}
	p, err := pool.New(factory, 0, 1, 0, true) // 0 min size, disable health check? Or just empty.
	require.NoError(t, err)

	pm := pool.NewManager()
	pm.Register("s2", p)

	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"s2",
		nil,
		&configv1.WebsocketCallDefinition{},
	)

	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get websocket connection")
}

func TestWebsocketTool_Execute_WriteError(t *testing.T) {
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
			defer resp.Body.Close() //nolint:errcheck
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, err := pool.New(factory, 1, 1, 0, true)
	require.NoError(t, err)
	defer p.Close() //nolint:errcheck

	pm := pool.NewManager()
	pm.Register("s3", p)

	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"s3",
		nil,
		&configv1.WebsocketCallDefinition{},
	)

	_, err = wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.Error(t, err)
}

func TestWebsocketTool_Execute_Transformer(t *testing.T) {
	upgrader := websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, _ := upgrader.Upgrade(w, r, nil)
		defer c.Close() //nolint:errcheck
		_, _, _ = c.ReadMessage()
		_ = c.WriteMessage(websocket.TextMessage, []byte(`raw text response`))
	}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		url := "ws" + strings.TrimPrefix(s.URL, "http")
		c, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if resp != nil {
			defer resp.Body.Close() //nolint:errcheck
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, _ := pool.New(factory, 1, 1, 0, false)
	defer p.Close() //nolint:errcheck
	pm := pool.NewManager()
	pm.Register("s4", p)

	format := configv1.OutputTransformer_TEXT
	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"s4",
		nil,
		&configv1.WebsocketCallDefinition{
			OutputTransformer: &configv1.OutputTransformer{
				Format: &format,
				ExtractionRules: map[string]string{
					"content": "(.*)",
				},
			},
		},
	)

	res, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage("{}")})
	assert.NoError(t, err)

	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "raw text response", resMap["content"])
}

func TestWebsocketTool_Execute_MalformedInputs(t *testing.T) {
	pm := pool.NewManager()
	wt := NewWebsocketTool(
		&pb.Tool{Name: proto.String("test-tool")},
		pm,
		"s1",
		nil,
		&configv1.WebsocketCallDefinition{},
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
			defer resp.Body.Close() //nolint:errcheck
		}
		return &client.WebsocketClientWrapper{Conn: c}, err
	}
	p, _ := pool.New(factory, 1, 1, 0, false)
	pm.Register("s5", p)

	wt.serviceID = "s5"
	wt.poolManager = pm

	_, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

func TestWebsocketTool_Execute_TemplateError(t *testing.T) {
	pm := pool.NewManager()
	s := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	defer s.Close()

	factory := func(_ context.Context) (*client.WebsocketClientWrapper, error) {
		return &client.WebsocketClientWrapper{}, nil
	}
	p, _ := pool.New(factory, 0, 1, 0, true)
	pm.Register("s", p)

	callDef := &configv1.WebsocketCallDefinition{
		InputTransformer: &configv1.InputTransformer{},
	}
	callDef.InputTransformer.SetTemplate("{{.invalid") // Bad template syntax

	wt := NewWebsocketTool(&pb.Tool{Name: proto.String("t")}, pm, "s", nil, callDef)

	_, err := wt.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create input template")
}
