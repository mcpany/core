package interop_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/mcpany/core/src/interop"
	"github.com/stretchr/testify/require"
)

// TestInteropIntegration verifies the interop API endpoints using a full local server setup
// and database seeding for the user state, matching the integration test requirements.
func TestInteropIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a full test server
	info := integration.StartInProcessMCPANYServer(t, "interop-test")
	if info.Process != nil {
		defer info.Process.Stop()
	}

	// We must ensure the server is ready, StartInProcessMCPANYServer waits for health check

	// Seed user data using database seeding helper
	seedData := []byte(`{"users": [{"id": "test-user-1", "authentication": {"bearer_token": {"token": {"plain_text": "test-token-123"}}}}]}`)
	err := info.SeedDatabase(ctx, seedData)
	require.NoError(t, err, "Failed to seed database for interop test")

	// Call GET /api/v1/interop/adapters
	req, err := http.NewRequestWithContext(ctx, "GET", info.JSONRPCEndpoint+"/api/v1/interop/adapters", nil)
	require.NoError(t, err)
	// We use the seeded user's token instead of the master API Key
	req.Header.Set("Authorization", "Bearer test-token-123")

	resp, err := info.HTTPClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var adaptersResp map[string][]string
	err = json.NewDecoder(resp.Body).Decode(&adaptersResp)
	require.NoError(t, err)
	require.Contains(t, adaptersResp["adapters"], "OpenClaw")

	// Call POST /api/v1/interop/task
	task := interop.Task{
		ID:        "int-test-1",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload:   map[string]string{"foo": "bar"},
	}
	bodyBytes, _ := json.Marshal(task)

	reqTask, err := http.NewRequestWithContext(ctx, "POST", info.JSONRPCEndpoint+"/api/v1/interop/task", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)
	reqTask.Header.Set("Authorization", "Bearer test-token-123")
	reqTask.Header.Set("Content-Type", "application/json")

	respTask, err := info.HTTPClient.Do(reqTask)
	require.NoError(t, err)
	defer respTask.Body.Close()

	require.Equal(t, http.StatusOK, respTask.StatusCode)

	var taskRes interop.TaskResult
	err = json.NewDecoder(respTask.Body).Decode(&taskRes)
	require.NoError(t, err)
	require.Equal(t, "success", taskRes.Status)
}
