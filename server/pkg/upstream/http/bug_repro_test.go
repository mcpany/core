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

func TestBugRepro_QueryParamInjection(t *testing.T) {
	// Create a test server that checks the query parameter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect q to be "hello&world" (decoded)
		// which means the raw query should have been "q=hello%26world"

		// If the bug exists, r.URL.Query() will have "q"=["hello"] and "world"=[""]
		// because "hello&world" was injected as "hello&world" instead of "hello%26world"

		q := r.URL.Query()
		if val := q.Get("q"); val != "hello&world" {
			t.Errorf("Expected q='hello&world', got '%s'", val)
		}
		if q.Has("world") {
			t.Errorf("Unexpected query param 'world'")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := httppkg.NewUpstream(pm)

	// Configure a tool with a query parameter placeholder
	configJSON := `{
		"name": "repro-service",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test?q={{val}}",
					"parameters": [
						{
							"schema": {
								"name": "val",
								"type": "STRING",
								"is_required": true
							}
						}
					]
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

	// Execute the tool with input containing '&'
	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: []byte(`{"val": "hello&world"}`),
	})
	require.NoError(t, err)
}
