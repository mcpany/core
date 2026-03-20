// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// --- HTTPTool Tests ---

func TestHTTPTool_Execute_Success(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer s.Close()

	pm := pool.NewManager()
	factory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: s.Client()}, nil
	}
	p, _ := pool.New(factory, 1, 1, 1, 0, false)
	pm.Register("http-service", p)

	toolDef := pb.Tool_builder{
		Name:                proto.String("test-http"),
		UnderlyingMethodFqn: proto.String(fmt.Sprintf("GET %s/api", s.URL)),
	}.Build()

	ht := NewHTTPTool(
		toolDef,
		pm,
		"http-service",
		nil,
		&configv1.HttpCallDefinition{},
		nil, // Resilience
		nil, // Policies
		"",  // CallID
	)

	res, err := ht.Execute(context.Background(), &ExecutionRequest{
		ToolName:   "test-http",
		ToolInputs: json.RawMessage(`{}`),
	})
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "ok", resMap["status"])
}

func TestHTTPTool_Execute_Post_WithBody(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["foo"] == "bar" {
				_, _ = w.Write([]byte(`{"accepted":true}`))
			} else {
				w.WriteHeader(400)
			}
		}
	}))
	defer s.Close()

	pm := pool.NewManager()
	factory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: s.Client()}, nil
	}
	p, _ := pool.New(factory, 1, 1, 1, 0, false)
	pm.Register("http-service", p)

	toolDef := pb.Tool_builder{
		Name:                proto.String("test-post"),
		UnderlyingMethodFqn: proto.String(fmt.Sprintf("POST %s/resource", s.URL)),
	}.Build()

	ht := NewHTTPTool(
		toolDef,
		pm,
		"http-service",
		nil,
		configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{
				configv1.HttpParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name: proto.String("foo"),
					}.Build(),
				}.Build(),
			},
		}.Build(),
		nil,
		nil,
		"",
	)

	req := &ExecutionRequest{
		ToolName:   "test-post",
		ToolInputs: json.RawMessage(`{"foo":"bar"}`),
	}
	res, err := ht.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, resMap["accepted"])
}

func TestHTTPTool_Execute_Auth(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer token" {
			_, _ = w.Write([]byte(`{"authed":true}`))
		} else {
			w.WriteHeader(401)
		}
	}))
	defer s.Close()

	pm := pool.NewManager()
	factory := func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: s.Client()}, nil
	}
	p, _ := pool.New(factory, 1, 1, 1, 0, false)
	pm.Register("s", p)

	mockAuth := &MockTypesAuthenticator{
		AuthenticateFunc: func(r *http.Request) error {
			r.Header.Set("Authorization", "Bearer token")
			return nil
		},
	}

	ht := NewHTTPTool(
		pb.Tool_builder{Name: proto.String("auth-tool"), UnderlyingMethodFqn: proto.String("GET " + s.URL)}.Build(),
		pm,
		"s",
		mockAuth,
		&configv1.HttpCallDefinition{},
		nil,
		nil,
		"",
	)

	res, err := ht.Execute(context.Background(), &ExecutionRequest{ToolInputs: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	resMap := res.(map[string]any)
	assert.True(t, resMap["authed"].(bool))
}

type MockTypesAuthenticator struct {
	AuthenticateFunc func(r *http.Request) error
}

func (m *MockTypesAuthenticator) Authenticate(r *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(r)
	}
	return nil
}

// --- MCPTool Tests ---

type MockTypesMCPClient struct {
	CallToolFunc func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error)
}

