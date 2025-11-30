package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the observability server on a random port.
	server, err := Start(ctx, ":0")
	require.NoError(t, err)
	require.NotNil(t, server)
	defer server.Shutdown(ctx)

	// Create a new test server.
	testServer := httptest.NewServer(server.Handler)
	defer testServer.Close()

	// Make a request to the /metrics endpoint.
	resp, err := testServer.Client().Get(testServer.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the response body.
	body := make([]byte, 4096)
	_, err = resp.Body.Read(body)
	require.NoError(t, err)

	// Verify that the response contains the expected metric.
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Make a dummy request to trigger the middleware.
	mcpServer, err := mcpserver.NewServer(ctx, tool.NewToolManager(), prompt.NewPromptManager(), resource.NewResourceManager(), auth.NewAuthManager(), serviceregistry.NewServiceRegistry(), &bus.BusProvider{})
	require.NoError(t, err)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	mcpServer.Server().ServeHTTP(rr, req)

	// Make another request to the /metrics endpoint to see the updated metrics.
	resp, err = testServer.Client().Get(testServer.URL + "/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read the response body.
	body = make([]byte, 4096)
	_, err = resp.Body.Read(body)
	require.NoError(t, err)

	assert.Contains(t, string(body), "mcp_server_requests")
}
