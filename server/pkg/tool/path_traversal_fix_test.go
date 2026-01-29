package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_PathTraversal_Blocked(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/api/v1/users/{{id}}"
	mcpTool := mcp_router_v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("id"),
		}.Build(),
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Attack 1: ../admin
	inputs := json.RawMessage(`{"id": "../admin"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")

	// Attack 2: ../../secret
	inputs = json.RawMessage(`{"id": "../../secret"}`)
	req = &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")

	// Attack 3: %2e%2e
	inputs = json.RawMessage(`{"id": "%2e%2e/admin"}`)
	req = &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")

	// Valid input
	inputs = json.RawMessage(`{"id": "123"}`)
	req = &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestLocalCommandTool_ShellInjection_Prevention_NewCommands(t *testing.T) {
	// Tests that newly added commands (busybox, expect, git) are treated as shells
	t.Parallel()

	commands := []string{"busybox", "expect", "git"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			toolDef := mcp_router_v1.Tool_builder{
				Name: proto.String("test-tool-" + cmd),
			}.Build()
			service := configv1.CommandLineUpstreamService_builder{
				Command: proto.String(cmd),
				Local:   proto.Bool(true),
			}.Build()
			callDef := configv1.CommandLineCallDefinition_builder{
				Parameters: []*configv1.CommandLineParameterMapping{
					configv1.CommandLineParameterMapping_builder{
						Schema: configv1.ParameterSchema_builder{
							Name: proto.String("arg"),
						}.Build(),
					}.Build(),
				},
				Args: []string{"{{arg}}"},
			}.Build()
			localTool := tool.NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

			// Unquoted injection attempt
			inputs := json.RawMessage(`{"arg": "foo; echo injected"}`)
			req := &tool.ExecutionRequest{ToolInputs: inputs}
			_, err := localTool.Execute(context.Background(), req)
			assert.Error(t, err)
			if err != nil {
				assert.Contains(t, err.Error(), "shell injection detected")
			}
		})
	}
}
