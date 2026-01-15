// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestStartupWithFailingUpstream(t *testing.T) {
	t.Parallel()

	// 1. Start a working upstream service
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `{"message": "hello"}`)
	}))
	defer ts.Close()

	failingPort := integration.FindFreePort(t)

	configContent := fmt.Sprintf(`
upstream_services:
  - name: "working-service"
    http_service:
      address: "%s"
      calls:
        hello:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/"
      tools:
        - call_id: "hello"
          name: "hello"
          description: "working tool"
  - name: "failing-service"
    http_service:
      address: "http://localhost:%d"
      calls:
        echo:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/"
      tools:
        - call_id: "echo"
          name: "echo"
          description: "failing tool"
`, ts.URL, failingPort)

    tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
    require.NoError(t, err)
    _, err = tmpFile.WriteString(configContent)
    require.NoError(t, err)
    err = tmpFile.Close()
    require.NoError(t, err)

	var jsonrpcPort int
	var ctx context.Context
	var cancel context.CancelFunc
	var appRunner *app.Application

	// Retry loop to handle "address already in use" race conditions from FindFreePort
	for attempt := 0; attempt < 3; attempt++ {
		jsonrpcPort = integration.FindFreePort(t)
		grpcRegPort := integration.FindFreePort(t)
		for grpcRegPort == jsonrpcPort {
			grpcRegPort = integration.FindFreePort(t)
		}

		ctx, cancel = context.WithCancel(context.Background())

		appRunner = app.NewApplication()

		errChan := make(chan error, 1)
		go func() {
			jsonrpcAddress := fmt.Sprintf(":%d", jsonrpcPort)
			grpcRegAddress := fmt.Sprintf(":%d", grpcRegPort)
			// Pass a test API key
			err := appRunner.Run(ctx, afero.NewOsFs(), false, jsonrpcAddress, grpcRegAddress, []string{tmpFile.Name()}, "test-api-key", 5*time.Second)
			if err != nil && err != context.Canceled {
				errChan <- err
			}
		}()

		// Wait briefly to check for immediate startup errors (like port conflicts)
		select {
		case err := <-errChan:
			cancel() // Clean up failed attempt
			if err != nil && (strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind")) {
				t.Logf("Port conflict detected on attempt %d, retrying...", attempt+1)
				continue
			}
			t.Fatalf("Server failed to start: %v", err)
		case <-time.After(500 * time.Millisecond):
			// Server started successfully (or didn't fail immediately)
			goto ServerStarted
		}
	}
	t.Fatal("Failed to find free ports after multiple attempts")

ServerStarted:
	defer cancel()

    httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
    integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

    // Include /mcp and api_key in the endpoint
    endpoint := fmt.Sprintf("http://127.0.0.1:%d/mcp?api_key=test-api-key", jsonrpcPort)

    client := mcp.NewClient(&mcp.Implementation{
        Name: "test-client",
        Version: "1.0.0",
    }, nil)

    transport := &mcp.StreamableClientTransport{
        Endpoint: endpoint,
    }

    ctxCall, cancelCall := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelCall()

    session, err := client.Connect(ctxCall, transport, nil)
    require.NoError(t, err)
    defer session.Close()

    listToolsResult, err := session.ListTools(ctxCall, &mcp.ListToolsParams{})
	require.NoError(t, err)

	workingToolFound := false
	failingToolFound := false

	for _, tool := range listToolsResult.Tools {
        if tool.Name == "working-service.hello" {
            workingToolFound = true
        }
        if tool.Name == "failing-service.echo" {
            failingToolFound = true
        }
	}

	require.True(t, workingToolFound, "Working tool should be registered")
    require.True(t, failingToolFound, "Failing tool should be registered even if service is down")

    t.Log("Verifying failing tool call fails")
    args := json.RawMessage(`{}`)
    result, err := session.CallTool(ctxCall, &mcp.CallToolParams{
            Name: "failing-service.echo",
            Arguments: args,
    })
    require.NoError(t, err)
    require.True(t, result.IsError, "Tool call should return isError=true for failing service")

    // 6. Fix the failing service in config (Reload Test)
    t.Log("Updating config to fix failing service...")
    newConfigContent := fmt.Sprintf(`
upstream_services:
  - name: "working-service"
    http_service:
      address: "%s"
      calls:
        hello:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/"
      tools:
        - call_id: "hello"
          name: "hello"
          description: "working tool"
  - name: "failing-service"
    http_service:
      address: "%s"
      calls:
        echo:
          method: "HTTP_METHOD_GET"
          endpoint_path: "/"
      tools:
        - call_id: "echo"
          name: "echo"
          description: "now working tool"
`, ts.URL, ts.URL) // Point failing-service to working URL

    // Overwrite config file
    err = os.WriteFile(tmpFile.Name(), []byte(newConfigContent), 0644)
    require.NoError(t, err)

    t.Log("Triggering config reload...")
    err = appRunner.ReloadConfig(context.Background(), afero.NewOsFs(), []string{tmpFile.Name()})
    require.NoError(t, err)

    // Verify failing-service is now working
    result, err = session.CallTool(ctxCall, &mcp.CallToolParams{
            Name: "failing-service.echo",
            Arguments: args,
    })
    require.NoError(t, err)
    t.Logf("Result from recovered tool: %v", result)
}
