package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/mcpany/core/server/pkg/testutil"
)

// Since E2E DB seeding is required by the product instructions,
// we ensure a test hits the API, creates a real execution log,
// and we can verify it contains the complex result.
func TestAuditLogE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverURL, cleanup := testutil.StartTestServer(ctx, t, "../../config.minimal.yaml")
	defer cleanup()

	// Wait for server to be ready
	err := testutil.WaitForReady(serverURL)
	if err != nil {
		t.Fatalf("Server not ready: %v", err)
	}

	// Wait for services to register
	time.Sleep(2 * time.Second)

	// Execute a tool to generate an audit log
	execURL := fmt.Sprintf("%s/api/v1/execute", serverURL)
	reqBody := `{"name": "echo_tool", "arguments": {"complex_result": {"nested": true, "list": [1,2,3]}}}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, execURL, strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-token") // From testutil.StartTestServer default

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute tool: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify the log was created with the complex result
	logsURL := fmt.Sprintf("%s/api/v1/audit/logs?limit=1", serverURL)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, logsURL, nil)
	if err != nil {
		t.Fatalf("Failed to create logs request: %v", err)
	}
	req.Header.Set("X-API-Key", "test-token")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to fetch logs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK for logs, got %d", resp.StatusCode)
	}

	var logsResponse struct {
		Entries []struct {
			Result string `json:"result"`
		} `json:"entries"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&logsResponse); err != nil {
		t.Fatalf("Failed to decode logs response: %v", err)
	}

	if len(logsResponse.Entries) == 0 {
		t.Fatalf("Expected at least one log entry")
	}

	resultStr := logsResponse.Entries[0].Result
	if !strings.Contains(resultStr, "complex_result") || !strings.Contains(resultStr, "nested") {
		t.Errorf("Expected log result to contain complex data, got: %s", resultStr)
	}
}
