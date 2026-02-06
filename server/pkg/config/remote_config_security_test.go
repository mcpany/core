// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TestRemoteConfigSecurity verifies that remote configurations (loaded via URL)
// cannot define services that execute local commands, preventing RCE.
func TestRemoteConfigSecurity(t *testing.T) {
	// 0. Bypass SSRF protection for local test server
	// This is necessary because httptest.NewServer binds to localhost,
	// which is blocked by the default SafeDialer.
	originalClient := httpClient
	httpClient = http.DefaultClient
	defer func() { httpClient = originalClient }()

	// 1. Setup malicious config server
	// This config attempts to define services that run local commands.
	maliciousConfig := `
upstream_services:
  - name: "cmd_service"
    command_line_service:
      command: "echo pwned_cmd"
  - name: "mcp_stdio_service"
    mcp_service:
      stdio_connection:
        command: "echo"
        args: ["pwned_stdio"]
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		fmt.Fprint(w, maliciousConfig)
	}))
	defer server.Close()

	// 2. Load config from URL
	fs := afero.NewMemMapFs()
	store := NewFileStore(fs, []string{server.URL + "/config.yaml"})

	_, err := store.Load(context.Background())

	// 3. Assertion: Expect FAILURE loading malicious config
	if err == nil {
		t.Fatal("Expected error loading malicious remote config, but got success")
	} else {
		// Verify it is the correct error
		require.ErrorContains(t, err, "forbidden in remote configurations")
		t.Logf("Successfully blocked malicious remote config: %v", err)
	}
}
