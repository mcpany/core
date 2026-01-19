// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_SecretInjection_Protection(t *testing.T) {
	// Cannot run parallel because we modify environment variables
	// t.Parallel()

	// Allow localhost for this test as httptest.NewServer uses 127.0.0.1
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	defer os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	// This test verifies that secrets ARE properly escaped when injected into URLs.

	injectedParam := "injected=true"
	// The secret value contains characters that should be escaped in a URL query
	secretValue := "value&" + injectedParam

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify "injected" is NOT parsed as a separate query parameter
		if r.URL.Query().Get("injected") == "true" {
			t.Errorf("Vulnerability detected: 'injected' parameter found in query: %s", r.URL.RawQuery)
		}

		// Verify the key parameter contains the full secret value (meaning & was escaped)
		keyVal := r.URL.Query().Get("key")
		require.Equal(t, secretValue, keyVal, "Secret should be preserved intact as a single parameter value")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	// Tool definition: GET /api?key={{apiKey}}
	methodAndURL := "GET " + server.URL + "/api?key={{apiKey}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	// Parameter mapping with a Secret
	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("apiKey"),
		}.Build(),
		Secret: configv1.SecretValue_builder{
			PlainText: proto.String(secretValue),
		}.Build(),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Execute
	req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
