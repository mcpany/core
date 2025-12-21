// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiUser_EdgeCases(t *testing.T) {
	fs := afero.NewMemMapFs()
	configContent := `
global_settings:
  api_key: "global-secret"
upstream_services:
  - name: "dev-service"
    profiles:
      - name: "dev"
        id: "dev-profile"
    http_service:
      address: "http://localhost:8081"
      tools:
        - name: "dev-tool"
          call_id: "dev-call"
      calls:
        dev-call:
          id: "dev-call"
          endpoint_path: "/dev"
          method: "HTTP_METHOD_POST"
users:
  - id: "user-dev"
    profile_ids: ["dev-profile"]
    # No user-level auth, so it should fall back to global auth
  - id: "user-no-access"
    profile_ids: []
    # User exists but has no profile access
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

    // Reset viper
    viper.Reset()
    viper.Set("profiles", []string{"dev"})

	app := NewApplication()

	// Start server on random ports
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("localhost:%d", httpPort), "localhost:0", []string{"/config.yaml"}, 5*time.Second)
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	// Wait for server to start
	require.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", baseURL+"/healthz", nil)
		req.Header.Set("X-API-Key", "global-secret")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	t.Run("Invalid Path", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/u/too/short")
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("User Not Found", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/u/unknown-user/profile/dev-profile")
		require.NoError(t, err)
		defer resp.Body.Close()
		// Logic: parts[3] is uid. If not found -> 404.
        // Wait, "User not found" maps to http.StatusNotFound (404) or http.Error?
        // server.go: http.Error(w, "User not found", http.StatusNotFound)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

    t.Run("Global Auth Success", func(t *testing.T) {
        // User "user-dev" has no auth configured, so it falls back to global key "global-secret"
        req, err := http.NewRequest("POST", baseURL + "/mcp/u/user-dev/profile/dev-profile", nil)
        require.NoError(t, err)
        req.Header.Set("X-API-Key", "global-secret")

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        // It matches "dev-profile". User has access.
        // Since it's a POST to root of handler (stripped), it triggers stateless JSON-RPC check
        // "if r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "")"
        // But here path is /mcp/u/user-dev/profile/dev-profile which maps to / in delegate?
        // Let's check server.go:
        // prefix := fmt.Sprintf("/mcp/u/%s/profile/%s", uid, profileID)
        // http.StripPrefix(prefix, delegate).ServeHTTP(w, r.WithContext(ctx))
        // So yes, it maps to "/"

        // But body is empty/nil -> JSON Unmarshal error in stateless handler?
        // server.go: body, _ := io.ReadAll(r.Body); if err := json.Unmarshal(body, &req); err != nil { http.Error(w, "Invalid JSON", http.StatusBadRequest) }

        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
        // This confirms Auth passed and we reached the handler which rejected empty JSON.
    })

    t.Run("Global Auth Failure", func(t *testing.T) {
        req, err := http.NewRequest("POST", baseURL + "/mcp/u/user-dev/profile/dev-profile", nil)
        require.NoError(t, err)
        req.Header.Set("X-API-Key", "wrong-secret")

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
    })

    t.Run("Forbidden Access (User has no access to profile)", func(t *testing.T) {
        // user-no-access exists, but profile_ids is empty.
        // Try accessing dev-profile.
        req, err := http.NewRequest("POST", baseURL + "/mcp/u/user-no-access/profile/dev-profile", nil)
        require.NoError(t, err)
        req.Header.Set("X-API-Key", "global-secret") // Pass auth

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusForbidden, resp.StatusCode)
    })

    t.Run("Stateless JSON-RPC: Malformed JSON", func(t *testing.T) {
        req, err := http.NewRequest("POST", baseURL + "/mcp/u/user-dev/profile/dev-profile", bytes.NewBufferString("{bad-json}"))
        require.NoError(t, err)
        req.Header.Set("X-API-Key", "global-secret")

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
    })

    t.Run("Stateless JSON-RPC: Unsupported Method", func(t *testing.T) {
         reqBody := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "unknown/method",
			"params": {}
		}`
        req, err := http.NewRequest("POST", baseURL + "/mcp/u/user-dev/profile/dev-profile", bytes.NewBufferString(reqBody))
        require.NoError(t, err)
        req.Header.Set("X-API-Key", "global-secret")

        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()

        assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
    })
}
