// Copyright 2025 Author(s) of MCP Any
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

func TestHTTPUpstream_URLConstruction_RuntimeDoubleSlashRootBug(t *testing.T) {
	// Use t.Setenv for thread-safe environment variable setting
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Start a test server to capture the request
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	testCases := []struct {
		name          string
		endpointPath  string
		expectedPath  string // Path received by the server (relative to root)
	}{
		{
			name:         "endpoint path with just double slash",
			endpointPath: "//",
			expectedPath: "//", // We expect // to be preserved
		},
		{
			name:         "endpoint path with triple slash",
			endpointPath: "///",
			expectedPath: "///",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capturedPath = ""
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			upstream := NewUpstream(pm)

			// We use the test server address
			configJSON := `{
				"name": "double-slash-runtime-service",
				"http_service": {
					"address": "` + server.URL + `",
					"tools": [{"name": "test-op", "call_id": "test-op-call"}],
					"calls": {
						"test-op-call": {
							"id": "test-op-call",
							"method": "HTTP_METHOD_GET",
							"endpoint_path": "` + tc.endpointPath + `"
						}
					}
				}
			}`
			serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
			require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

			serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
			assert.NoError(t, err)

			sanitizedToolName, _ := util.SanitizeToolName("test-op")
			toolID := serviceID + "." + sanitizedToolName
			registeredTool, ok := tm.GetTool(toolID)
			assert.True(t, ok)

			// Execute the tool
			req := &tool.ExecutionRequest{
				ToolName:   "test-op",
				ToolInputs: []byte("{}"),
			}
			_, err = registeredTool.Execute(context.Background(), req)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedPath, capturedPath)
		})
	}
}
