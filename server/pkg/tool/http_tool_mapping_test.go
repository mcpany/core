// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Execute_HeaderMapping(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "user-123", r.Header.Get("X-User-ID"))
		assert.Equal(t, "secret-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	headerLoc := configv1.ParameterLocation_PARAMETER_LOCATION_HEADER

	// Mapping 1: Custom Header Name
	param1 := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{Name: proto.String("userId")}.Build(),
		Location: &headerLoc,
		TargetName: proto.String("X-User-ID"),
	}.Build()

	// Mapping 2: Default Header Name (same as param)
	// But let's test specific override.
	// Actually let's test mapping 'authToken' to 'Authorization'
	param2 := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{Name: proto.String("authToken")}.Build(),
		Location: &headerLoc,
		TargetName: proto.String("Authorization"),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{param1, param2},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"userId": "user-123", "authToken": "secret-token"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_CookieMapping(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		require.NoError(t, err)
		assert.Equal(t, "sess-abc", cookie.Value)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	cookieLoc := configv1.ParameterLocation_PARAMETER_LOCATION_COOKIE

	param1 := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{Name: proto.String("session")}.Build(),
		Location: &cookieLoc,
		TargetName: proto.String("session_id"),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{param1},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"session": "sess-abc"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_ExplicitQueryMapping(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bar", r.URL.Query().Get("foo"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	// URL does NOT contain {{foo}}
	methodAndURL := "GET " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	queryLoc := configv1.ParameterLocation_PARAMETER_LOCATION_QUERY

	param1 := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{Name: proto.String("myParam")}.Build(),
		Location: &queryLoc,
		TargetName: proto.String("foo"),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{param1},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"myParam": "bar"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
