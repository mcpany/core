// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	httppkg "github.com/mcpany/core/pkg/upstream/http"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestSSRFProtection(t *testing.T) {
	// 1. Setup a local server (acts as internal/private resource)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Helper to execute tool
	executeTool := func(t *testing.T, allowLoopback string) error {
		if allowLoopback != "" {
			os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", allowLoopback)
		} else {
			os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		}
		// Ensure cleanup
		defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

		pm := pool.NewManager()
		defer pm.CloseAll()
		tm := tool.NewManager(nil)
		upstream := httppkg.NewUpstream(pm)

		configJSON := `{
			"name": "ssrf-service",
			"http_service": {
				"address": "` + server.URL + `",
				"tools": [{"name": "status", "call_id": "status-call"}],
				"calls": {
					"status-call": {
						"id": "status-call",
						"method": "HTTP_METHOD_GET",
						"endpoint_path": "/"
					}
				}
			}
		}`
		serviceConfig := &configv1.UpstreamServiceConfig{}
		require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
		if err != nil {
			return err
		}

		sanitizedToolName, _ := util.SanitizeToolName("status")
		toolID := serviceID + "." + sanitizedToolName
		registeredTool, ok := tm.GetTool(toolID)
		if !ok {
			return fmt.Errorf("tool not found")
		}

		_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
			ToolName:   toolID,
			ToolInputs: []byte(`{}`),
		})
		return err
	}

	// Case 1: Default behavior (Should BLOCK)
	t.Run("Default blocks localhost", func(t *testing.T) {
		err := executeTool(t, "")
		require.Error(t, err, "Connection to localhost should be blocked by default")
		require.True(t, strings.Contains(err.Error(), "ssrf attempt blocked") || strings.Contains(err.Error(), "loopback"), "Error should indicate SSRF block")
	})

	// Case 2: Opt-in allow
	t.Run("Opt-in allows localhost", func(t *testing.T) {
		err := executeTool(t, "true")
		require.NoError(t, err, "Connection to localhost should succeed when opted in")
	})
}
