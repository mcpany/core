// Copyright 2026 Author(s) of MCP Any
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

func TestHTTPTool_ExtraCoverage(t *testing.T) {
	t.Parallel()

	t.Run("path traversal in path parameter", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		methodAndURL := "GET " + server.URL + "/{{path}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("path"),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{"path": "../secret"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	})

	t.Run("double encoded path traversal", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		server := httptest.NewServer(handler)
		defer server.Close()

		poolManager := pool.NewManager()
		p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{Client: server.Client()}, nil
		}, 1, 1, 1, 0, true)
		poolManager.Register("test-service", p)

		methodAndURL := "GET " + server.URL + "/{{path}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("path"),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		// %252e%252e is double encoded ..
		inputs := json.RawMessage(`{"path": "%252e%252e"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	})

	t.Run("NewHTTPTool with invalid URL", func(t *testing.T) {
		toolProto := &v1.Tool{}
		toolProto.SetName("testTool")
		// Invalid URL (control char)
		methodAndURL := "GET http://test.com/foo\x7fbar"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		poolManager := pool.NewManager()
		p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{}, nil
		}, 1, 1, 1, 0, true)
		require.NoError(t, err)
		poolManager.Register("test-service", p)

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

		_, err = httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse url")
	})

	t.Run("Secret resolution failure", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
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

		// Secret pointing to non-existent env var
		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{Name: proto.String("token")}.Build(),
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("NON_EXISTENT_VAR_12345"),
			}.Build(),
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR_12345\" is not set")
	})
}
