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

func TestHTTPTool_ParameterPollution_Repro(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if "admin" parameter exists and equals "true"
		adminVal := r.URL.Query().Get("admin")
		if adminVal == "true" {
			// This means pollution was successful
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"polluted": true}`))
			return
		}

		// Normal response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"polluted": false}`))
	}))
	defer server.Close()

	// 2. Setup Upstream and Tool Manager
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// 3. Define Config with DisableEscape: true
	configJSON := `{
		"name": "pollution-test",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [
				{"name": "search", "call_id": "search-call"}
			],
			"calls": {
				"search-call": {
					"id": "search-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/search?q={{query}}",
					"parameters": [
						{
							"schema": {"name": "query"},
							"disable_escape": true
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// 4. Register
	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// 5. Get Tool
	sanitizedToolName, _ := util.SanitizeToolName("search")
	toolID := serviceID + "." + sanitizedToolName
	searchTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// 6. Execute with Malicious Input
	// Payload: "test&admin=true"
	req := &tool.ExecutionRequest{
		ToolName: "search",
		ToolInputs: []byte(`{"query": "test&admin=true"}`),
	}

	_, err = searchTool.Execute(context.Background(), req)

	// 7. Check Result - Expect Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter pollution detected")
}

func TestHTTPTool_PathInjection(t *testing.T) {
	// 1. Setup Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// 2. Setup Upstream and Tool Manager
	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := NewUpstream(pm)

	// 3. Define Config with DisableEscape: true for path param
	configJSON := `{
		"name": "path-injection-test",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [
				{"name": "get-user", "call_id": "get-user-call"}
			],
			"calls": {
				"get-user-call": {
					"id": "get-user-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/users/{{id}}",
					"parameters": [
						{
							"schema": {"name": "id"},
							"disable_escape": true
						}
					]
				}
			}
		}
	}`
	serviceConfig := &configv1.UpstreamServiceConfig{}
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	// 4. Register
	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	// 5. Get Tool
	sanitizedToolName, _ := util.SanitizeToolName("get-user")
	toolID := serviceID + "." + sanitizedToolName
	userTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// 6. Execute with Malicious Input
	// Payload: "1?admin=true"
	req := &tool.ExecutionRequest{
		ToolName: "get-user",
		ToolInputs: []byte(`{"id": "1?admin=true"}`),
	}

	_, err = userTool.Execute(context.Background(), req)

	// 7. Check Result - Expect Error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path injection detected")
}
