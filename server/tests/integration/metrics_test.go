//go:build integration
// +build integration

package integration

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpany-integration-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configContent := `
upstreamServices:
  - name: "http-echo-server"
    httpService:
      address: "http://127.0.0.1:8080"
      calls:
        - operationId: "echo"
          description: "Echoes back the request body"
          method: "HTTP_METHOD_POST"
          endpointPath: "/echo"
`
	configPath := filepath.Join(tempDir, "config.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mcpany := testutil.NewMCPAny(t, ctx,
		"--config-path", configPath,
		"--metrics-listen-address", "127.0.0.1:9090",
	)
	defer mcpany.Stop()

	// Wait for the server to be ready
	time.Sleep(2 * time.Second)

	// Make a request to the echo tool
	_, err = testutil.CallTool(t, "http-echo-server/-/echo", `{"message": "hello"}`)
	require.NoError(t, err)

	// Make a request to the metrics endpoint
	resp, err := http.Get("http://127.0.0.1:9090/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check if the metrics are present in the response
	assert.Contains(t, string(body), "mcpany_tool_http_echo_server_echo_call_total 1")
	assert.Contains(t, string(body), "mcpany_tools_call_total 1")
}
