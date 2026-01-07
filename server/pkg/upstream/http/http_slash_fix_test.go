// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	httppkg "github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestEncodedLeadingSlashBug(t *testing.T) {
	// Create a test server that checks the RequestURI
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to be exactly "/%2F123" relative to root, or "/api/%2F123"
		// The test config uses address ending in /api if needed, or just root.
		// Let's use address "http://localhost/api"

		expectedSuffix := "/api/%2F123"
		if r.RequestURI != expectedSuffix {
			t.Errorf("Expected RequestURI='%s', got '%s'", expectedSuffix, r.RequestURI)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := httppkg.NewUpstream(pm)

	// Configure a tool with endpoint_path starting with %2F
	configJSON := `{
		"name": "slash-service",
		"http_service": {
			"address": "` + server.URL + `/api",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "%2F123"
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// Execute the tool
	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: []byte("{}"),
	})
	require.NoError(t, err)
}
