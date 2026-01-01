package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestServiceRetry(t *testing.T) {
	// 1. Create a config file pointing to a non-existent service
	fs := afero.NewMemMapFs()

    // Pick a random port (hopefully free) or use port 0 if we could update config dynamically,
    // but here we need to fix the port in config first.
    // Let's assume port 54321 is free.
    targetPort := 54321
    targetURL := fmt.Sprintf("http://127.0.0.1:%d", targetPort)

	// Use mcp_service to force connection attempt at startup
	configContent := fmt.Sprintf(`
upstream_services:
  - name: "delayed-service"
    mcp_service:
      http_connection:
        http_address: "%s"
`, targetURL)

	err := afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// 2. Start the Application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := app.NewApplication()
	go func() {
		err := a.Run(ctx, fs, false, ":8081", "", []string{"config.yaml"}, 1*time.Second)
        if err != nil && ctx.Err() == nil {
            t.Logf("Application run error: %v", err)
        }
	}()

	// Wait for app to start
	time.Sleep(2 * time.Second)

	// Verify service is NOT registered (or is in error state)
    require.Eventually(t, func() bool {
        return a.ServiceRegistry != nil
    }, 5*time.Second, 100*time.Millisecond, "ServiceRegistry not initialized")

    errStr, hasError := a.ServiceRegistry.GetServiceError("delayed-service")
    if !hasError {
        // Retry with sanitized name if needed
         errStr, hasError = a.ServiceRegistry.GetServiceError("delayed_service")
    }

    t.Logf("Initial Service Error: %s (hasError: %v)", errStr, hasError)

    // With mcp_service, registration MUST fail if upstream is down.
    if !hasError {
        t.Log("Expected registration error but got none!")
        t.Fail()
    }

    // 3. Start the mock service
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // MCP Server needs to respond to JSON-RPC or tool discovery
        w.Header().Set("Content-Type", "application/json")
        // Just return a valid JSON-RPC response to whatever comes in
        // Ideally we parse the request ID, but for connection check, maybe it's enough to respond JSON
        w.Write([]byte(`{"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": "2024-11-05", "capabilities": {}, "serverInfo": {"name": "mock", "version": "1.0"}}}`))
	}))

    // Listen on all interfaces to handle IPv4/IPv6 ambiguity
    l, err := net.Listen("tcp", fmt.Sprintf(":%d", targetPort))
    if err != nil {
        t.Skipf("Port %d busy: %v", targetPort, err)
        return
    }
    ts.Listener.Close()
    ts.Listener = l
	ts.Start()
	defer ts.Close()

    t.Logf("Started mock service at %s", targetURL)

	// 4. Wait and see if it recovers
	// Wait for longer than the likely backoff (start with 5s)

    t.Log("Waiting for retry...")
    time.Sleep(10 * time.Second)

    // Check if service is now healthy
    // Check ToolManager for ServiceInfo as that's the source of truth for "Working" services
    _, infoOk := a.ToolManager.GetServiceInfo("delayed-service")
    if !infoOk {
        _, infoOk = a.ToolManager.GetServiceInfo("delayed_service")
    }

    if !infoOk {
        t.Log("Service still missing info (failed to recover)")
        errStr, _ := a.ServiceRegistry.GetServiceError("delayed-service")
        t.Logf("Current Service Error: %s", errStr)
        t.Fail() // IT FAILED TO RECOVER
    } else {
        t.Log("Service recovered successfully!")
    }
}
