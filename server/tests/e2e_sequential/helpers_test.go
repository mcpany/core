//go:build e2e

// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e_sequential

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func verifyEndpoint(t *testing.T, url string, expectedStatus int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		//nolint:gosec // G107: Url is constructed internally in test
		resp, err := http.Get(url)
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

// translatePath translates a container path to a host path for Docker-in-Docker
func translatePath(p string) string {
	hostRoot := os.Getenv("HOST_WORKSPACE_ROOT")
	if hostRoot == "" {
		return p
	}
	abs, _ := filepath.Abs(p)
	if strings.HasPrefix(abs, "/workspace") {
		return strings.Replace(abs, "/workspace", hostRoot, 1)
	}
	return p
}
