// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_E2E_QueryOverride(t *testing.T) {
	// 0. Enable loopback resources for test
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// 1. Start mock server
	receivedQuery := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// 2. Setup Upstream with invalid base query param
	pm := pool.NewManager()
	tm := tool.NewManager(nil)

	upstream := NewUpstream(pm)

	// address has ?foo=bad% (invalid encoding)
	// endpoint has ?foo=good
	// We expect foo=good to win and foo=bad% to be gone.
	configJSON := `{
        "name": "query-override-service",
        "http_service": {
            "address": "` + server.URL + `?foo=bad%",
            "tools": [{"name": "test-op", "call_id": "test-op-call"}],
            "calls": {
                "test-op-call": {
                    "id": "test-op-call",
                    "method": "HTTP_METHOD_GET",
                    "endpoint_path": "/test?foo=good"
                }
            }
        }
    }`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// 3. Register
	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	// 4. Get Tool
	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName

	httpTool, ok := tm.GetTool(toolID)
	require.True(t, ok, "Tool should be registered")

	// 5. Execute
	req := &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	// 6. Verify
	// The expected query should contain foo=good and NOT foo=bad%
	assert.Contains(t, receivedQuery, "foo=good")
	assert.NotContains(t, receivedQuery, "bad%")
}
