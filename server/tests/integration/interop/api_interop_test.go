package interop_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
)

func TestInteropAPIIntegration(t *testing.T) {
	info := integration.StartMCPANYServer(t, "interop-test")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	payload := map[string]interface{}{
		"framework": "CrewAI",
		"intent":    "task_delegation",
		"payload":   map[string]string{"role": "data_analyst"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	url := fmt.Sprintf("%s/api/v1/interop/task", info.HTTPEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+info.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	seedData := []byte(`{}`)
	err = info.SeedDatabase(ctx, seedData)
	if err != nil {
		t.Logf("Seed failed: %v", err)
	}

	var resp *http.Response
	var reqErr error
	for i := 0; i < 5; i++ {
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		resp, reqErr = client.Do(req)
		if reqErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	if reqErr != nil {
		t.Fatalf("Failed to send request: %v", reqErr)
	}

	if resp == nil || resp.StatusCode != http.StatusOK {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		t.Fatalf("Expected status 200 OK, got %d", status)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if res["status"] != "success" {
		t.Errorf("Expected status 'success', got '%v'", res["status"])
	}
}
