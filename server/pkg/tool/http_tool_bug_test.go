package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Execute_MissingAndEmptyParams(t *testing.T) {
	t.Parallel()
	// Setup a server that echoes the path
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"path": "` + r.URL.Path + `"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	t.Run("missing_optional_param_slash_prefixed", func(t *testing.T) {
		// Definition: /users/{{id}}
		// Missing id -> /users/ (Trailing slash is preserved if resulting path ends with it)
		methodAndURL := "GET " + server.URL + "/users/{{id}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name:       proto.String("id"),
				IsRequired: proto.Bool(false),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "/users/", resultMap["path"])
	})

	t.Run("missing_optional_param_embedded", func(t *testing.T) {
		// Definition: /files/image-{{id}}.png
		// Missing id -> /files/image-.png
		methodAndURL := "GET " + server.URL + "/files/image-{{id}}.png"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name:       proto.String("id"),
				IsRequired: proto.Bool(false),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "/files/image-.png", resultMap["path"])
	})

	t.Run("empty_string_param_creates_double_slash", func(t *testing.T) {
		// Definition: /users/{{id}}/items
		// id="" -> /users//items (Double slash preserved)
		methodAndURL := "GET " + server.URL + "/users/{{id}}/items"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name:       proto.String("id"),
				IsRequired: proto.Bool(false),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{"id": ""}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		result, err := httpTool.Execute(context.Background(), req)
		require.NoError(t, err)
		resultMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "/users//items", resultMap["path"])
	})

	t.Run("missing_required_param", func(t *testing.T) {
		// Definition: /users/{{id}}
		// Missing id -> Error
		methodAndURL := "GET " + server.URL + "/users/{{id}}"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name:       proto.String("id"),
				IsRequired: proto.Bool(true),
			}.Build(),
		}.Build()
		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing required parameter: id")
	})
}
