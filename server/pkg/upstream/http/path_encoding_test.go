// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	httppkg "github.com/mcpany/core/pkg/upstream/http"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPathEncodingBug(t *testing.T) {
	// Create a test server that checks the RequestURI
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to be exactly "/users/user%2F1"
		// If the bug exists, it will likely be "/users/user/1" (double slash, or just decoded)

		expectedURI := "/users/user%2F1"
		if r.RequestURI != expectedURI {
			t.Errorf("Expected RequestURI='%s', got '%s'", expectedURI, r.RequestURI)
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

	// Configure a tool with a path parameter containing encoded slash (%2F)
	configJSON := `{
		"name": "encoding-service",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/users/user%2F1"
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
    // If status code is 400, Execute might return error or not depending on implementation.
    // Assuming tool returns error on 400.
	require.NoError(t, err)
}
