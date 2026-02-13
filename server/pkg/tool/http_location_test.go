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

import (
	"os"
)

func TestHTTPTool_Execute_ExplicitParameterLocation(t *testing.T) {
	// Not parallel because it modifies environment variables
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	defer os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Header
		assert.Equal(t, "header-val", r.Header.Get("X-Test-Header"))

		// Verify Cookie
		cookie, err := r.Cookie("test-cookie")
		require.NoError(t, err)
		assert.Equal(t, "cookie-val", cookie.Value)

		// Verify Explicit Query (even for POST if needed, but here testing GET)
		assert.Equal(t, "query-val", r.URL.Query().Get("q_param"))

		// Verify Implicit Path
		assert.Contains(t, r.URL.Path, "/test/path-val")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/test/{{p_param}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	params := []*configv1.HttpParameterMapping{
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("h_param")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_HEADER.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("c_param")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_COOKIE.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("q_param")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_QUERY.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("p_param")}.Build(),
			// Location implicit (path template)
		}.Build(),
	}

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: params,
	}.Build()

	// Update params with the correct name mappings to header/cookie names
	// Wait, the logic in types.go uses the parameter name as the key for header/cookie/query param.
	// h_param -> Header: h_param.
	// But usually we want to map "apiKey" to "X-Api-Key".
	// The proto `HttpParameterMapping` doesn't have a "targetName" field?
	// `UpstreamAuth` has `paramName`.
	// `HttpParameterMapping` ONLY has `Schema` and `Secret`. `Schema` has `Name`.
	// So `h_param` will be the header name.
	// If I want to map `h_param` input to `X-Test-Header` header, I MUST name the input `X-Test-Header`?
	// That's ugly for the LLM. "Please provide X-Test-Header".
	//
	// This reveals another Product Gap! We need `target_name` or similar in `HttpParameterMapping`?
	// Or maybe we abuse `description`? No.
	//
	// For now, I will test with the assumption that Input Name = Header Name.
	// But for the test, I used "X-Test-Header" in assertion.
	// So I should name the parameter "X-Test-Header".

	// Re-building params with correct names for the test expectations
	params = []*configv1.HttpParameterMapping{
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("X-Test-Header")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_HEADER.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("test-cookie")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_COOKIE.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("q_param")}.Build(),
			Location: configv1.ParameterLocation_PARAMETER_LOCATION_QUERY.Enum(),
		}.Build(),
		configv1.HttpParameterMapping_builder{
			Schema:   configv1.ParameterSchema_builder{Name: proto.String("p_param")}.Build(),
		}.Build(),
	}
	callDef.SetParameters(params)

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{
		"X-Test-Header": "header-val",
		"test-cookie": "cookie-val",
		"q_param": "query-val",
		"p_param": "path-val"
	}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
