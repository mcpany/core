package framework

import (
	"bytes"
	"context"

	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

// TestE2ECaching tests the end-to-end caching functionality.
//
// t is the t.
func TestE2ECaching(t *testing.T) {
	t.Parallel()
	RunE2ETest(t, &E2ETestCase{
		Name:                "caching",
		UpstreamServiceType: "http",
		BuildUpstream:       BuildCachingServer,
		RegisterUpstream:    RegisterCachingService,
		ValidateMiddlewares: func(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
			ValidateCaching(t, mcpanyEndpoint, upstreamEndpoint)
		},
		InvokeAIClient:      func(_ *testing.T, _ string) {},
		RegistrationMethods: []RegistrationMethod{GRPCRegistration},
	})
}

// BuildCachingServer builds and starts a caching server for testing.
//
// t is the t.
//
// Returns the result.
func BuildCachingServer(t *testing.T) *integration.ManagedProcess {
	port := integration.FindFreePort(t)
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	proc := integration.NewManagedProcess(t, "http_caching_server", filepath.Join(root, "../build/test/bin/http_caching_server"), []string{"--port", fmt.Sprintf("%d", port)}, nil)
	proc.Port = port
	return proc
}

// RegisterCachingService registers the caching service with the MCP server.
//
// t is the t.
// registrationClient is the registrationClient.
// upstreamEndpoint is the upstreamEndpoint.
func RegisterCachingService(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
	const serviceID = "e2e_caching_server"
	integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "get_data", "/", "GET", nil)
}

// NoOpMiddleware is a middleware that does nothing and calls the next handler.
//
// _ is an unused parameter.
// next is the next.
//
// Returns the result.
func NoOpMiddleware(_ *testing.T, next http.Handler) http.Handler {
	return next
}

func callTool(t *testing.T, mcpanyEndpoint, toolName string) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": toolName,
		},
		"id": "1",
	})
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), "POST", mcpanyEndpoint, bytes.NewBuffer(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// ValidateCaching validates that caching is working correctly.
//
// t is the t.
// mcpanyEndpoint is the mcpanyEndpoint.
// upstreamEndpoint is the upstreamEndpoint.
func ValidateCaching(t *testing.T, mcpanyEndpoint, upstreamEndpoint string) {
	// 1. Reset the upstream server's counter.
	req, err := http.NewRequestWithContext(context.Background(), "POST", fmt.Sprintf("http://%s/reset", upstreamEndpoint), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 2. Make a request to the tool and check that the upstream service was called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics := getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(1), metrics["counter"])

	// 3. Make another request to the tool and check that the upstream service was NOT called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(1), metrics["counter"])

	// 4. Advance the fake clock to expire the cache.
	time.Sleep(6 * time.Second)

	// 5. Make another request to the tool and check that the upstream service was called.
	callTool(t, mcpanyEndpoint, "e2e_caching_server.get_data")

	metrics = getUpstreamMetrics(t, upstreamEndpoint)
	require.Equal(t, int64(2), metrics["counter"])
}

func getUpstreamMetrics(t *testing.T, upstreamEndpoint string) map[string]int64 {
	req, err := http.NewRequestWithContext(context.Background(), "GET", fmt.Sprintf("http://%s/metrics", upstreamEndpoint), nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var metrics map[string]int64
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	require.NoError(t, err)

	return metrics
}
