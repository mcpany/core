// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_PathTraversal_Encoded_Vulnerability(t *testing.T) {
	// Handler checks where the request lands
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the path traversal succeeds, the request might land on /secret
		// or /files/%2e%2e%2fsecret depending on how the client sends it.
		// If the client normalizes it, it lands on /secret.
		if r.URL.Path == "/secret" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status": "pwned"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status": "not found"}`))
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

	methodAndURL := "GET " + server.URL + "/files/{{file}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
		Name:                proto.String("test-tool"),
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
		DisableEscape: proto.Bool(true), // Disable escape to allow injection
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Encoded ../secret
	inputs := json.RawMessage(`{"file": "%2e%2e%2fsecret"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)

	// This should fail with an error, but if the bug exists, it will succeed (err == nil)
	require.Error(t, err, "Expected path traversal error for encoded input")
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}