func (m *MockTypesMCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	if m.CallToolFunc != nil {
		return m.CallToolFunc(ctx, params)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockTypesMCPClient) ReadResource(_ context.Context, _ *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) ListTools(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) ListResources(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) ListPrompts(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) GetPrompt(_ context.Context, _ *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) Initialize(_ context.Context, _ *mcp.InitializeParams) (*mcp.InitializeResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (m *MockTypesMCPClient) Close() error                   { return nil }
func (m *MockTypesMCPClient) Ping(_ context.Context) error { return nil }

func TestMCPTool_Execute(t *testing.T) {
	t.Parallel()
	mockClient := &MockTypesMCPClient{
		CallToolFunc: func(_ context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			assert.Equal(t, "remote-tool", params.Name)
			// Verify arguments passed correctly
			args := make(map[string]any)

			// params.Arguments is 'any'. In mock, it receives what we sent.
			// In Execute, we sent `arguments = req.ToolInputs` (which is json.RawMessage).
			// So it should be castable to json.RawMessage ([]byte).
			match, ok := params.Arguments.(json.RawMessage)
			if !ok {
				// Maybe SDK transformed it?
				// But we are mocking the client, so we receive exactly what was passed to CallTool.
				// Wait, if SDK interface defines Arguments as `any`, Go doesn't auto-convert.
				// However, maybe json.RawMessage IS castable to []byte directly?
				// json.RawMessage is `type RawMessage []byte`.
				// So `params.Arguments.([]byte)` might fail if it's `json.RawMessage` type?
				// Need to cast to `json.RawMessage`.
				require.Fail(t, fmt.Sprintf("Expected json.RawMessage, got %T", params.Arguments))
			}
			_ = json.Unmarshal(match, &args)
			assert.Equal(t, "val", args["key"])

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: `{"result":"success"}`,
					},
				},
			}, nil
		},
	}

	mt := NewMCPTool(
		pb.Tool_builder{Name: proto.String("remote-tool")}.Build(),
		mockClient,
		configv1.MCPCallDefinition_builder{}.Build(),
	)

	req := &ExecutionRequest{
		ToolName:   "ignored-local-name",
		ToolInputs: json.RawMessage(`{"key":"val"}`),
	}
	res, err := mt.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "success", resMap["result"])
}

func TestMCPTool_Execute_Errors(t *testing.T) {
	t.Parallel()
	mockClient := &MockTypesMCPClient{
		CallToolFunc: func(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return nil, fmt.Errorf("remote error")
		},
	}
	mt := NewMCPTool(pb.Tool_builder{Name: proto.String("err")}.Build(), mockClient, configv1.MCPCallDefinition_builder{}.Build())
	req := &ExecutionRequest{ToolName: "err", ToolInputs: json.RawMessage(`{}`)}
	_, err := mt.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remote error")
}

// --- OpenAPITool Tests ---

func TestOpenAPITool_Execute_Success(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/items/123" {
			_, _ = w.Write([]byte(`{"id":"123","name":"item"}`))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer s.Close()

	client := &client.HTTPClientWrapper{Client: s.Client()}

	toolDef := pb.Tool_builder{Name: proto.String("openapi-tool")}.Build()
	callDef := configv1.OpenAPICallDefinition_builder{}.Build()

	ot := NewOpenAPITool(
		toolDef,
		client,
		map[string]string{"id": "path"},
		"GET",
		s.URL+"/items/{{id}}",
		nil,
		callDef,
	)

	req := &ExecutionRequest{
		ToolName:   "openapi-tool",
		ToolInputs: json.RawMessage(`{"id": 123}`),
	}
	res, err := ot.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "item", resMap["name"])
}

func TestOpenAPITool_Execute_QueryParam(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") == "search" {
			_, _ = w.Write([]byte(`{"found":true}`))
		} else {
			w.WriteHeader(400)
		}
	}))
	defer s.Close()

	client := &client.HTTPClientWrapper{Client: s.Client()}
	ot := NewOpenAPITool(
		pb.Tool_builder{Name: proto.String("search")}.Build(),
		client,
		map[string]string{"q": "query"},
		"GET",
		s.URL+"/search",
		nil,
		configv1.OpenAPICallDefinition_builder{}.Build(),
	)

	req := &ExecutionRequest{
		ToolName:   "search",
		ToolInputs: json.RawMessage(`{"q": "search"}`),
	}
	res, err := ot.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap, ok := res.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, true, resMap["found"])
}
