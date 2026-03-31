package interop_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/src/interop"
	"github.com/stretchr/testify/require"
)

// TestInteropAPIIntegration tests the API endpoint for the Interop tasks.
// According to Level 2 (Integration): Implement API-to-API integration tests for every feature at the CUJ level.
// Verify full request/response cycles. Data MUST be seeded via the database.
func TestInteropAPIIntegration(t *testing.T) {
	// Start an MCPANY Server using the integration test framework to satisfy
	// "Integration Forge: Write tests that hit real local endpoints."

	testCase := &framework.E2ETestCase{
		Name:                "Interop API Test",
		UpstreamServiceType: "interop",
		GenerateUpstreamConfig: func(endpoint string) string {
			return `api_keys: ["test-api-key"]`
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			task := interop.Task{
				ID:        "test-task-1",
				Framework: "CrewAI",
				Intent:    "task_delegation",
				Payload:   map[string]string{"role": "data_analyst"},
			}
			bodyBytes, err := json.Marshal(task)
			require.NoError(t, err)

			// The framework sets up mcpanyEndpoint typically like http://localhost:50050
			url := mcpanyEndpoint + "/api/v1/interop/task"

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
			require.NoError(t, err)

			req.Header.Set("Authorization", "Bearer test-api-key")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode, "Expected HTTP 200 OK")

			var res interop.TaskResult
			err = json.NewDecoder(resp.Body).Decode(&res)
			require.NoError(t, err)
			require.Equal(t, "success", res.Status)
			require.Equal(t, "data_analyst", res.Telemetry["delegated_role"])
		},
	}

	framework.RunE2ETest(t, testCase)
}
