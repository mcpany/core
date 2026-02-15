package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPUpstream_URLConstruction_SpaceBug(t *testing.T) {
	// This test verifies that endpoint paths with spaces in query parameters
	// do NOT cause tool initialization failure due to strings.Fields splitting.

	tc := struct {
		name          string
		address       string
		endpointPath  string
	}{
		name:         "endpoint with space and invalid encoding in query param",
		address:      "http://example.com/api",
		endpointPath: "/v1/test?q=hello world%",
	}

	t.Run(tc.name, func(t *testing.T) {
		pm := pool.NewManager()
		tm := tool.NewManager(nil)
		upstream := NewUpstream(pm)

		configJSON := `{
			"name": "space-bug-service",
			"http_service": {
				"address": "` + tc.address + `",
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

		// The bug is that the tool might be registered but fail initialization inside NewHTTPTool if we were checking errors there.
		// However, Upstream.Register calls tool.NewHTTPTool which returns a *HTTPTool.
		// Wait, NewHTTPTool returns *HTTPTool. It sets t.initError if it fails.
		// Upstream.Register calls toolManager.AddTool(httpTool).
		// So the tool IS registered.
		// But when we try to execute it, it will fail.

		assert.True(t, ok, "Tool should be registered")
		if !ok {
			return
		}

		// Let's check if the tool has initError by trying to Execute it?
		// Or we can check internal state if possible.
		// HTTPTool struct has initError field but it's private.
		// But Execute() returns initError first thing.


		// Mock execution request
		req := &tool.ExecutionRequest{
			ToolName:   "test-op",
			ToolInputs: []byte("{}"),
		}

		// We expect Execute to NOT return an "invalid http tool definition" error.
		// Since we don't have a real pool/client, it will fail later with "no http pool found" or connection error,
		// but we want to ensure it passes the init check.

		_, err = registeredTool.Execute(context.Background(), req)

		// If the bug exists, err will be "invalid http tool definition: expected method and URL, got ..."
		// If fixed, it should be something else (like pool error).

		if err != nil {
			assert.NotContains(t, err.Error(), "invalid http tool definition", "Tool failed initialization due to space in URL")
		}
	})
}
