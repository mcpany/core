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

func TestHTTPTool_Execute_Coverage_Secrets(t *testing.T) {
	t.Parallel()

	// Test secret resolution error
	t.Run("secret_resolution_error", func(t *testing.T) {
		poolManager := pool.NewManager()
		// No need for real backend as it should fail before request
		p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{}, nil
		}, 1, 1, 1, 0, true)
		require.NoError(t, errPool)
		poolManager.Register("test-service", p)

		methodAndURL := "GET http://example.com"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		// Parameter backed by missing env var
		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("token"),
			}.Build(),
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("NON_EXISTENT_ENV_VAR_FOR_TESTING_12345"),
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
		assert.Contains(t, err.Error(), "failed to resolve secret")
	})
}

func TestHTTPTool_Execute_Coverage_PathTraversal(t *testing.T) {
	t.Parallel()

	poolManager := pool.NewManager()
	p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, errPool)
	poolManager.Register("test-service", p)

	methodAndURL := "GET http://example.com/{{file}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Test encoded path traversal: %2e%2e -> ..
	inputs := json.RawMessage(`{"file": "%2e%2e"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestHTTPTool_Execute_Coverage_CallPolicy(t *testing.T) {
	t.Parallel()

	poolManager := pool.NewManager()
	p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, errPool)
	poolManager.Register("test-service", p)

	methodAndURL := "GET http://example.com"
	mcpTool := v1.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	// Policy that denies the tool by name
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				NameRegex: proto.String("test-tool"),
				Action:    configv1.CallPolicy_DENY.Enum(),
			}.Build(),
		},
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{}.Build()

	// NewHTTPTool accepts policies
	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, []*configv1.CallPolicy{policy}, "")

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}

	_, err := httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution blocked by policy")
}

func TestHTTPTool_Execute_Coverage_DryRun(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	methodAndURL := "POST " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("key"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"key": "value"}`)
	req := &tool.ExecutionRequest{
		ToolInputs: inputs,
		DryRun:     true,
		ToolName:   "test-tool",
	}

	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.True(t, resultMap["dry_run"].(bool))
	reqMap := resultMap["request"].(map[string]any)
	assert.Equal(t, "POST", reqMap["method"])

	// Check body
	body, ok := reqMap["body"].(string)
	if ok {
		assert.Contains(t, body, "key")
	}
}
