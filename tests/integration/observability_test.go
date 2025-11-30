package integration

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservability(t *testing.T) {
	// Start the server in a separate goroutine.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	runner := app.NewApplication()
	go func() {
		err := runner.Run(ctx, afero.NewMemMapFs(), false, "localhost:50050", "", nil, 5*time.Second)
		assert.NoError(t, err)
	}()

	// Wait for the server to start.
	var resp *http.Response
	var err error
	for i := 0; i < 10; i++ {
		_, err = http.Post("http://localhost:50050", "application/json", strings.NewReader(`{"jsonrpc":"2.0","method":"tools/list","id":1}`))
		if err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, err)

	// Make a request to the /metrics endpoint.
	resp, err = http.Get("http://localhost:9090/metrics")
	require.NoError(t, err)
	defer resp.Body.Close()


	// Verify the response.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body := make([]byte, 4096)
	_, err = resp.Body.Read(body)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(body), "mcp_server_requests"))
}
