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

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestStartupWithFailingUpstream(t *testing.T) {
	// Enable file config for this test
	t.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")

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
      address: "http://127.0.0.1:%d"
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

	// Use dynamic ports to avoid race conditions
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)

	appRunner = app.NewApplication()

	errChan := make(chan error, 1)
	go func() {
		// Use 127.0.0.1:0 for dynamic port allocation to match e2e_helpers and avoid :0 issues
		jsonrpcAddress := "127.0.0.1:0"
		grpcRegAddress := "127.0.0.1:0"
		// Pass a test API key
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              afero.NewOsFs(),
			Stdio:           false,
			JSONRPCPort:     jsonrpcAddress,
			GRPCPort:        grpcRegAddress,
			ConfigPaths:     []string{tmpFile.Name()},
			APIKey:          "test-api-key",
			ShutdownTimeout: 5 * time.Second,
		}
		err := appRunner.Run(opts)
		if err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	// Wait for startup to get the bound ports
	select {
	case err := <-errChan:
		cancel()
		t.Fatalf("Server failed to start: %v", err)
	case <-time.After(500 * time.Millisecond):
		// Give it a bit of time to initialize listeners if WaitForStartup isn't available
		// But appRunner should have Bound ports populated once listeners are started.
		// We can loop wait for them.
	}

	// Wait for ports to be assigned
	// Wait for startup to ensure ports are bound
	err = appRunner.WaitForStartup(ctx)
	require.NoError(t, err, "failed to wait for startup")

	require.NotZero(t, appRunner.BoundHTTPPort.Load(), "BoundHTTPPort should be set")
	// require.NotZero(t, appRunner.BoundGRPCPort.Load(), "BoundGRPCPort should be set") // Start with just HTTP if that's what we use

	jsonrpcPort = int(appRunner.BoundHTTPPort.Load())
	// grpcRegPort := appRunner.BoundGRPCPort

	defer cancel()

	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	// Include /mcp and api_key in the endpoint
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/mcp?api_key=test-api-key", jsonrpcPort)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
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

	require.Eventually(t, func() bool {
		listToolsResult, err := session.ListTools(ctxCall, &mcp.ListToolsParams{})
		if err != nil {
			return false
		}

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
		return workingToolFound && failingToolFound
	}, 10*time.Second, 100*time.Millisecond, "All tools should be registered")

	t.Log("Verifying failing tool call fails")
	args := json.RawMessage(`{}`)
	result, err := session.CallTool(ctxCall, &mcp.CallToolParams{
		Name:      "failing-service.echo",
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
		Name:      "failing-service.echo",
		Arguments: args,
	})
	require.NoError(t, err)
	t.Logf("Result from recovered tool: %v", result)
}
