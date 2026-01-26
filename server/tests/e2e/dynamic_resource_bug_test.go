// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicResourceBug_E2E(t *testing.T) {
	// 1. Start Mock HTTP Server that returns primitives
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/int":
			w.Write([]byte("123"))
		case "/float":
			w.Write([]byte("12.34"))
		case "/bool":
			w.Write([]byte("true"))
		case "/slice":
			w.Write([]byte(`["a", 1]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	// 2. Configure MCP Any
	configContent := fmt.Sprintf(`
global_settings:
  mcp_listen_address: "127.0.0.1:0"
upstream_services:
  - name: "mock-primitive"
    http_service:
      address: "%s"
      tools:
        - name: "get-int-tool"
          call_id: "get-int"
        - name: "get-float-tool"
          call_id: "get-float"
        - name: "get-bool-tool"
          call_id: "get-bool"
        - name: "get-slice-tool"
          call_id: "get-slice"
      calls:
        get-int:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/int"
        get-float:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/float"
        get-bool:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/bool"
        get-slice:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/slice"
      resources:
        - uri: "primitive://int"
          name: "int-resource"
          dynamic:
            http_call:
              id: "get-int"
        - uri: "primitive://float"
          name: "float-resource"
          dynamic:
            http_call:
              id: "get-float"
        - uri: "primitive://bool"
          name: "bool-resource"
          dynamic:
            http_call:
              id: "get-bool"
        - uri: "primitive://slice"
          name: "slice-resource"
          dynamic:
            http_call:
              id: "get-slice"
`, mockServer.URL)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// 3. Start Application with intercepted Server
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	defer os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	appRunner := app.NewApplication()

	serverCh := make(chan *mcpserver.Server, 1)

	// Override runStdioModeFunc to capture the server instance
	appRunner.SetRunStdioModeFunc(func(ctx context.Context, mcpSrv *mcpserver.Server) error {
		serverCh <- mcpSrv
		<-ctx.Done() // Block until context canceled
		return nil
	})

	go func() {
		fs := afero.NewOsFs()
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           true, // Use Stdio mode to trigger our override
			ConfigPaths:     []string{tmpFile.Name()},
			ShutdownTimeout: 5 * time.Second,
		}
		if err := appRunner.Run(opts); err != nil && err != context.Canceled {
			t.Logf("Application run error: %v", err)
		}
	}()

	// Wait for server capture
	var mcpSrv *mcpserver.Server
	select {
	case mcpSrv = <-serverCh:
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for server startup")
	}

	// Wait for resource registration
	require.Eventually(t, func() bool {
		res, err := mcpSrv.ListResources(ctx, &mcp.ListResourcesRequest{})
		if err != nil {
			return false
		}
		for _, r := range res.Resources {
			if r.URI == "primitive://int" {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "Timed out waiting for resource registration")

	// 4. Test reading resources
	testCases := []struct {
		uri      string
		expected string
	}{
		{"primitive://int", "123"},
		{"primitive://float", "12.34"},
		{"primitive://bool", "true"},
		{"primitive://slice", `["a",1]`},
	}

	for _, tc := range testCases {
		t.Run(tc.uri, func(t *testing.T) {
			req := &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: tc.uri,
				},
			}

			// Call ReadResource directly on the server
			// Note: We skip the router for simplicity, calling the method implementation directly if exposed
			// mcpserver.Server exposes ReadResource method.

			res, err := mcpSrv.ReadResource(ctx, req)

			// Fails here currently
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Len(t, res.Contents, 1)
			assert.Equal(t, tc.expected, res.Contents[0].Text)
		})
	}
}
