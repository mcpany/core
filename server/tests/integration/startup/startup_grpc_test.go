// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"context"
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

func TestStartupWithFailingGrpcUpstream(t *testing.T) {
	t.Parallel()

	// 1. Start a working upstream service (HTTP)
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
  - name: "failing-grpc-service"
    grpc_service:
      address: "localhost:%d"
      use_reflection: true
`, ts.URL, failingPort)

    tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
    require.NoError(t, err)
    _, err = tmpFile.WriteString(configContent)
    require.NoError(t, err)
    err = tmpFile.Close()
    require.NoError(t, err)

    jsonrpcPort := integration.FindFreePort(t)
    grpcRegPort := integration.FindFreePort(t)
    for grpcRegPort == jsonrpcPort {
        grpcRegPort = integration.FindFreePort(t)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    appRunner := app.NewApplication()

    // We expect the server to start even if grpc service fails
    serverErr := make(chan error, 1)
    go func() {
        err := appRunner.Run(ctx, afero.NewOsFs(), false, fmt.Sprintf(":%d", jsonrpcPort), fmt.Sprintf(":%d", grpcRegPort), []string{tmpFile.Name()}, 5*time.Second)
        if err != nil && err != context.Canceled {
             serverErr <- err
        }
        close(serverErr)
    }()

    httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
    // Wait for health check. If it takes too long, it means it crashed or is stuck.
    integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

    select {
    case err := <-serverErr:
        t.Fatalf("Server failed to start: %v", err)
    default:
    }

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

	for _, tool := range listToolsResult.Tools {
        if tool.Name == "working-service.hello" {
            workingToolFound = true
        }
	}

	require.True(t, workingToolFound, "Working tool should be registered")
}
