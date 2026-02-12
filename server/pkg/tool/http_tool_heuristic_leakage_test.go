// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestHTTPTool_Execute_HeuristicRedaction verifies that secrets passed as query parameters
// are redacted in the logs even if they are NOT explicitly configured as secrets,
// as long as they match the heuristic sensitive keys (e.g. "api_key").
func TestHTTPTool_Execute_HeuristicRedaction(t *testing.T) {
	// Allow local IPs for testing
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Reset logger for this test
	logging.ForTestsOnlyResetLogger()
	defer logging.ForTestsOnlyResetLogger()

	// Capture logs
	var logBuf bytes.Buffer
	logging.Init(slog.LevelDebug, &logBuf, "")

	// Setup upstream that returns error (500) to trigger logging
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Setup Pool
	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	// Define Tool with Secret Parameter in Query using Placeholder.
	// IMPORTANT: We do NOT configure it as a Secret in the mapping.
	methodAndURL := "GET " + server.URL + "?api_key={{api_key}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("api_key"),
		}.Build(),
		// Secret: is NOT set.
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Method:     configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Execute
	inputs := json.RawMessage(`{"api_key": "super_secret_value"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = httpTool.Execute(context.Background(), req)
	require.Error(t, err)

	// Verify logs
	logOutput := logBuf.String()

	// The log should contain the error log with the URL
	if assert.Contains(t, logOutput, "Upstream HTTP error", "Should log upstream error") {
		// BEFORE FIX: This assertion is expected to FAIL (finding the secret)
		// AFTER FIX: This assertion should PASS.
		assert.NotContains(t, logOutput, "super_secret_value", "Log should NOT contain sensitive value even if not configured as secret")
		// Check that it IS redacted
		assert.Contains(t, logOutput, "api_key=%5BREDACTED%5D", "Log should contain redacted secret")
	} else {
		t.Logf("Full Log Output:\n%s", logOutput)
	}
}
