//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//nolint:revive // This is a test package name
package e2e_sequential

import (
	"context"
	"net/http"
	"testing"
	"time"
)

//nolint:unused // It's used in tests
func verifyEndpoint(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == expectedStatus {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("Failed to verify endpoint %s within %v", url, timeout)
}
