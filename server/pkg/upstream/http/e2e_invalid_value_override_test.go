// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestE2E_InvalidValueOverride(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	var capturedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Base URL has invalid value for 'q'
	address := server.URL + "?q=invalid%"
	// Endpoint tries to override 'q'
	endpointPath := "/test?q=valid"

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	configJSON := `{
		"name": "e2e-override-service",
		"http_service": {
			"address": "` + address + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_POST",
					"endpoint_path": "` + endpointPath + `"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	assert.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// Execute the tool
	req := &tool.ExecutionRequest{
		ToolName:   "test-op",
		ToolInputs: []byte("{}"),
	}
	_, err = registeredTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	// Check captured query
	t.Logf("Captured query: %s", capturedQuery)
	// We expect the invalid base query param to be overridden.
	// If the bug exists, we will see duplicates.
	assert.Equal(t, "q=valid", capturedQuery)
}
