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

func TestHTTPTool_Execute_Injection_DisableEscape(t *testing.T) {
	// Bypass safe URL check for test
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Handler that checks if "admin=true" was injected as a query parameter
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If injection worked, "admin" will be in the query
		if r.URL.Query().Get("admin") == "true" {
			// This represents a successful exploit
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "exploited"}`))
		} else {
			// Normal behavior
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "safe"}`))
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/data/{{resource}}"
	// Use struct literal instead of Builder if Builder is not available/working,
    // but the existing tests use Builder. I'll try to match exact import style.
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("resource"),
		}.Build(),
		DisableEscape: proto.Bool(true), // DANGER!
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Attempt injection
	// We pass "foo?admin=true".
	// If DisableEscape is true, path becomes ".../data/foo?admin=true"
	inputs := json.RawMessage(`{"resource": "foo?admin=true"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)

	// We expect an error because injection should be blocked
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter \"resource\": contains forbidden characters")
}
