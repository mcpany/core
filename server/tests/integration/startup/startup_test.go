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
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/mcpany/core/tests/integration"
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

	var jsonrpcPort, grpcRegPort int
	var appRunner *app.Application
	var ctx context.Context
	var cancel context.CancelFunc

	started := false
	for i := 0; i < 3; i++ {
		jsonrpcPort = integration.FindFreePort(t)
		grpcRegPort = integration.FindFreePort(t)
		for grpcRegPort == jsonrpcPort {
			grpcRegPort = integration.FindFreePort(t)
		}

		ctx, cancel = context.WithCancel(context.Background())
		appRunner = app.NewApplication()

		serverErrChan := make(chan error, 1)

		go func() {
			err := appRunner.Run(ctx, afero.NewOsFs(), false, fmt.Sprintf(":%d", jsonrpcPort), fmt.Sprintf(":%d", grpcRegPort), []string{tmpFile.Name()}, 5*time.Second)
			if err != nil && err != context.Canceled {
				t.Logf("Attempt %d: Server exited with error: %v", i+1, err)
				serverErrChan <- err
			}
			close(serverErrChan)
		}()

		// Wait briefly to check for immediate start failure (e.g. bind error)
		select {
		case err := <-serverErrChan:
			if err != nil {
				cancel()
				continue // Retry
			}
		case <-time.After(500 * time.Millisecond):
			// Proceed to check health
		}

		httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)

		// Soft check for health
		healthClient := http.Client{Timeout: 1 * time.Second}
		healthy := false
		for j := 0; j < 10; j++ {
			resp, err := healthClient.Get(httpUrl)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				healthy = true
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			if j < 9 {
				time.Sleep(500 * time.Millisecond) // Wait reduced for retry loop
			}
		}

		if healthy {
			started = true
			break
		}

		// Failed, cleanup and retry
		cancel()
	}
	require.True(t, started, "Failed to start server after retries")
	defer cancel() // Defer cancel for the successful context

    endpoint := fmt.Sprintf("http://127.0.0.1:%d", jsonrpcPort)

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
    _, err = session.CallTool(ctxCall, &mcp.CallToolParams{
            Name: "failing-service.echo",
            Arguments: args,
    })
    require.Error(t, err)

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
    err = appRunner.ReloadConfig(afero.NewOsFs(), []string{tmpFile.Name()})
    require.NoError(t, err)

    // Verify failing-service is now working
    result, err := session.CallTool(ctxCall, &mcp.CallToolParams{
            Name: "failing-service.echo",
            Arguments: args,
    })
    require.NoError(t, err)
    t.Logf("Result from recovered tool: %v", result)
}
